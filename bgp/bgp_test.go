//go:build linux

package bgp

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/fabiolb/fabio/config"
	api "github.com/osrg/gobgp/v4/api"
	"github.com/osrg/gobgp/v4/pkg/apiutil"
	"github.com/osrg/gobgp/v4/pkg/packet/bgp"
	"github.com/osrg/gobgp/v4/pkg/server"
)

// bgpTestParams describes one peering scenario.  The gobgpd side is
// configured by cfgFile, the fabio side by the remaining fields, and the
// two must agree on addresses, ports and ASNs.
type bgpTestParams struct {
	cfgFile     string
	localAddr   string
	peerAddr    string
	nextHop     string
	grpcAddr    string
	peerAPIAddr string
	listenPort  int
	peerASN     uint
}

const localASN = 65000

func TestBGPHandler(t *testing.T) {
	// An eBGP peer.  gobgp prepends our ASN on export, so the peer must see
	// an AS_PATH of exactly [65000] - not [65000 65000].
	t.Run("ebgp", func(t *testing.T) {
		testBGPHandler(t, bgpTestParams{
			cfgFile:     "bgp.toml",
			localAddr:   "127.0.0.2",
			peerAddr:    "127.0.0.3",
			nextHop:     "1.2.3.4",
			grpcAddr:    "127.0.0.2:50051",
			peerAPIAddr: "127.0.0.3:50051",
			listenPort:  1790,
			peerASN:     65001,
		}, []int{localASN})
	})

	// An iBGP peer, i.e. one sharing our ASN.  gobgp drops any path whose
	// AS_PATH contains the peer's ASN, so the route only reaches the peer
	// if we leave AS_PATH alone and let gobgp attach an empty one.
	t.Run("ibgp", func(t *testing.T) {
		testBGPHandler(t, bgpTestParams{
			cfgFile:     "bgp_ibgp.toml",
			localAddr:   "127.0.0.4",
			peerAddr:    "127.0.0.5",
			grpcAddr:    "127.0.0.4:50051",
			peerAPIAddr: "127.0.0.5:50051",
			listenPort:  1791,
			peerASN:     localASN,
		}, nil)
	})
}

// testBGPHandler brings up a gobgpd peer, peers fabio with it and asserts
// that routes added through the BGPHandler show up in - and disappear from -
// the peer's global RIB with the expected AS_PATH.
func testBGPHandler(t *testing.T, p bgpTestParams, wantASPath []int) {
	serverCmd := &gobgpserver{
		cmdPath: "gobgpd",
		cfgFile: p.cfgFile,
		apiHost: p.peerAPIAddr,
	}
	err := serverCmd.start()
	if err != nil {
		t.Logf("error calling gobgpd command, probably not installed. skipping: %s", err)
		t.SkipNow()
	}
	defer serverCmd.stop()
	cfg := &config.BGP{
		BGPEnabled:       true,
		Asn:              localASN,
		AnycastAddresses: []string{"1.2.3.4/32"},
		RouterID:         p.localAddr,
		ListenPort:       p.listenPort,
		ListenAddresses:  []string{p.localAddr},
		Peers: []config.BGPPeer{
			{
				NeighborAddress: p.peerAddr,
				NeighborPort:    uint(p.listenPort),
				Asn:             p.peerASN,
				MultiHop:        false,
			},
		},
		EnableGRPC:        true,
		GRPCListenAddress: p.grpcAddr,
		NextHop:           p.nextHop,
	}
	bh, err := NewBGPHandler(cfg)
	if err != nil {
		t.Fatal(err)
	}
	go bh.server.Serve()
	defer bh.server.Stop()
	err = bh.startBGP(context.Background())
	if err != nil {
		t.Fatalf("error starting BGP: %s", err)
	}

	// setPolicies installs the deny-all-neighbors import policy.  It is the
	// only thing keeping peer-advertised routes out of the global RIB, so
	// exercise it here rather than trusting that it still applies cleanly
	// against whatever gobgp version we are built with.
	err = bh.setPolicies()
	if err != nil {
		t.Fatalf("error setting policies: %s", err)
	}

	err = bh.addNeighbors(context.Background(), cfg.Peers)
	if err != nil {
		t.Fatalf("error adding neighbors: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	callbacks := server.WatchEventMessageCallbacks{
		OnPeerUpdate: func(p *apiutil.WatchEventMessage_PeerEvent, ts time.Time) {
			t.Logf("EVENT RECEIVED %#v", p.Peer)
			if p.Peer.State.SessionState == bgp.BGP_FSM_ESTABLISHED {
				cancel()
			}
		},
	}

	if err := bh.server.WatchEvent(ctx, callbacks, server.WatchPeer()); err != nil {
		t.Fatal(err)
	}

	<-ctx.Done()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatal("context deadline exceeded")
	}

	gc := gobpgclient{
		cmdPath:  "gobgp",
		hostAddr: p.peerAddr,
	}

	// now start a test table

	for _, tst := range []struct {
		name      string
		cmd       func() error
		routeKeys []string
	}{
		{
			name: "test add route",
			cmd: func() error {
				return bh.AddRoutes(context.Background(), cfg.AnycastAddresses)
			},
			routeKeys: []string{"1.2.3.4/32"},
		},
		{
			name: "test delete route",
			cmd: func() error {
				return bh.DeleteRoutes(context.Background(), []string{"1.2.3.4/32"})
			},
			routeKeys: nil,
		},
	} {
		t.Run(tst.name, func(t *testing.T) {
			err := tst.cmd()
			if err != nil {
				t.Fatal(err)
			}
			routes, err := gc.globalRib(t)
			if err != nil {
				t.Fatal(err)
			}
			if len(routes) != len(tst.routeKeys) {
				t.Fatalf("routes don't match, have %d want %d",
					len(routes), len(tst.routeKeys))
			}
			for _, r := range tst.routeKeys {
				entries, ok := routes[r]
				if !ok {
					t.Fatalf("route %s not found", r)
				}
				for _, e := range entries {
					if got := e.asPath(); !slices.Equal(got, wantASPath) {
						t.Errorf("route %s as path = %v, want %v", r, got, wantASPath)
					}
				}
			}
		})
	}

}

type ribEntry struct {
	Nlri struct {
		Prefix string `json:"prefix"`
	} `json:"nlri"`
	Age   int  `json:"age"`
	Best  bool `json:"best"`
	Attrs []struct {
		Type    int `json:"type"`
		Value   int `json:"value,omitempty"`
		AsPaths []struct {
			SegmentType int   `json:"segment_type"`
			Num         int   `json:"num"`
			Asns        []int `json:"asns"`
		} `json:"as_paths,omitempty"`
		Nexthop string `json:"nexthop,omitempty"`
	} `json:"attrs"`
	Stale bool `json:"stale"`
}

// asPath flattens the ASNs of the entry's AS_PATH attribute, in order.
func (e ribEntry) asPath() []int {
	var asns []int
	for _, a := range e.Attrs {
		for _, seg := range a.AsPaths {
			asns = append(asns, seg.Asns...)
		}
	}
	return asns
}

type gobpgclient struct {
	cmdPath  string
	hostAddr string
}

func (gc *gobpgclient) globalRib(t *testing.T) (map[string][]ribEntry, error) {
	out, err := exec.Command(gc.cmdPath, "-u", gc.hostAddr, "-j", "global", "rib").Output()
	if err != nil {
		return nil, err
	}
	var rv map[string][]ribEntry
	err = json.Unmarshal(out, &rv)
	if err != nil {
		t.Logf("raw: %s\n", out)
		return nil, err
	}
	return rv, nil
}

type gobgpserver struct {
	cmdPath string
	cfgFile string
	apiHost string
	cmd     *exec.Cmd
}

func (gs *gobgpserver) start() error {
	gs.cmd = exec.Command(gs.cmdPath,
		"-p",
		"-f", filepath.Join("test_data", gs.cfgFile),
		"--api-hosts", gs.apiHost,
		"-l", "info")
	gs.cmd.Stdout = os.Stdout
	gs.cmd.Stderr = os.Stderr
	return gs.cmd.Start()
}

func (gs *gobgpserver) stop() error {
	if gs.cmd.Process != nil {
		return gs.cmd.Process.Kill()
	}
	return nil
}

// TestGOBGPDConfigFile covers the bgp.gobgpdcfgfile path.  The operator's file
// is the only source of peers and policy on that path - fabio does not apply
// its own deny-all import policy there - so applying just the global section
// would leave fabio peered with nobody and filtering nothing.
func TestGOBGPDConfigFile(t *testing.T) {
	cfg := &config.BGP{
		BGPEnabled:    true,
		RouterID:      "127.0.0.6",
		GOBGPDCfgFile: filepath.Join("test_data", "bgp_cfgfile.toml"),
	}
	// An operator who configures everything in the gobgpd file and advertises
	// nothing themselves has no reason to set bgp.routerid, and ValidateConfig
	// does not require one on this path, so construction must not insist on it.
	if _, err := NewBGPHandler(&config.BGP{
		BGPEnabled:    true,
		GOBGPDCfgFile: cfg.GOBGPDCfgFile,
	}); err != nil {
		t.Errorf("handler with a config file and no routerID: %s", err)
	}

	bh, err := NewBGPHandler(cfg)
	if err != nil {
		t.Fatal(err)
	}
	go bh.server.Serve()
	defer bh.server.Stop()

	if err := bh.applyGOBGPDConfig(context.Background()); err != nil {
		t.Fatalf("error applying gobgpd config file: %s", err)
	}

	var peers []string
	err = bh.server.ListPeer(context.Background(), &api.ListPeerRequest{}, func(p *api.Peer) {
		peers = append(peers, p.Conf.NeighborAddress)
	})
	if err != nil {
		t.Fatalf("error listing peers: %s", err)
	}
	if !slices.Contains(peers, "127.0.0.7") {
		t.Errorf("neighbour from config file not applied, peers = %v", peers)
	}

	var policies []string
	err = bh.server.ListPolicy(context.Background(), &api.ListPolicyRequest{}, func(p *api.Policy) {
		policies = append(policies, p.Name)
	})
	if err != nil {
		t.Fatalf("error listing policies: %s", err)
	}
	if !slices.Contains(policies, "cfgfile-reject") {
		t.Errorf("policy from config file not applied, policies = %v", policies)
	}
}

// TestAddRoutesFamilies checks that anycast prefixes of both families end up in
// the global RIB under the right AFI, and are removed again on delete.
func TestAddRoutesFamilies(t *testing.T) {
	cfg := &config.BGP{
		BGPEnabled:      true,
		Asn:             localASN,
		RouterID:        "127.0.0.8",
		ListenPort:      1794,
		ListenAddresses: []string{"127.0.0.8"},
	}
	bh, err := NewBGPHandler(cfg)
	if err != nil {
		t.Fatal(err)
	}
	go bh.server.Serve()
	defer bh.server.Stop()
	if err := bh.startBGP(context.Background()); err != nil {
		t.Fatalf("error starting BGP: %s", err)
	}

	routes := []string{"1.2.3.4/32", "2001:db8::1/128"}
	if err := bh.AddRoutes(context.Background(), routes); err != nil {
		t.Fatalf("error adding routes: %s", err)
	}

	for _, tc := range []struct {
		family bgp.Family
		want   string
	}{
		{bgp.RF_IPv4_UC, "1.2.3.4/32"},
		{bgp.RF_IPv6_UC, "2001:db8::1/128"},
	} {
		got := listPrefixes(t, bh, tc.family)
		if !slices.Contains(got, tc.want) {
			t.Errorf("%s: prefix %s missing from global rib, have %v", tc.family, tc.want, got)
		}
	}

	if err := bh.DeleteRoutes(context.Background(), routes); err != nil {
		t.Fatalf("error deleting routes: %s", err)
	}
	for _, family := range []bgp.Family{bgp.RF_IPv4_UC, bgp.RF_IPv6_UC} {
		if got := listPrefixes(t, bh, family); len(got) != 0 {
			t.Errorf("%s: prefixes remain after delete: %v", family, got)
		}
	}
}

func listPrefixes(t *testing.T, bh *BGPHandler, family bgp.Family) []string {
	t.Helper()
	var got []string
	err := bh.server.ListPath(apiutil.ListPathRequest{
		TableType: api.TableType_TABLE_TYPE_GLOBAL,
		Family:    family,
	}, func(prefix bgp.NLRI, paths []*apiutil.Path) {
		got = append(got, prefix.String())
	})
	if err != nil {
		t.Fatalf("error listing paths for %s: %s", family, err)
	}
	return got
}

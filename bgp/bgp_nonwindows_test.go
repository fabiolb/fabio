//go:build !windows

package bgp

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/fabiolb/fabio/config"
	api "github.com/osrg/gobgp/v4/api"
	"github.com/osrg/gobgp/v4/pkg/apiutil"
	"github.com/osrg/gobgp/v4/pkg/packet/bgp"
	"github.com/osrg/gobgp/v4/pkg/server"
)

// loopback is the only address every platform actually assigns to the loopback
// interface.  Linux treats the whole of 127.0.0.0/8 as local, so a test can
// hand each speaker its own 127.0.0.x; macOS assigns 127.0.0.1 alone and
// binding anything else fails with EADDRNOTAVAIL.  Everything here therefore
// shares this one address and is kept apart by port instead.
const loopback = "127.0.0.1"

const localASN = 65000

// bgpTestParams describes one peering scenario.  Addresses and ports are not
// part of it: both ends sit on loopback and take freshly reserved ports, so
// all that varies between scenarios is who the peer claims to be.
type bgpTestParams struct {
	peerASN      uint
	peerRouterID string
	nextHop      string
}

// gobgpdCfg is the data behind the templates in test_data.  Rendering the
// gobgpd config from the same values the fabio side is built from is what
// keeps the two ends of a peering in agreement.
type gobgpdCfg struct {
	Addr       string // address gobgpd binds, and the peer's view of us
	ASN        uint
	RouterID   string
	ListenPort int
	PeerAddr   string
	PeerASN    uint
	PeerPort   int
}

func TestBGPHandler(t *testing.T) {
	// An eBGP peer.  gobgp prepends our ASN on export, so the peer must see
	// an AS_PATH of exactly [65000] - not [65000 65000].
	t.Run("ebgp", func(t *testing.T) {
		testBGPHandler(t, bgpTestParams{
			peerASN:      65001,
			peerRouterID: "192.0.2.3",
			nextHop:      "1.2.3.4",
		}, []int{localASN})
	})

	// An iBGP peer, i.e. one sharing our ASN.  gobgp drops any path whose
	// AS_PATH contains the peer's ASN, so the route only reaches the peer
	// if we leave AS_PATH alone and let gobgp attach an empty one.
	t.Run("ibgp", func(t *testing.T) {
		testBGPHandler(t, bgpTestParams{
			peerASN:      localASN,
			peerRouterID: "192.0.2.5",
		}, nil)
	})
}

// testBGPHandler brings up a gobgpd peer, peers fabio with it and asserts
// that routes added through the BGPHandler show up in - and disappear from -
// the peer's global RIB with the expected AS_PATH.
func testBGPHandler(t *testing.T, p bgpTestParams, wantASPath []int) {
	listenPort, peerPort := freePort(t), freePort(t)
	grpcPort, peerAPIPort := freePort(t), freePort(t)

	cfgFile := renderCfg(t, "gobgpd_peer.toml.tmpl", gobgpdCfg{
		Addr:       loopback,
		ASN:        p.peerASN,
		RouterID:   p.peerRouterID,
		ListenPort: peerPort,
		PeerAddr:   loopback,
		PeerASN:    localASN,
		PeerPort:   listenPort,
	})

	serverCmd := &gobgpserver{
		cmdPath: "gobgpd",
		cfgFile: cfgFile,
		apiHost: net.JoinHostPort(loopback, strconv.Itoa(peerAPIPort)),
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
		// addNeighbors reuses the routerID as the transport local address, so
		// on this path it has to be an address we can actually bind.  That
		// makes it the peer's router ID that has to be something else: the
		// two identifiers still have to differ or the session collides.
		RouterID:        loopback,
		ListenPort:      listenPort,
		ListenAddresses: []string{loopback},
		Peers: []config.BGPPeer{
			{
				NeighborAddress: loopback,
				NeighborPort:    uint(peerPort),
				Asn:             p.peerASN,
				MultiHop:        false,
			},
		},
		EnableGRPC:        true,
		GRPCListenAddress: net.JoinHostPort(loopback, strconv.Itoa(grpcPort)),
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
		hostAddr: loopback,
		port:     peerAPIPort,
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

// freePort reserves an unused TCP port on the loopback address.  The listener
// is closed again before the port is handed back, so this trades a small race
// against anything else grabbing ports at that instant for tests that do not
// have to carve up a fixed range between themselves.
func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", net.JoinHostPort(loopback, "0"))
	if err != nil {
		t.Fatalf("error reserving a port: %s", err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

// renderCfg renders one of the templates in test_data into the test's temp
// directory and returns the path to the result.
func renderCfg(t *testing.T, name string, data gobgpdCfg) string {
	t.Helper()
	tmpl, err := template.ParseFiles(filepath.Join("test_data", name))
	if err != nil {
		t.Fatalf("error parsing %s: %s", name, err)
	}
	path := filepath.Join(t.TempDir(), strings.TrimSuffix(name, ".tmpl"))
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("error creating %s: %s", path, err)
	}
	defer f.Close()
	if err := tmpl.Execute(f, data); err != nil {
		t.Fatalf("error rendering %s: %s", name, err)
	}
	return path
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
	port     int
}

func (gc *gobpgclient) globalRib(t *testing.T) (map[string][]ribEntry, error) {
	out, err := exec.Command(gc.cmdPath,
		"-u", gc.hostAddr,
		"-p", strconv.Itoa(gc.port),
		"-j", "global", "rib").Output()
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
		"-f", gs.cfgFile,
		"--api-hosts", gs.apiHost,
		"--pprof-disable",
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
	// Nothing ever answers on the neighbour address, so it only has to be
	// something we can recognise again in the peer list.
	cfgFile := renderCfg(t, "bgp_cfgfile.toml.tmpl", gobgpdCfg{
		Addr:       loopback,
		ASN:        localASN,
		RouterID:   "192.0.2.6",
		ListenPort: freePort(t),
		PeerAddr:   "192.0.2.7",
		PeerASN:    65001,
	})
	cfg := &config.BGP{
		BGPEnabled:    true,
		RouterID:      loopback,
		GOBGPDCfgFile: cfgFile,
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
	if !slices.Contains(peers, "192.0.2.7") {
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
		RouterID:        loopback,
		ListenPort:      freePort(t),
		ListenAddresses: []string{loopback},
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

//go:build !windows
// +build !windows

package bgp

import (
	"context"
	"encoding/json"
	"github.com/fabiolb/fabio/config"
	api "github.com/osrg/gobgp/v3/api"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestBGPHandler(t *testing.T) {
	serverCmd := &gobgpserver{
		cmdPath: "gobgpd",
	}
	err := serverCmd.start()
	if err != nil {
		t.Logf("error calling gobgpd command, probably not installed. skipping: %s", err)
		t.SkipNow()
	}
	defer serverCmd.stop()
	cfg := &config.BGP{
		BGPEnabled:       true,
		Asn:              65000,
		AnycastAddresses: []string{"1.2.3.4/32"},
		RouterID:         "127.0.0.2",
		ListenPort:       1790,
		ListenAddresses:  []string{"127.0.0.2"},
		Peers: []config.BGPPeer{
			{
				NeighborAddress: "127.0.0.3",
				NeighborPort:    1790,
				Asn:             65001,
				MultiHop:        false,
			},
		},
		EnableGRPC:        true,
		GRPCListenAddress: "127.0.0.2:50051",
		NextHop:           "1.2.3.4",
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
	err = bh.addNeighbors(context.Background(), cfg.Peers)
	if err != nil {
		t.Fatalf("error adding neighbors: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	if err := bh.server.WatchEvent(context.Background(), &api.WatchEventRequest{Peer: &api.WatchEventRequest_Peer{}}, func(r *api.WatchEventResponse) {
		if p := r.GetPeer(); p != nil && p.Type == api.WatchEventResponse_PeerEvent_STATE {
			t.Logf("EVENT RECEIVED %#v", p.Peer)
			if p.Peer.State.SessionState == api.PeerState_ESTABLISHED {
				cancel()
			}
		}
	}); err != nil {
		t.Fatal(err)
	}

	<-ctx.Done()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatal("context deadline exceeded")
	}

	gc := gobpgclient{
		cmdPath:  "gobgp",
		hostAddr: "127.0.0.3",
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
				if _, ok := routes[r]; !ok {
					t.Fatalf("route %s not found", r)
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
	cmd     *exec.Cmd
}

func (gs *gobgpserver) start() error {
	gs.cmd = exec.Command(gs.cmdPath,
		"-p",
		"-f", filepath.Join("test_data", "bgp.toml"),
		"--api-hosts", "127.0.0.3:50051",
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

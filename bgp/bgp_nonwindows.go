//go:build !windows

package bgp

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/netip"
	"os"

	"github.com/fabiolb/fabio/config"
	"github.com/fabiolb/fabio/exit"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	api "github.com/osrg/gobgp/v4/api"
	"github.com/osrg/gobgp/v4/pkg/apiutil"
	bgpconfig "github.com/osrg/gobgp/v4/pkg/config/oc"
	"github.com/osrg/gobgp/v4/pkg/packet/bgp"
	"github.com/osrg/gobgp/v4/pkg/server"
)

var (
	ErrMissingAnycast  = errors.New("you must specify at least one anycast address to advertise")
	ErrMissingPeers    = errors.New("you must specify at least one peer to advertise routes to")
	ErrMissingRouterID = errors.New("you must specify the routerID of this host, i.e. a non anycast address")
	ErrNoRoutesAdded   = errors.New("no routes were successfully added")
	ErrNoRoutesDeleted = errors.New("no routes were successfully deleted")
	ErrNoPeersAdded    = errors.New("no peers were successfully added")
)

const (
	denyAllNeighbors = "deny-all-neighbors"
	matchAnyPeer     = "match-any-peer"
	globalTable      = "global"
	rejectAll        = "reject-all"
)

type BGPHandler struct {
	server     *server.BgpServer
	config     *config.BGP
	routeAttrs []bgp.PathAttributeInterface
}

func NewBGPHandler(config *config.BGP) (*BGPHandler, error) {
	// pre-chew path attributes that are part of
	// every anycast route we'll be adding.
	nextHop := config.RouterID
	if len(config.NextHop) > 0 {
		nextHop = config.NextHop
	}

	nextHopAddr, err := netip.ParseAddr(nextHop)
	if err != nil {
		return nil, fmt.Errorf("invalid next hop address %s: %w", nextHop, err)
	}

	nextHopAttr, err := bgp.NewPathAttributeNextHop(nextHopAddr)
	if err != nil {
		return nil, fmt.Errorf("error creating next hop attribute: %w", err)
	}

	attributes := []bgp.PathAttributeInterface{
		bgp.NewPathAttributeOrigin(0), // IGP
		nextHopAttr,
		bgp.NewPathAttributeAsPath([]bgp.AsPathParamInterface{
			bgp.NewAs4PathParam(bgp.BGP_ASPATH_ATTR_TYPE_SEQ, []uint32{uint32(config.Asn)}),
		}),
	}

	logger := newBGPLogger()
	levelVar := &slog.LevelVar{}
	levelVar.Set(slog.LevelInfo)
	var opts = []server.ServerOption{server.LoggerOption(logger, levelVar)}
	if config.EnableGRPC {
		maxSize := 256 << 20
		grpcOpts := []grpc.ServerOption{grpc.MaxRecvMsgSize(maxSize), grpc.MaxSendMsgSize(maxSize)}
		if config.GRPCTLS {
			creds, err := credentials.NewServerTLSFromFile(config.CertFile, config.KeyFile)
			if err != nil {
				// shouldn't get here if validate was called first.
				return nil, fmt.Errorf("error parsing bgp TLS credentials: %s", err)
			}
			grpcOpts = append(grpcOpts, grpc.Creds(creds))
		}
		opts = append(opts,
			server.GrpcOption(grpcOpts),
			server.GrpcListenAddress(config.GRPCListenAddress),
		)
	}
	return &BGPHandler{
		server:     server.NewBgpServer(opts...),
		config:     config,
		routeAttrs: attributes,
	}, nil
}

func (bgph *BGPHandler) Start() error {
	s := bgph.server
	go s.Serve()

	if len(bgph.config.GOBGPDCfgFile) > 0 {
		initialCfg, err := bgpconfig.ReadConfigfile(bgph.config.GOBGPDCfgFile, "toml")
		if err != nil {
			// shouldn't happen if we called validate first.
			return err
		}
		// Apply global config
		if err := s.StartBgp(context.Background(), &api.StartBgpRequest{
			Global: bgpconfig.NewGlobalFromConfigStruct(&initialCfg.Global),
		}); err != nil {
			return fmt.Errorf("bgp: error starting from gobgp config: %w", err)
		}
		// Note: Additional config like peers, policies would need to be applied separately
		log.Printf("[WARN] bgp: config file support is limited, only global config applied")
	} else {
		// If we weren't passed a gobgp config file, configure using the values passed from the fabio
		// config, and make sure we have a sane policy where we export our routes to peers but don't
		// import from any peers.
		err := bgph.startBGP(context.Background())
		if err != nil {
			return fmt.Errorf("bgp: error starting: %w", err)
		}

		err = bgph.setPolicies()
		if err != nil {
			return fmt.Errorf("bgp error setting policy: %w", err)
		}

	}

	errCh := make(chan error, 1)
	exit.Listen(func(sig os.Signal) {
		log.Printf("[INFO] Stopping BGP")
		err := s.StopBgp(context.Background(), &api.StopBgpRequest{})
		errCh <- err
	})

	if len(bgph.config.GOBGPDCfgFile) == 0 || len(bgph.config.Peers) > 0 {
		// add peers
		err := bgph.addNeighbors(context.Background(), bgph.config.Peers)
		if err != nil {
			return fmt.Errorf("bgp error adding neighbors: %w", err)
		}
	}
	if len(bgph.config.AnycastAddresses) > 0 {
		err := bgph.AddRoutes(context.Background(), bgph.config.AnycastAddresses)
		if err != nil {
			return fmt.Errorf("bgp error adding anycastaddresses: %w", err)
		}
	}
	// hang until exit handler completes above.
	return <-errCh
}

func (bgph *BGPHandler) startBGP(ctx context.Context) error {
	return bgph.server.StartBgp(ctx, &api.StartBgpRequest{
		Global: &api.Global{
			Asn:             uint32(bgph.config.Asn),
			RouterId:        bgph.config.RouterID,
			ListenPort:      int32(bgph.config.ListenPort),
			ListenAddresses: bgph.config.ListenAddresses,
		},
	})
}

func (bgph *BGPHandler) setPolicies() error {
	// Create a policy that denies all routes from any neighbor.
	err := bgph.server.SetPolicies(context.Background(), &api.SetPoliciesRequest{
		DefinedSets: []*api.DefinedSet{
			{
				DefinedType: api.DefinedType_DEFINED_TYPE_NEIGHBOR,
				Name:        matchAnyPeer,
				List:        []string{"0.0.0.0/0", "::/0"},
			},
		},
		Policies: []*api.Policy{
			{
				Name: denyAllNeighbors,
				Statements: []*api.Statement{
					{
						Name: rejectAll,
						Conditions: &api.Conditions{
							NeighborSet: &api.MatchSet{
								Name: matchAnyPeer,
							},
						},
						Actions: &api.Actions{
							RouteAction: api.RouteAction_ROUTE_ACTION_REJECT,
						},
					},
				},
			},
		},
	})
	if err != nil {
		return err
	}

	// Assign the above to the global policy
	return bgph.server.SetPolicyAssignment(context.Background(), &api.SetPolicyAssignmentRequest{
		Assignment: &api.PolicyAssignment{
			Name:      globalTable, // this is the global rib
			Direction: api.PolicyDirection_POLICY_DIRECTION_IMPORT,
			Policies: []*api.Policy{
				{
					Name: denyAllNeighbors,
				},
			},
			// Need to set default action to accept here because otherwise
			// even routes added via API calls get rejected.
			DefaultAction: api.RouteAction_ROUTE_ACTION_ACCEPT,
		},
	})
}

func (bgph *BGPHandler) addNeighbors(ctx context.Context, peers []config.BGPPeer) error {
	var errs []error
	peerCount := 0
	for _, peer := range peers {
		var hop *api.EbgpMultihop
		if peer.MultiHop {
			hop = &api.EbgpMultihop{
				Enabled:     true,
				MultihopTtl: uint32(peer.MultiHopLength),
			}
		}
		var trans *api.Transport
		if peer.NeighborPort > 0 {
			trans = &api.Transport{
				LocalAddress:  bgph.config.RouterID,
				MtuDiscovery:  false,
				PassiveMode:   false,
				RemoteAddress: peer.NeighborAddress,
				RemotePort:    uint32(peer.NeighborPort),
				TcpMss:        0,
				BindInterface: "",
			}
		}
		err := bgph.server.AddPeer(ctx, &api.AddPeerRequest{
			Peer: &api.Peer{
				Conf: &api.PeerConf{
					AuthPassword:    peer.Password,
					NeighborAddress: peer.NeighborAddress,
					PeerAsn:         uint32(peer.Asn),
				},
				EbgpMultihop: hop,
				Transport:    trans,
			},
		})
		if err != nil {
			errs = append(errs, err)
			continue
		}
		peerCount++
	}
	if peerCount == 0 {
		errs = append(errs, ErrNoPeersAdded)
	}
	return errors.Join(errs...)
}

func (bgph *BGPHandler) AddRoutes(ctx context.Context, routes []string) error {
	var errs []error
	// Add our Anycast routes

	routesAdded := 0
	var paths []*apiutil.Path

	for _, addr := range routes {
		prefix, err := netip.ParsePrefix(addr)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		var nlri bgp.NLRI
		var family bgp.Family
		if prefix.Addr().Is6() {
			family = bgp.RF_IPv6_UC
		} else {
			family = bgp.RF_IPv4_UC
		}
		nlri, err = bgp.NewIPAddrPrefix(prefix)
		if err != nil {
			errs = append(errs, fmt.Errorf("error creating NLRI for %s: %w", addr, err))
			continue
		}

		path := &apiutil.Path{
			Family: family,
			Nlri:   nlri,
			Attrs:  bgph.routeAttrs,
		}
		paths = append(paths, path)
	}

	if len(paths) == 0 {
		return errors.Join(append(errs, ErrNoRoutesAdded)...)
	}

	// Add all paths at once
	responses, err := bgph.server.AddPath(apiutil.AddPathRequest{
		VRFID: "",
		Paths: paths,
	})
	if err != nil {
		return fmt.Errorf("bgp error adding paths: %w", err)
	}

	// Check individual path results
	for i, resp := range responses {
		if resp.Error != nil {
			log.Printf("[ERROR] bgp error adding path for %s: %s", routes[i], resp.Error)
			errs = append(errs, fmt.Errorf("error adding %s: %w", routes[i], resp.Error))
		} else {
			log.Printf("[INFO] bgp successfully added path for %s", routes[i])
			routesAdded++
		}
	}

	if routesAdded == 0 {
		errs = append(errs, ErrNoRoutesAdded)
	}
	return errors.Join(errs...)
}

func (bgph *BGPHandler) DeleteRoutes(ctx context.Context, routes []string) error {
	var errs []error
	delCount := 0
	var paths []*apiutil.Path

	for _, addr := range routes {
		prefix, err := netip.ParsePrefix(addr)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		var nlri bgp.NLRI
		var family bgp.Family
		if prefix.Addr().Is6() {
			family = bgp.RF_IPv6_UC
		} else {
			family = bgp.RF_IPv4_UC
		}
		nlri, err = bgp.NewIPAddrPrefix(prefix)
		if err != nil {
			errs = append(errs, fmt.Errorf("error creating NLRI for %s: %w", addr, err))
			continue
		}

		path := &apiutil.Path{
			Family: family,
			Nlri:   nlri,
			Attrs:  bgph.routeAttrs,
		}
		paths = append(paths, path)
		delCount++
	}

	if delCount == 0 {
		return errors.Join(append(errs, ErrNoRoutesDeleted)...)
	}

	// Delete all paths at once
	err := bgph.server.DeletePath(apiutil.DeletePathRequest{
		VRFID: "",
		Paths: paths,
	})
	if err != nil {
		return fmt.Errorf("bgp error deleting paths: %w", err)
	}

	log.Printf("[INFO] bgp successfully deleted %d paths", len(paths))
	return errors.Join(errs...)
}

func ValidateConfig(config *config.BGP) error {
	if !config.BGPEnabled {
		return nil
	}

	for _, addr := range config.AnycastAddresses {
		_, _, err := net.ParseCIDR(addr)
		if err != nil {
			return fmt.Errorf("could not parse cidr for anycast address %s: %w", addr, err)
		}
	}

	if config.EnableGRPC && config.GRPCTLS {
		_, err := credentials.NewServerTLSFromFile(config.CertFile, config.KeyFile)
		if err != nil {
			return fmt.Errorf("could not parse bgp tls credentials: %w", err)
		}
	}

	for _, peer := range config.Peers {
		if _, err := netip.ParseAddr(peer.NeighborAddress); err != nil {
			return fmt.Errorf("peer address %s is not a valid IP: %w", peer.NeighborAddress, err)
		}
	}

	if len(config.GOBGPDCfgFile) > 0 {
		_, err := bgpconfig.ReadConfigfile(config.GOBGPDCfgFile, "toml")
		if err != nil {
			return fmt.Errorf("could not open %s: %w", config.GOBGPDCfgFile, err)
		}
		// otherwise we skip the rest of these checks, hopefully the provided bobgpd config is sane.
		return nil
	}

	if len(config.AnycastAddresses) == 0 {
		return ErrMissingAnycast
	}
	if len(config.Peers) == 0 {
		return ErrMissingPeers
	}

	if len(config.RouterID) == 0 {
		return ErrMissingRouterID
	}
	if _, err := netip.ParseAddr(config.RouterID); err != nil {
		return fmt.Errorf("router ID %s is not a valid IP: %w", config.RouterID, err)
	}
	if len(config.NextHop) > 0 {
		if _, err := netip.ParseAddr(config.NextHop); err != nil {
			return fmt.Errorf("invalid NextHop: %s: %w", config.NextHop, err)
		}
	}
	return nil
}

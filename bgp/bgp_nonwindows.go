//go:build !windows
// +build !windows

package bgp

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/fabiolb/fabio/config"
	"github.com/fabiolb/fabio/exit"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/proto"
	apb "google.golang.org/protobuf/types/known/anypb"

	api "github.com/osrg/gobgp/v3/api"
	bgpconfig "github.com/osrg/gobgp/v3/pkg/config"
	"github.com/osrg/gobgp/v3/pkg/server"
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
	routeAttrs []*apb.Any
}

func NewBGPHandler(config *config.BGP) (*BGPHandler, error) {
	// pre-chew some protobuf messages that are part of
	// every anycast route we'll be adding.
	nextHop := config.RouterID
	if len(config.NextHop) > 0 {
		nextHop = config.NextHop
	}
	var messages = []proto.Message{
		&api.OriginAttribute{
			Origin: 0,
		},
		&api.NextHopAttribute{
			NextHop: nextHop,
		},
		&api.AsPathAttribute{
			Segments: []*api.AsSegment{
				{
					Type:    api.AsSegment_AS_SEQUENCE,
					Numbers: []uint32{uint32(config.Asn)},
				},
			},
		},
	}
	attributes := make([]*apb.Any, 0, len(messages))
	for _, p := range messages {
		attr, err := apb.New(p)
		if err != nil {
			// should never happen
			panic(err)
		}
		attributes = append(attributes, attr)
	}

	var opts = []server.ServerOption{server.LoggerOption(bgpLogger{})}
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
		initialCfg, err := bgpconfig.ReadConfigFile(bgph.config.GOBGPDCfgFile, "toml")
		if err != nil {
			// shouldn't happen if we called validate first.
			return err
		}
		_, err = bgpconfig.InitialConfig(context.Background(), s, initialCfg, false)
		if err != nil {
			return fmt.Errorf("bgp: error initializing from gobgp config: %w", err)
		}
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

	// monitor the change of the peer state
	if err := s.WatchEvent(context.Background(), &api.WatchEventRequest{Peer: &api.WatchEventRequest_Peer{}}, func(r *api.WatchEventResponse) {
		if p := r.GetPeer(); p != nil && p.Type == api.WatchEventResponse_PeerEvent_STATE {
			log.Printf("[DEBUG] bgp event: %#v", p)
		}
	}); err != nil {
		log.Printf("[ERROR] bgp watcher failed: %s", err)
	}
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
				DefinedType: api.DefinedType_NEIGHBOR,
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
							RouteAction: api.RouteAction_REJECT,
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
			Direction: api.PolicyDirection_IMPORT,
			Policies: []*api.Policy{
				{
					Name: denyAllNeighbors,
				},
			},
			// Need to set default action to accept here because otherwise
			// even routes added via API calls get rejected.
			DefaultAction: api.RouteAction_ACCEPT,
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

	for _, addr := range routes {
		_, ipnet, err := net.ParseCIDR(addr)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		prefixLen, _ := ipnet.Mask.Size()
		af := api.Family_AFI_IP
		if ipnet.IP.To4() == nil {
			af = api.Family_AFI_IP6
		}
		nlri, _ := apb.New(&api.IPAddressPrefix{
			PrefixLen: uint32(prefixLen),
			Prefix:    ipnet.IP.String(),
		})
		_, err = bgph.server.AddPath(ctx, &api.AddPathRequest{
			Path: &api.Path{
				Nlri:   nlri,
				Pattrs: bgph.routeAttrs,
				Family: &api.Family{
					Afi:  af,
					Safi: api.Family_SAFI_UNICAST,
				},
			},
		})
		if err != nil {
			log.Printf("[ERROR] bgp error adding path for %s: %s", addr, err)
			errs = append(errs, fmt.Errorf("error adding %s: %w", addr, err))
		} else {
			log.Printf("[INFO] bgp successfully added path for %s", addr)
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
	for _, addr := range routes {
		_, ipnet, err := net.ParseCIDR(addr)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		prefixLen, _ := ipnet.Mask.Size()
		af := api.Family_AFI_IP
		if ipnet.IP.To4() == nil {
			af = api.Family_AFI_IP6
		}
		nlri, _ := apb.New(&api.IPAddressPrefix{
			PrefixLen: uint32(prefixLen),
			Prefix:    ipnet.IP.String(),
		})
		err = bgph.server.DeletePath(ctx, &api.DeletePathRequest{
			TableType: api.TableType_GLOBAL,
			Path: &api.Path{
				Nlri: nlri,
				Family: &api.Family{
					Afi:  af,
					Safi: api.Family_SAFI_UNICAST,
				},
				Pattrs: bgph.routeAttrs,
			},
		})
		if err != nil {
			errs = append(errs, err)
			continue
		}
		delCount++
	}
	if delCount == 0 {
		errs = append(errs, ErrNoRoutesDeleted)
	}
	return errors.Join(errs...)
}

func ValidateConfig(config *config.BGP) error {
	if config.BGPEnabled == false {
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
		if net.ParseIP(peer.NeighborAddress) == nil {
			return fmt.Errorf("peer address %s is not a valid IP", peer.NeighborAddress)
		}
	}

	if len(config.GOBGPDCfgFile) > 0 {
		_, err := bgpconfig.ReadConfigFile(config.GOBGPDCfgFile, "toml")
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
	if net.ParseIP(config.RouterID) == nil {
		return fmt.Errorf("router ID %s is not a valid ID", config.RouterID)
	}
	if len(config.NextHop) > 0 {
		if ip := net.ParseIP(config.NextHop); ip == nil {
			return fmt.Errorf("invalid NextHop: %s", config.NextHop)
		}
	}
	return nil
}

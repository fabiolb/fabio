package config

import (
	"golang.org/x/net/context"
	apb "google.golang.org/protobuf/types/known/anypb"

	api "github.com/osrg/gobgp/v3/api"
	"github.com/osrg/gobgp/v3/internal/pkg/config"
	"github.com/osrg/gobgp/v3/internal/pkg/table"
	"github.com/osrg/gobgp/v3/pkg/apiutil"
	"github.com/osrg/gobgp/v3/pkg/log"
	"github.com/osrg/gobgp/v3/pkg/packet/bgp"
	"github.com/osrg/gobgp/v3/pkg/server"
)

// ReadConfigFile parses a config file into a BgpConfigSet which can be applied
// using InitialConfig and UpdateConfig.
func ReadConfigFile(configFile, configType string) (*config.BgpConfigSet, error) {
	return config.ReadConfigfile(configFile, configType)
}

func marshalRouteTargets(l []string) ([]*apb.Any, error) {
	rtList := make([]*apb.Any, 0, len(l))
	for _, rtString := range l {
		rt, err := bgp.ParseRouteTarget(rtString)
		if err != nil {
			return nil, err
		}
		a, err := apiutil.MarshalRT(rt)
		if err != nil {
			return nil, err
		}
		rtList = append(rtList, a)
	}
	return rtList, nil
}

func assignGlobalpolicy(ctx context.Context, bgpServer *server.BgpServer, a *config.ApplyPolicyConfig) {
	toDefaultTable := func(r config.DefaultPolicyType) table.RouteType {
		var def table.RouteType
		switch r {
		case config.DEFAULT_POLICY_TYPE_ACCEPT_ROUTE:
			def = table.ROUTE_TYPE_ACCEPT
		case config.DEFAULT_POLICY_TYPE_REJECT_ROUTE:
			def = table.ROUTE_TYPE_REJECT
		}
		return def
	}
	toPolicies := func(r []string) []*table.Policy {
		p := make([]*table.Policy, 0, len(r))
		for _, n := range r {
			p = append(p, &table.Policy{
				Name: n,
			})
		}
		return p
	}

	def := toDefaultTable(a.DefaultImportPolicy)
	ps := toPolicies(a.ImportPolicyList)
	bgpServer.SetPolicyAssignment(ctx, &api.SetPolicyAssignmentRequest{
		Assignment: table.NewAPIPolicyAssignmentFromTableStruct(&table.PolicyAssignment{
			Name:     table.GLOBAL_RIB_NAME,
			Type:     table.POLICY_DIRECTION_IMPORT,
			Policies: ps,
			Default:  def,
		}),
	})

	def = toDefaultTable(a.DefaultExportPolicy)
	ps = toPolicies(a.ExportPolicyList)
	bgpServer.SetPolicyAssignment(ctx, &api.SetPolicyAssignmentRequest{
		Assignment: table.NewAPIPolicyAssignmentFromTableStruct(&table.PolicyAssignment{
			Name:     table.GLOBAL_RIB_NAME,
			Type:     table.POLICY_DIRECTION_EXPORT,
			Policies: ps,
			Default:  def,
		}),
	})

}

func addPeerGroups(ctx context.Context, bgpServer *server.BgpServer, addedPg []config.PeerGroup) {
	for _, pg := range addedPg {
		bgpServer.Log().Info("Add PeerGroup",
			log.Fields{
				"Topic": "config",
				"Key":   pg.Config.PeerGroupName,
			})

		if err := bgpServer.AddPeerGroup(ctx, &api.AddPeerGroupRequest{
			PeerGroup: config.NewPeerGroupFromConfigStruct(&pg),
		}); err != nil {
			bgpServer.Log().Warn("Failed to add PeerGroup",
				log.Fields{
					"Topic": "config",
					"Key":   pg.Config.PeerGroupName,
					"Error": err})
		}
	}
}

func deletePeerGroups(ctx context.Context, bgpServer *server.BgpServer, deletedPg []config.PeerGroup) {
	for _, pg := range deletedPg {
		bgpServer.Log().Info("delete PeerGroup",
			log.Fields{
				"Topic": "config",
				"Key":   pg.Config.PeerGroupName})
		if err := bgpServer.DeletePeerGroup(ctx, &api.DeletePeerGroupRequest{
			Name: pg.Config.PeerGroupName,
		}); err != nil {
			bgpServer.Log().Warn("Failed to delete PeerGroup",
				log.Fields{
					"Topic": "config",
					"Key":   pg.Config.PeerGroupName,
					"Error": err})
		}
	}
}

func updatePeerGroups(ctx context.Context, bgpServer *server.BgpServer, updatedPg []config.PeerGroup) bool {
	for _, pg := range updatedPg {
		bgpServer.Log().Info("update PeerGroup",
			log.Fields{
				"Topic": "config",
				"Key":   pg.Config.PeerGroupName})
		if u, err := bgpServer.UpdatePeerGroup(ctx, &api.UpdatePeerGroupRequest{
			PeerGroup: config.NewPeerGroupFromConfigStruct(&pg),
		}); err != nil {
			bgpServer.Log().Warn("Failed to update PeerGroup",
				log.Fields{
					"Topic": "config",
					"Key":   pg.Config.PeerGroupName,
					"Error": err})
		} else {
			return u.NeedsSoftResetIn
		}
	}
	return false
}

func addDynamicNeighbors(ctx context.Context, bgpServer *server.BgpServer, dynamicNeighbors []config.DynamicNeighbor) {
	for _, dn := range dynamicNeighbors {
		bgpServer.Log().Info("Add Dynamic Neighbor to PeerGroup",
			log.Fields{
				"Topic":  "config",
				"Key":    dn.Config.PeerGroup,
				"Prefix": dn.Config.Prefix})
		if err := bgpServer.AddDynamicNeighbor(ctx, &api.AddDynamicNeighborRequest{
			DynamicNeighbor: &api.DynamicNeighbor{
				Prefix:    dn.Config.Prefix,
				PeerGroup: dn.Config.PeerGroup,
			},
		}); err != nil {
			bgpServer.Log().Warn("Failed to add Dynamic Neighbor to PeerGroup",
				log.Fields{
					"Topic":  "config",
					"Key":    dn.Config.PeerGroup,
					"Prefix": dn.Config.Prefix,
					"Error":  err})
		}
	}
}

func addNeighbors(ctx context.Context, bgpServer *server.BgpServer, added []config.Neighbor) {
	for _, p := range added {
		bgpServer.Log().Info("Add Peer",
			log.Fields{
				"Topic": "config",
				"Key":   p.State.NeighborAddress})
		if err := bgpServer.AddPeer(ctx, &api.AddPeerRequest{
			Peer: config.NewPeerFromConfigStruct(&p),
		}); err != nil {
			bgpServer.Log().Warn("Failed to add Peer",
				log.Fields{
					"Topic": "config",
					"Key":   p.State.NeighborAddress,
					"Error": err})
		}
	}
}

func deleteNeighbors(ctx context.Context, bgpServer *server.BgpServer, deleted []config.Neighbor) {
	for _, p := range deleted {
		bgpServer.Log().Info("Delete Peer",
			log.Fields{
				"Topic": "config",
				"Key":   p.State.NeighborAddress})
		if err := bgpServer.DeletePeer(ctx, &api.DeletePeerRequest{
			Address: p.State.NeighborAddress,
		}); err != nil {
			bgpServer.Log().Warn("Failed to delete Peer",
				log.Fields{
					"Topic": "config",
					"Key":   p.State.NeighborAddress,
					"Error": err})
		}
	}
}

func updateNeighbors(ctx context.Context, bgpServer *server.BgpServer, updated []config.Neighbor) bool {
	for _, p := range updated {
		bgpServer.Log().Info("Update Peer",
			log.Fields{
				"Topic": "config",
				"Key":   p.State.NeighborAddress})
		if u, err := bgpServer.UpdatePeer(ctx, &api.UpdatePeerRequest{
			Peer: config.NewPeerFromConfigStruct(&p),
		}); err != nil {
			bgpServer.Log().Warn("Failed to update Peer",
				log.Fields{
					"Topic": "config",
					"Key":   p.State.NeighborAddress,
					"Error": err})
		} else {
			return u.NeedsSoftResetIn
		}
	}
	return false
}

// InitialConfig applies initial configuration to a pristine gobgp instance. It
// can only be called once for an instance. Subsequent changes to the
// configuration can be applied using UpdateConfig. The BgpConfigSet can be
// obtained by calling ReadConfigFile. If graceful restart behavior is desired,
// pass true for isGracefulRestart. Otherwise, pass false.
func InitialConfig(ctx context.Context, bgpServer *server.BgpServer, newConfig *config.BgpConfigSet, isGracefulRestart bool) (*config.BgpConfigSet, error) {
	if err := bgpServer.StartBgp(ctx, &api.StartBgpRequest{
		Global: config.NewGlobalFromConfigStruct(&newConfig.Global),
	}); err != nil {
		bgpServer.Log().Fatal("failed to set global config",
			log.Fields{"Topic": "config", "Error": err})
	}

	if newConfig.Zebra.Config.Enabled {
		tps := newConfig.Zebra.Config.RedistributeRouteTypeList
		l := make([]string, 0, len(tps))
		l = append(l, tps...)
		if err := bgpServer.EnableZebra(ctx, &api.EnableZebraRequest{
			Url:                  newConfig.Zebra.Config.Url,
			RouteTypes:           l,
			Version:              uint32(newConfig.Zebra.Config.Version),
			NexthopTriggerEnable: newConfig.Zebra.Config.NexthopTriggerEnable,
			NexthopTriggerDelay:  uint32(newConfig.Zebra.Config.NexthopTriggerDelay),
			MplsLabelRangeSize:   uint32(newConfig.Zebra.Config.MplsLabelRangeSize),
			SoftwareName:         newConfig.Zebra.Config.SoftwareName,
		}); err != nil {
			bgpServer.Log().Fatal("failed to set zebra config",
				log.Fields{"Topic": "config", "Error": err})
		}
	}

	if len(newConfig.Collector.Config.Url) > 0 {
		bgpServer.Log().Fatal("collector feature is not supported",
			log.Fields{"Topic": "config"})
	}

	for _, c := range newConfig.RpkiServers {
		if err := bgpServer.AddRpki(ctx, &api.AddRpkiRequest{
			Address:  c.Config.Address,
			Port:     c.Config.Port,
			Lifetime: c.Config.RecordLifetime,
		}); err != nil {
			bgpServer.Log().Fatal("failed to set rpki config",
				log.Fields{"Topic": "config", "Error": err})
		}
	}
	for _, c := range newConfig.BmpServers {
		if err := bgpServer.AddBmp(ctx, &api.AddBmpRequest{
			Address:           c.Config.Address,
			Port:              c.Config.Port,
			SysName:           c.Config.SysName,
			SysDescr:          c.Config.SysDescr,
			Policy:            api.AddBmpRequest_MonitoringPolicy(c.Config.RouteMonitoringPolicy.ToInt()),
			StatisticsTimeout: int32(c.Config.StatisticsTimeout),
		}); err != nil {
			bgpServer.Log().Fatal("failed to set bmp config",
				log.Fields{"Topic": "config", "Error": err})
		}
	}
	for _, vrf := range newConfig.Vrfs {
		rd, err := bgp.ParseRouteDistinguisher(vrf.Config.Rd)
		if err != nil {
			bgpServer.Log().Fatal("failed to load vrf rd config",
				log.Fields{"Topic": "config", "Error": err})
		}

		importRtList, err := marshalRouteTargets(vrf.Config.ImportRtList)
		if err != nil {
			bgpServer.Log().Fatal("failed to load vrf import rt config",
				log.Fields{"Topic": "config", "Error": err})
		}
		exportRtList, err := marshalRouteTargets(vrf.Config.ExportRtList)
		if err != nil {
			bgpServer.Log().Fatal("failed to load vrf export rt config",
				log.Fields{"Topic": "config", "Error": err})
		}

		a, err := apiutil.MarshalRD(rd)
		if err != nil {
			bgpServer.Log().Fatal("failed to set vrf config",
				log.Fields{"Topic": "config", "Error": err})
		}
		if err := bgpServer.AddVrf(ctx, &api.AddVrfRequest{
			Vrf: &api.Vrf{
				Name:     vrf.Config.Name,
				Rd:       a,
				Id:       uint32(vrf.Config.Id),
				ImportRt: importRtList,
				ExportRt: exportRtList,
			},
		}); err != nil {
			bgpServer.Log().Fatal("failed to set vrf config",
				log.Fields{"Topic": "config", "Error": err})
		}
	}
	for _, c := range newConfig.MrtDump {
		if len(c.Config.FileName) == 0 {
			continue
		}
		if err := bgpServer.EnableMrt(ctx, &api.EnableMrtRequest{
			Type:             api.EnableMrtRequest_DumpType(c.Config.DumpType.ToInt()),
			Filename:         c.Config.FileName,
			DumpInterval:     c.Config.DumpInterval,
			RotationInterval: c.Config.RotationInterval,
		}); err != nil {
			bgpServer.Log().Fatal("failed to set mrt config",
				log.Fields{"Topic": "config", "Error": err})
		}
	}
	p := config.ConfigSetToRoutingPolicy(newConfig)
	rp, err := table.NewAPIRoutingPolicyFromConfigStruct(p)
	if err != nil {
		bgpServer.Log().Fatal("failed to update policy config",
			log.Fields{"Topic": "config", "Error": err})
	} else {
		bgpServer.SetPolicies(ctx, &api.SetPoliciesRequest{
			DefinedSets: rp.DefinedSets,
			Policies:    rp.Policies,
		})
	}

	assignGlobalpolicy(ctx, bgpServer, &newConfig.Global.ApplyPolicy.Config)

	added := newConfig.Neighbors
	addedPg := newConfig.PeerGroups
	if isGracefulRestart {
		for i, n := range added {
			if n.GracefulRestart.Config.Enabled {
				added[i].GracefulRestart.State.LocalRestarting = true
			}
		}
	}

	addPeerGroups(ctx, bgpServer, addedPg)
	addDynamicNeighbors(ctx, bgpServer, newConfig.DynamicNeighbors)
	addNeighbors(ctx, bgpServer, added)
	return newConfig, nil
}

// UpdateConfig updates the configuration of a running gobgp instance.
// InitialConfig must have been called once before this can be called for
// subsequent changes to config. The differences are that this call 1) does not
// hangle graceful restart and 2) requires a BgpConfigSet for the previous
// configuration so that it can compute the delta between it and the new
// config. The new BgpConfigSet can be obtained using ReadConfigFile.
func UpdateConfig(ctx context.Context, bgpServer *server.BgpServer, c, newConfig *config.BgpConfigSet) (*config.BgpConfigSet, error) {
	addedPg, deletedPg, updatedPg := config.UpdatePeerGroupConfig(bgpServer.Log(), c, newConfig)
	added, deleted, updated := config.UpdateNeighborConfig(bgpServer.Log(), c, newConfig)
	updatePolicy := config.CheckPolicyDifference(bgpServer.Log(), config.ConfigSetToRoutingPolicy(c), config.ConfigSetToRoutingPolicy(newConfig))

	if updatePolicy {
		bgpServer.Log().Info("policy config is update", log.Fields{"Topic": "config"})
		p := config.ConfigSetToRoutingPolicy(newConfig)
		rp, err := table.NewAPIRoutingPolicyFromConfigStruct(p)
		if err != nil {
			bgpServer.Log().Warn("failed to update policy config",
				log.Fields{
					"Topic": "config",
					"Error": err})
		} else {
			bgpServer.SetPolicies(ctx, &api.SetPoliciesRequest{
				DefinedSets: rp.DefinedSets,
				Policies:    rp.Policies,
			})
		}
	}
	// global policy update
	if !newConfig.Global.ApplyPolicy.Config.Equal(&c.Global.ApplyPolicy.Config) {
		assignGlobalpolicy(ctx, bgpServer, &newConfig.Global.ApplyPolicy.Config)
		updatePolicy = true
	}

	addPeerGroups(ctx, bgpServer, addedPg)
	deletePeerGroups(ctx, bgpServer, deletedPg)
	needsSoftResetIn := updatePeerGroups(ctx, bgpServer, updatedPg)
	updatePolicy = updatePolicy || needsSoftResetIn
	addDynamicNeighbors(ctx, bgpServer, newConfig.DynamicNeighbors)
	addNeighbors(ctx, bgpServer, added)
	deleteNeighbors(ctx, bgpServer, deleted)
	needsSoftResetIn = updateNeighbors(ctx, bgpServer, updated)
	updatePolicy = updatePolicy || needsSoftResetIn

	if updatePolicy {
		if err := bgpServer.ResetPeer(ctx, &api.ResetPeerRequest{
			Address:   "",
			Direction: api.ResetPeerRequest_IN,
			Soft:      true,
		}); err != nil {
			bgpServer.Log().Fatal("failed to update policy config",
				log.Fields{"Topic": "config", "Error": err})
		}
	}
	return newConfig, nil
}

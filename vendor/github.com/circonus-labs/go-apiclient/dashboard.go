// Copyright 2016 Circonus, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Dashboard API support - Fetch, Create, Update, Delete, and Search
// See: https://login.circonus.com/resources/api/calls/dashboard

package apiclient

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/circonus-labs/go-apiclient/config"
	"github.com/pkg/errors"
)

// DashboardGridLayout defines layout
type DashboardGridLayout struct {
	Height uint `json:"height"`
	Width  uint `json:"width"`
}

// DashboardAccessConfig defines access config
type DashboardAccessConfig struct {
	Nickname            string `json:"nickname"`
	SharedID            string `json:"shared_id"`
	TextSize            uint   `json:"text_size"`
	BlackDash           bool   `json:"black_dash"`
	Enabled             bool   `json:"enabled"`
	Fullscreen          bool   `json:"fullscreen"`
	FullscreenHideTitle bool   `json:"fullscreen_hide_title"`
	ScaleText           bool   `json:"scale_text"`
}

// DashboardOptions defines options
type DashboardOptions struct {
	AccessConfigs       []DashboardAccessConfig `json:"access_configs"`
	Linkages            [][]string              `json:"linkages"`
	TextSize            uint                    `json:"text_size"`
	FullscreenHideTitle bool                    `json:"fullscreen_hide_title"`
	HideGrid            bool                    `json:"hide_grid"`
	ScaleText           bool                    `json:"scale_text"`
}

// ChartTextWidgetDatapoint defines datapoints for charts
type ChartTextWidgetDatapoint struct {
	AccountID    string `json:"account_id,omitempty"`     // metric cluster, metric
	ClusterTitle string `json:"_cluster_title,omitempty"` // metric cluster
	Label        string `json:"label,omitempty"`          // metric
	Label2       string `json:"_label,omitempty"`         // metric cluster
	Metric       string `json:"metric,omitempty"`         // metric
	MetricType   string `json:"_metric_type,omitempty"`   // metric
	CheckID      uint   `json:"_check_id,omitempty"`      // metric
	ClusterID    uint   `json:"cluster_id,omitempty"`     // metric cluster
	NumericOnly  bool   `json:"numeric_only,omitempty"`   // metric cluster
}

// ChartWidgetDefinitionLegend defines chart widget definition legend
type ChartWidgetDefinitionLegend struct {
	Type string `json:"type,omitempty"`
	Show bool   `json:"show,omitempty"`
}

// ChartWidgetWedgeLabels defines chart widget wedge labels
type ChartWidgetWedgeLabels struct {
	OnChart  bool `json:"on_chart,omitempty"`
	ToolTips bool `json:"tooltips,omitempty"`
}

// ChartWidgetWedgeValues defines chart widget wedge values
type ChartWidgetWedgeValues struct {
	Angle string `json:"angle,omitempty"`
	Color string `json:"color,omitempty"`
	Show  bool   `json:"show,omitempty"`
}

// ChartWidgtDefinition defines chart widget definition
type ChartWidgtDefinition struct {
	Datasource        string                      `json:"datasource,omitempty"`
	Derive            string                      `json:"derive,omitempty"`
	Formula           string                      `json:"formula,omitempty"`
	WedgeValues       ChartWidgetWedgeValues      `json:"wedge_values,omitempty"`
	Legend            ChartWidgetDefinitionLegend `json:"legend,omitempty"`
	Period            uint                        `json:"period,omitempty"`
	WedgeLabels       ChartWidgetWedgeLabels      `json:"wedge_labels,omitempty"`
	DisableAutoformat bool                        `json:"disable_autoformat,omitempty"`
	PopOnHover        bool                        `json:"pop_onhover,omitempty"`
}

// ForecastGaugeWidgetThresholds defines forecast widget thresholds
type ForecastGaugeWidgetThresholds struct {
	Colors []string `json:"colors,omitempty"` // forecasts, gauges
	Values []string `json:"values,omitempty"` // forecasts, gauges
	Flip   bool     `json:"flip"`             // gauges 2019-11-01, flip is required
}

// StatusWidgetAgentStatusSettings defines agent status settings
type StateWidgetBadRulesSettings struct {
	Value     string `json:"value,omitempty"`
	Criterion string `json:"criterion,omitempty"`
	Color     string `json:"color,omitempty"`
}

// StatusWidgetAgentStatusSettings defines agent status settings
type StatusWidgetAgentStatusSettings struct {
	Search         string `json:"search,omitempty"`
	ShowAgentTypes string `json:"show_agent_types,omitempty"`
	ShowContact    bool   `json:"show_contact,omitempty"`
	ShowFeeds      bool   `json:"show_feeds,omitempty"`
	ShowSetup      bool   `json:"show_setup,omitempty"`
	ShowSkew       bool   `json:"show_skew,omitempty"`
	ShowUpdates    bool   `json:"show_updates,omitempty"`
}

// StatusWidgetHostStatusSettings defines host status settings
type StatusWidgetHostStatusSettings struct {
	LayoutStyle  string   `json:"layout_style,omitempty"`
	Search       string   `json:"search,omitempty"`
	SortBy       string   `json:"sort_by,omitempty"`
	TagFilterSet []string `json:"tag_filter_set,omitempty"`
}

// DashboardWidgetSettings defines settings specific to widget
// Note: optional attributes which are structs need to be pointers so they will be omitted
type DashboardWidgetSettings struct {
	Definition          *ChartWidgtDefinition            `json:"definition,omitempty"`            // charts
	AgentStatusSettings *StatusWidgetAgentStatusSettings `json:"agent_status_settings,omitempty"` // status
	HostStatusSettings  *StatusWidgetHostStatusSettings  `json:"host_status_settings,omitempty"`  // status
	Thresholds          *ForecastGaugeWidgetThresholds   `json:"thresholds,omitempty"`            // forecasts, gauges
	RangeHigh           *int                             `json:"range_high,omitempty"`            // gauges 2019-11-01 switch to pointer for 0 ranges
	RangeLow            *int                             `json:"range_low,omitempty"`             // gauges 2019-11-01 switch to pointer for 0 ranges
	ShowValue           *bool                            `json:"show_value,omitempty"`            // state
	AccountID           string                           `json:"account_id,omitempty"`            // alerts, clusters, gauges, graphs, lists, state, status
	Acknowledged        string                           `json:"acknowledged,omitempty"`          // alerts
	Algorithm           string                           `json:"algorithm,omitempty"`             // clusters
	BodyFormat          string                           `json:"body_format,omitempty"`           // text
	Caql                string                           `json:"caql,omitempty"`                  // state
	ChartType           string                           `json:"chart_type,omitempty"`            // charts
	CheckUUID           string                           `json:"check_uuid,omitempty"`            // gauges, state
	Cleared             string                           `json:"cleared,omitempty"`               // alerts
	ClusterName         string                           `json:"cluster_name,omitempty"`          // clusters
	ContentType         string                           `json:"content_type,omitempty"`          // status
	DateWindow          string                           `json:"date_window,omitempty"`           // graphs
	Dependents          string                           `json:"dependents,omitempty"`            // alerts
	Display             string                           `json:"display,omitempty"`               // alerts
	DisplayMarkup       string                           `json:"display_markup,omitempty"`        // state
	Format              string                           `json:"format,omitempty"`                // forecasts
	Formula             string                           `json:"formula,omitempty"`               // gauges
	GoodColor           string                           `json:"good_color,omitempty"`            // state
	GraphUUID           string                           `json:"graph_id,omitempty"`              // graphs
	KeyLoc              string                           `json:"key_loc,omitempty"`               // graphs
	Label               string                           `json:"label,omitempty"`                 // graphs
	Layout              string                           `json:"layout,omitempty"`                // clusters
	LayoutStyle         string                           `json:"layout_style,omitempty"`          // state
	LinkURL             string                           `json:"link_url,omitempty"`              // state
	Maintenance         string                           `json:"maintenance,omitempty"`           // alerts
	Markup              string                           `json:"markup,omitempty"`                // html
	MetricDisplayName   string                           `json:"metric_display_name,omitempty"`   // gauges, state
	MetricName          string                           `json:"metric_name,omitempty"`           // gauges, state
	MetricType          string                           `json:"metric_type,omitempty"`           // state
	MinAge              string                           `json:"min_age,omitempty"`               // alerts
	OverlaySetID        string                           `json:"overlay_set_id,omitempty"`        // graphs
	ResourceLimit       string                           `json:"resource_limit,omitempty"`        // forecasts
	ResourceUsage       string                           `json:"resource_usage,omitempty"`        // forecasts
	Search              string                           `json:"search,omitempty"`                // alerts, lists
	Severity            string                           `json:"severity,omitempty"`              // alerts
	Size                string                           `json:"size,omitempty"`                  // clusters
	TextAlign           string                           `json:"text_align,omitempty"`            // state
	TimeWindow          string                           `json:"time_window,omitempty"`           // alerts
	Title               string                           `json:"title,omitempty"`                 // alerts, charts, forecasts, gauges, html, state
	TitleFormat         string                           `json:"title_format,omitempty"`          // text
	Trend               string                           `json:"trend,omitempty"`                 // forecasts
	Type                string                           `json:"type,omitempty"`                  // gauges, lists
	ValueType           string                           `json:"value_type,omitempty"`            // gauges, text
	TagFilterSet        []string                         `json:"tag_filter_set,omitempty"`        // alerts
	WeekDays            []string                         `json:"weekdays,omitempty"`              // alerts
	BadRules            []StateWidgetBadRulesSettings    `json:"bad_rules,omitempty"`             // state
	Datapoints          []ChartTextWidgetDatapoint       `json:"datapoints,omitempty"`            // charts, text
	ContactGroups       []uint                           `json:"contact_groups,omitempty"`        // alerts
	OffHours            []uint                           `json:"off_hours,omitempty"`             // alerts
	KeySize             uint                             `json:"key_size,omitempty"`              // graphs
	Limit               uint                             `json:"limit,omitempty"`                 // lists
	Period              uint                             `json:"period,omitempty"`                // gauges, text, graphs
	ClusterID           uint                             `json:"cluster_id,omitempty"`            // clusters
	Threshold           float32                          `json:"threshold,omitempty"`             // clusters
	KeyWrap             bool                             `json:"key_wrap,omitempty"`              // graphs
	KeyInline           bool                             `json:"key_inline,omitempty"`            // graphs
	DisableAutoformat   bool                             `json:"disable_autoformat,omitempty"`    // gauges
	HideXAxis           bool                             `json:"hide_xaxis,omitempty"`            // graphs
	HideYAxis           bool                             `json:"hide_yaxis,omitempty"`            // graphs
	Realtime            bool                             `json:"realtime,omitempty"`              // graphs
	ShowFlags           bool                             `json:"show_flags,omitempty"`            // graphs
	UseDefault          bool                             `json:"use_default,omitempty"`           // text
	Autoformat          bool                             `json:"autoformat,omitempty"`            // text
}

// DashboardWidget defines widget
type DashboardWidget struct {
	Name     string                  `json:"name"`
	Origin   string                  `json:"origin"`
	Type     string                  `json:"type"`
	WidgetID string                  `json:"widget_id"`
	Settings DashboardWidgetSettings `json:"settings"`
	Height   uint                    `json:"height"`
	Width    uint                    `json:"width"`
	Active   bool                    `json:"active"`
}

// Dashboard defines a dashboard. See https://login.circonus.com/resources/api/calls/dashboard for more information.
type Dashboard struct {
	CID            string              `json:"_cid,omitempty"`
	CreatedBy      string              `json:"_created_by,omitempty"`
	Title          string              `json:"title"`
	UUID           string              `json:"_dashboard_uuid,omitempty"`
	Widgets        []DashboardWidget   `json:"widgets"`
	Options        DashboardOptions    `json:"options"`
	GridLayout     DashboardGridLayout `json:"grid_layout"`
	Created        uint                `json:"_created,omitempty"`
	LastModified   uint                `json:"_last_modified,omitempty"`
	AccountDefault bool                `json:"account_default"`
	Active         bool                `json:"_active,omitempty"`
	Shared         bool                `json:"shared"`
}

// NewDashboard returns a new Dashboard (with defaults, if applicable)
func NewDashboard() *Dashboard {
	return &Dashboard{}
}

// FetchDashboard retrieves dashboard with passed cid.
func (a *API) FetchDashboard(cid CIDType) (*Dashboard, error) {
	if cid == nil || *cid == "" {
		return nil, errors.New("invalid dashboard CID (none)")
	}

	var dashboardCID string
	if !strings.HasPrefix(*cid, config.DashboardPrefix) {
		dashboardCID = fmt.Sprintf("%s/%s", config.DashboardPrefix, *cid)
	} else {
		dashboardCID = *cid
	}

	matched, err := regexp.MatchString(config.DashboardCIDRegex, dashboardCID)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, errors.Errorf("invalid dashboard CID (%s)", dashboardCID)
	}

	result, err := a.Get(dashboardCID)
	if err != nil {
		return nil, errors.Wrap(err, "fetching dashobard")
	}

	if a.Debug {
		a.Log.Printf("fetch dashboard, received JSON: %s", string(result))
	}

	dashboard := new(Dashboard)
	if err := json.Unmarshal(result, dashboard); err != nil {
		return nil, errors.Wrap(err, "parsing dashboard")
	}

	return dashboard, nil
}

// FetchDashboards retrieves all dashboards available to the API Token.
func (a *API) FetchDashboards() (*[]Dashboard, error) {
	result, err := a.Get(config.DashboardPrefix)
	if err != nil {
		return nil, errors.Wrap(err, "fetching dashboards")
	}

	var dashboards []Dashboard
	if err := json.Unmarshal(result, &dashboards); err != nil {
		return nil, errors.Wrap(err, "parsing dashboards")
	}

	return &dashboards, nil
}

// UpdateDashboard updates passed dashboard.
func (a *API) UpdateDashboard(cfg *Dashboard) (*Dashboard, error) {
	if cfg == nil {
		return nil, errors.New("invalid dashboard config (nil)")
	}

	dashboardCID := cfg.CID

	matched, err := regexp.MatchString(config.DashboardCIDRegex, dashboardCID)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, errors.Errorf("invalid dashboard CID (%s)", dashboardCID)
	}

	jsonCfg, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	if a.Debug {
		a.Log.Printf("update dashboard, sending JSON: %s", string(jsonCfg))
	}

	result, err := a.Put(dashboardCID, jsonCfg)
	if err != nil {
		return nil, errors.Wrap(err, "updating dashobard")
	}

	dashboard := &Dashboard{}
	if err := json.Unmarshal(result, dashboard); err != nil {
		return nil, errors.Wrap(err, "parsing dashboard")
	}

	return dashboard, nil
}

// CreateDashboard creates a new dashboard.
func (a *API) CreateDashboard(cfg *Dashboard) (*Dashboard, error) {
	if cfg == nil {
		return nil, errors.New("invalid dashboard config (nil)")
	}

	jsonCfg, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	if a.Debug {
		a.Log.Printf("create dashboard, sending JSON: %s", string(jsonCfg))
	}

	result, err := a.Post(config.DashboardPrefix, jsonCfg)
	if err != nil {
		return nil, errors.Wrap(err, "creating dashboard")
	}

	dashboard := &Dashboard{}
	if err := json.Unmarshal(result, dashboard); err != nil {
		return nil, errors.Wrap(err, "parsing dashboard")
	}

	return dashboard, nil
}

// DeleteDashboard deletes passed dashboard.
func (a *API) DeleteDashboard(cfg *Dashboard) (bool, error) {
	if cfg == nil {
		return false, errors.New("invalid dashboard config (nil)")
	}
	return a.DeleteDashboardByCID(CIDType(&cfg.CID))
}

// DeleteDashboardByCID deletes dashboard with passed cid.
func (a *API) DeleteDashboardByCID(cid CIDType) (bool, error) {
	if cid == nil || *cid == "" {
		return false, errors.New("invalid dashboard CID (none)")
	}

	var dashboardCID string
	if !strings.HasPrefix(*cid, config.DashboardPrefix) {
		dashboardCID = fmt.Sprintf("%s/%s", config.DashboardPrefix, *cid)
	} else {
		dashboardCID = *cid
	}

	matched, err := regexp.MatchString(config.DashboardCIDRegex, dashboardCID)
	if err != nil {
		return false, err
	}
	if !matched {
		return false, errors.Errorf("invalid dashboard CID (%s)", dashboardCID)
	}

	_, err = a.Delete(dashboardCID)
	if err != nil {
		return false, errors.Wrap(err, "deleting dashboard")
	}

	return true, nil
}

// SearchDashboards returns dashboards matching the specified
// search query and/or filter. If nil is passed for both parameters
// all dashboards will be returned.
func (a *API) SearchDashboards(searchCriteria *SearchQueryType, filterCriteria *SearchFilterType) (*[]Dashboard, error) {
	q := url.Values{}

	if searchCriteria != nil && *searchCriteria != "" {
		q.Set("search", string(*searchCriteria))
	}

	if filterCriteria != nil && len(*filterCriteria) > 0 {
		for filter, criteria := range *filterCriteria {
			for _, val := range criteria {
				q.Add(filter, val)
			}
		}
	}

	if q.Encode() == "" {
		return a.FetchDashboards()
	}

	reqURL := url.URL{
		Path:     config.DashboardPrefix,
		RawQuery: q.Encode(),
	}

	result, err := a.Get(reqURL.String())
	if err != nil {
		return nil, errors.Wrap(err, "searching dashboards")
	}

	var dashboards []Dashboard
	if err := json.Unmarshal(result, &dashboards); err != nil {
		return nil, errors.Wrap(err, "parsing dashobards")
	}

	return &dashboards, nil
}

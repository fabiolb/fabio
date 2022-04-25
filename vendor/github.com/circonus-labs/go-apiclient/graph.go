// Copyright 2016 Circonus, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Graph API support - Fetch, Create, Update, Delete, and Search
// See: https://login.circonus.com/resources/api/calls/graph

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

// GraphAccessKey defines an access key for a graph
type GraphAccessKey struct {
	Key            string `json:"key,omitempty"`              // string
	LockMode       string `json:"lock_mode,omitempty"`        // string
	LockZoom       string `json:"lock_zoom,omitempty"`        // string
	Nickname       string `json:"nickname,omitempty"`         // string
	Height         uint   `json:"height,omitempty"`           // uint
	LockRangeEnd   uint   `json:"lock_range_end,omitempty"`   // uint
	LockRangeStart uint   `json:"lock_range_start,omitempty"` // uint
	Width          uint   `json:"width,omitempty"`            // uint
	Active         bool   `json:"active,omitempty"`           // boolean
	Legend         bool   `json:"legend,omitempty"`           // boolean
	LockDate       bool   `json:"lock_date,omitempty"`        // boolean
	LockShowTimes  bool   `json:"lock_show_times,omitempty"`  // boolean
	Title          bool   `json:"title,omitempty"`            // boolean
	XLabels        bool   `json:"x_labels,omitempty"`         // boolean
	YLabels        bool   `json:"y_labels,omitempty"`         // boolean
}

// GraphComposite defines a composite
type GraphComposite struct {
	Stack         *uint   `json:"stack"`          // uint or null
	DataFormula   *string `json:"data_formula"`   // string or null
	LegendFormula *string `json:"legend_formula"` // string or null
	Axis          string  `json:"axis"`           // string
	Color         string  `json:"color"`          // string
	Name          string  `json:"name"`           // string
	Hidden        bool    `json:"hidden"`         // boolean
}

// GraphDatapoint defines a datapoint
type GraphDatapoint struct {
	Derive        interface{} `json:"derive,omitempty"`      // BUG doc: string, api: string or boolean(for caql statements)
	Search        *string     `json:"search"`                // string or null
	Alpha         *string     `json:"alpha,omitempty"`       // BUG: doc: floating point number, api: string
	CAQL          *string     `json:"caql,omitempty"`        // string or null
	Color         *string     `json:"color,omitempty"`       // string
	DataFormula   *string     `json:"data_formula"`          // string or null
	LegendFormula *string     `json:"legend_formula"`        // string or null
	Stack         *uint       `json:"stack"`                 // uint or null
	Axis          string      `json:"axis,omitempty"`        // string
	MetricName    string      `json:"metric_name,omitempty"` // string
	MetricType    string      `json:"metric_type,omitempty"` // string
	Name          string      `json:"name"`                  // string
	CheckID       uint        `json:"check_id,omitempty"`    // uint
	Hidden        bool        `json:"hidden"`                // boolean
}

// GraphGuide defines a guide
type GraphGuide struct {
	DataFormula   *string `json:"data_formula"`   // string or null
	LegendFormula *string `json:"legend_formula"` // string or null
	Color         string  `json:"color"`          // string
	Name          string  `json:"name"`           // string
	Hidden        bool    `json:"hidden"`         // boolean
}

// GraphMetricCluster defines a metric cluster
type GraphMetricCluster struct {
	Color         *string `json:"color,omitempty"`              // string
	DataFormula   *string `json:"data_formula"`                 // string or null
	LegendFormula *string `json:"legend_formula"`               // string or null
	Stack         *uint   `json:"stack"`                        // uint or null
	MetricCluster string  `json:"metric_cluster,omitempty"`     // string
	Name          string  `json:"name,omitempty"`               // string
	AggregateFunc string  `json:"aggregate_function,omitempty"` // string
	Axis          string  `json:"axis,omitempty"`               // string
	Hidden        bool    `json:"hidden"`                       // boolean
}

// OverlaySet defines an overlay set for a graph
type GraphOverlaySet struct {
	Overlays map[string]GraphOverlay `json:"overlays"`
	Title    string                  `json:"title"`
}

// GraphOverlay defines a single overlay in an overlay set
type GraphOverlay struct {
	DataOpts OverlayDataOptions `json:"data_opts,omitempty"` // OverlayDataOptions
	ID       string             `json:"id,omitempty"`        // string
	Title    string             `json:"title,omitempty"`     // string
	UISpecs  OverlayUISpecs     `json:"ui_specs,omitempty"`  // OverlayUISpecs
}

// OverlayUISpecs defines UI specs for overlay
type OverlayUISpecs struct {
	ID       string `json:"id,omitempty"`       // string
	Label    string `json:"label,omitempty"`    // string
	Type     string `json:"type,omitempty"`     // string
	Z        string `json:"z,omitempty"`        // int encoded as string BUG doc: numeric, api: string
	Decouple bool   `json:"decouple,omitempty"` // boolean
}

// OverlayDataOptions defines overlay options for data. Note, each overlay type requires
// a _subset_ of the options. See Graph API documentation (URL above) for details.
type OverlayDataOptions struct {
	Alerts        string `json:"alerts,omitempty"`         // int encoded as string BUG doc: numeric, api: string
	ArrayOutput   string `json:"array_output,omitempty"`   // int encoded as string BUG doc: numeric, api: string
	BasePeriod    string `json:"base_period,omitempty"`    // int encoded as string BUG doc: numeric, api: string
	Delay         string `json:"delay,omitempty"`          // int encoded as string BUG doc: numeric, api: string
	Extension     string `json:"extension,omitempty"`      // string
	GraphTitle    string `json:"graph_title,omitempty"`    // string
	GraphUUID     string `json:"graph_id,omitempty"`       // string
	InPercent     string `json:"in_percent,omitempty"`     // boolean encoded as string BUG doc: boolean, api: string
	Inverse       string `json:"inverse,omitempty"`        // int encoded as string BUG doc: numeric, api: string
	Method        string `json:"method,omitempty"`         // string
	Model         string `json:"model,omitempty"`          // string
	ModelEnd      string `json:"model_end,omitempty"`      // string
	ModelPeriod   string `json:"model_period,omitempty"`   // string
	ModelRelative string `json:"model_relative,omitempty"` // int encoded as string BUG doc: numeric, api: string
	Out           string `json:"out,omitempty"`            // string
	Prequel       string `json:"prequel,omitempty"`        // int
	Presets       string `json:"presets,omitempty"`        // string
	Quantiles     string `json:"quantiles,omitempty"`      // string
	SeasonLength  string `json:"season_length,omitempty"`  // int encoded as string BUG doc: numeric, api: string
	Sensitivity   string `json:"sensitivity,omitempty"`    // int encoded as string BUG doc: numeric, api: string
	SingleValue   string `json:"single_value,omitempty"`   // int encoded as string BUG doc: numeric, api: string
	TargetPeriod  string `json:"target_period,omitempty"`  // string
	TimeOffset    string `json:"time_offset,omitempty"`    // string
	TimeShift     string `json:"time_shift,omitempty"`     // int encoded as string BUG doc: numeric, api: string
	Transform     string `json:"transform,omitempty"`      // string
	Version       string `json:"version,omitempty"`        // int encoded as string BUG doc: numeric, api: string
	Window        string `json:"window,omitempty"`         // int encoded as string BUG doc: numeric, api: string
	XShift        string `json:"x_shift,omitempty"`        // string
}

// Graph defines a graph. See https://login.circonus.com/resources/api/calls/graph for more information.
type Graph struct {
	LineStyle      *string                     `json:"line_style"`                           // string or null
	Style          *string                     `json:"style"`                                // string or null
	Notes          *string                     `json:"notes,omitempty"`                      // string or null
	LogLeftY       *int                        `json:"logarithmic_left_y,string,omitempty"`  // int encoded as string or null BUG doc: number (not string)
	LogRightY      *int                        `json:"logarithmic_right_y,string,omitempty"` // int encoded as string or null BUG doc: number (not string)
	MaxLeftY       *float64                    `json:"max_left_y,string,omitempty"`          // float64 encoded as string or null BUG doc: number (not string)
	MaxRightY      *float64                    `json:"max_right_y,string,omitempty"`         // float64 encoded as string or null BUG doc: number (not string)
	MinLeftY       *float64                    `json:"min_left_y,string,omitempty"`          // float64 encoded as string or null BUG doc: number (not string)
	MinRightY      *float64                    `json:"min_right_y,string,omitempty"`         // float64 encoded as string or null BUG doc: number (not string)
	OverlaySets    *map[string]GraphOverlaySet `json:"overlay_sets,omitempty"`               // GroupOverLaySets or null
	CID            string                      `json:"_cid,omitempty"`                       // string
	Description    string                      `json:"description,omitempty"`                // string
	Title          string                      `json:"title,omitempty"`                      // string
	Tags           []string                    `json:"tags,omitempty"`                       // [] len >= 0
	AccessKeys     []GraphAccessKey            `json:"access_keys,omitempty"`                // [] len >= 0
	Composites     []GraphComposite            `json:"composites,omitempty"`                 // [] len >= 0
	Datapoints     []GraphDatapoint            `json:"datapoints,omitempty"`                 // [] len >= 0
	Guides         []GraphGuide                `json:"guides,omitempty"`                     // [] len >= 0
	MetricClusters []GraphMetricCluster        `json:"metric_clusters,omitempty"`            // [] len >= 0
}

// NewGraph returns a Graph (with defaults, if applicable)
func NewGraph() *Graph {
	return &Graph{}
}

// FetchGraph retrieves graph with passed cid.
func (a *API) FetchGraph(cid CIDType) (*Graph, error) {
	if cid == nil || *cid == "" {
		return nil, errors.New("invalid graph CID (none)")
	}

	var graphCID string
	if !strings.HasPrefix(*cid, config.GraphPrefix) {
		graphCID = fmt.Sprintf("%s/%s", config.GraphPrefix, *cid)
	} else {
		graphCID = *cid
	}

	matched, err := regexp.MatchString(config.GraphCIDRegex, graphCID)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, errors.Errorf("invalid graph CID (%s)", graphCID)
	}

	result, err := a.Get(graphCID)
	if err != nil {
		return nil, errors.Wrap(err, "fetching graph")
	}
	if a.Debug {
		a.Log.Printf("fetch graph, received JSON: %s", string(result))
	}

	graph := new(Graph)
	if err := json.Unmarshal(result, graph); err != nil {
		return nil, errors.Wrap(err, "parsing graph")
	}

	return graph, nil
}

// FetchGraphs retrieves all graphs available to the API Token.
func (a *API) FetchGraphs() (*[]Graph, error) {
	result, err := a.Get(config.GraphPrefix)
	if err != nil {
		return nil, errors.Wrap(err, "fetching graphs")
	}

	var graphs []Graph
	if err := json.Unmarshal(result, &graphs); err != nil {
		return nil, errors.Wrap(err, "parsing graphs")
	}

	return &graphs, nil
}

// UpdateGraph updates passed graph.
func (a *API) UpdateGraph(cfg *Graph) (*Graph, error) {
	if cfg == nil {
		return nil, errors.New("invalid graph config (nil)")
	}

	graphCID := cfg.CID

	matched, err := regexp.MatchString(config.GraphCIDRegex, graphCID)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, errors.Errorf("invalid graph CID (%s)", graphCID)
	}

	jsonCfg, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	if a.Debug {
		a.Log.Printf("update graph, sending JSON: %s", string(jsonCfg))
	}

	result, err := a.Put(graphCID, jsonCfg)
	if err != nil {
		return nil, errors.Wrap(err, "updating graph")
	}

	graph := &Graph{}
	if err := json.Unmarshal(result, graph); err != nil {
		return nil, errors.Wrap(err, "parsing graph")
	}

	return graph, nil
}

// CreateGraph creates a new graph.
func (a *API) CreateGraph(cfg *Graph) (*Graph, error) {
	if cfg == nil {
		return nil, errors.New("invalid graph config (nil)")
	}

	jsonCfg, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	if a.Debug {
		a.Log.Printf("update graph, sending JSON: %s", string(jsonCfg))
	}

	result, err := a.Post(config.GraphPrefix, jsonCfg)
	if err != nil {
		return nil, errors.Wrap(err, "creating graph")
	}

	graph := &Graph{}
	if err := json.Unmarshal(result, graph); err != nil {
		return nil, errors.Wrap(err, "parsing graph")
	}

	return graph, nil
}

// DeleteGraph deletes passed graph.
func (a *API) DeleteGraph(cfg *Graph) (bool, error) {
	if cfg == nil {
		return false, errors.New("invalid graph config (nil)")
	}
	return a.DeleteGraphByCID(CIDType(&cfg.CID))
}

// DeleteGraphByCID deletes graph with passed cid.
func (a *API) DeleteGraphByCID(cid CIDType) (bool, error) {
	if cid == nil || *cid == "" {
		return false, errors.New("invalid graph CID (none)")
	}

	var graphCID string
	if !strings.HasPrefix(*cid, config.GraphPrefix) {
		graphCID = fmt.Sprintf("%s/%s", config.GraphPrefix, *cid)
	} else {
		graphCID = *cid
	}

	matched, err := regexp.MatchString(config.GraphCIDRegex, graphCID)
	if err != nil {
		return false, err
	}
	if !matched {
		return false, errors.Errorf("invalid graph CID (%s)", graphCID)
	}

	_, err = a.Delete(graphCID)
	if err != nil {
		return false, errors.Wrap(err, "deleting graph")
	}

	return true, nil
}

// SearchGraphs returns graphs matching the specified search query
// and/or filter. If nil is passed for both parameters all graphs
// will be returned.
func (a *API) SearchGraphs(searchCriteria *SearchQueryType, filterCriteria *SearchFilterType) (*[]Graph, error) {
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
		return a.FetchGraphs()
	}

	reqURL := url.URL{
		Path:     config.GraphPrefix,
		RawQuery: q.Encode(),
	}

	result, err := a.Get(reqURL.String())
	if err != nil {
		return nil, errors.Wrap(err, "searching graphs")
	}

	var graphs []Graph
	if err := json.Unmarshal(result, &graphs); err != nil {
		return nil, errors.Wrap(err, "parsing graphs")
	}

	return &graphs, nil
}

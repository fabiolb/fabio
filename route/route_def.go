package route

type Cmd string

const (
	RouteAddCmd    Cmd = "route add"
	RouteDelCmd    Cmd = "route del"
	RouteWeightCmd Cmd = "route weight"
)

type RouteDef struct {
	Opts    map[string]string `json:"opts,omitempty"`
	Cmd     Cmd               `json:"cmd"`
	Service string            `json:"service"`
	Src     string            `json:"src"`
	Dst     string            `json:"dst"`
	Tags    []string          `json:"tags,omitempty"`
	Weight  float64           `json:"weight"`
}

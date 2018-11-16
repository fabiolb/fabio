package route

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	reRouteAdd    = regexp.MustCompile(`^route\s+add`)
	reRouteDel    = regexp.MustCompile(`^route\s+del`)
	reRouteWeight = regexp.MustCompile(`^route\s+weight`)
	reComment     = regexp.MustCompile(`^(#|//)`)
	reBlankLine   = regexp.MustCompile(`^\s*$`)
)

const Commands = `
Route commands can have the following form:

route add <svc> <src> <dst>[ weight <w>][ tags "<t1>,<t2>,..."][ opts "k1=v1 k2=v2 ..."]
  - Add route for service svc from src to dst with optional weight, tags and options.
    Valid options are:

	  strip=/path        : forward '/path/to/file' as '/to/file'
	  proto=tcp          : upstream service is TCP, dst is ':port'
	  proto=https        : upstream service is HTTPS
	  tlsskipverify=true : disable TLS cert validation for HTTPS upstream
	  host=name          : set the Host header to 'name'. If 'name == "dst"' then the 'Host' header will be set to the registered upstream host name
	  register=name      : register fabio as new service 'name'. Useful for registering hostnames for host specific routes.
      auth=name          : name of the auth scheme to use (defined in proxy.auth)

route del <svc>[ <src>[ <dst>]]
  - Remove route matching svc, src and/or dst

route del <svc> tags "<t1>,<t2>,..."
  - Remove all routes of service matching svc and tags

route del tags "<t1>,<t2>,..."
  - Remove all routes matching tags

route weight <svc> <src> weight <w> tags "<t1>,<t2>,..."
  - Route w% of traffic to all services matching svc, src and tags

route weight <src> weight <w> tags "<t1>,<t2>,..."
  - Route w% of traffic to all services matching src and tags

route weight <svc> <src> weight <w>
  - Route w% of traffic to all services matching svc and src

route weight service host/path weight w tags "tag1,tag2"
  - Route w% of traffic to all services matching service, host/path and tags

    w is a float > 0 describing a percentage, e.g. 0.5 == 50%
    w <= 0: means no fixed weighting. Traffic is evenly distributed
    w > 0: route will receive n% of traffic. If sum(w) > 1 then w is normalized.
    sum(w) >= 1: only matching services will receive traffic

   Note that the total sum of traffic sent to all matching routes is w%.
`

// Parse loads a routing table from a set of route commands.
//
// The commands are parsed in order and order matters.
// Deleting a route that has not been created yet yields
// a different result than the other way around.
func Parse(in string) (defs []*RouteDef, err error) {
	var def *RouteDef
	for i, s := range strings.Split(in, "\n") {
		def, err = nil, nil
		s = strings.TrimSpace(s)
		switch {
		case reComment.MatchString(s) || reBlankLine.MatchString(s):
			continue
		case reRouteAdd.MatchString(s):
			def, err = parseRouteAdd(s)
		case reRouteDel.MatchString(s):
			def, err = parseRouteDel(s)
		case reRouteWeight.MatchString(s):
			def, err = parseRouteWeight(s)
		default:
			err = errors.New("syntax error: 'route' expected")
		}
		if err != nil {
			return nil, fmt.Errorf("line %d: %s", i+1, err)
		}
		defs = append(defs, def)
	}
	return defs, nil
}

// ParseAliases scans a set of route commands for the "register" option and
// returns a list of services which should be registered by the backend.
func ParseAliases(in string) (names []string, err error) {
	var defs []*RouteDef
	var def *RouteDef
	for i, s := range strings.Split(in, "\n") {
		def, err = nil, nil
		s = strings.TrimSpace(s)
		switch {
		case reComment.MatchString(s) || reBlankLine.MatchString(s):
			continue
		case reRouteAdd.MatchString(s):
			def, err = parseRouteAdd(s)
		case reRouteDel.MatchString(s):
			def, err = parseRouteDel(s)
		case reRouteWeight.MatchString(s):
			def, err = parseRouteWeight(s)
		default:
			err = errors.New("syntax error: 'route' expected")
		}
		if err != nil {
			return nil, fmt.Errorf("line %d: %s", i+1, err)
		}
		defs = append(defs, def)
	}

	var aliases []string

	for _, d := range defs {
		registerName, ok := d.Opts["register"]
		if ok {
			aliases = append(aliases, registerName)
		}
	}
	return aliases, nil
}

// route add <svc> <src> <dst>[ weight <w>][ tags "<t1>,<t2>,..."][ opts "k=v k=v ..."]
// 1: service 2: src 3: dst 4: weight expr 5: weight val 6: tags expr 7: tags val 8: opts expr 9: opts val
var reAdd = mustCompileWithFlexibleSpace(`^route add (\S+) (\S+) (\S+)( weight (\S+))?( tags "([^"]*)")?( opts "([^"]*)")?$`)

func parseRouteAdd(s string) (*RouteDef, error) {
	if m := reAdd.FindStringSubmatch(s); m != nil {
		w, err := parseWeight(m[5])
		return &RouteDef{
			Cmd:     RouteAddCmd,
			Service: m[1],
			Src:     m[2],
			Dst:     m[3],
			Weight:  w,
			Tags:    parseTags(m[7]),
			Opts:    parseOpts(m[9]),
		}, err
	}
	return nil, errors.New("syntax error: 'route add' invalid")
}

// route del <svc>[ <src>][ <dst>]
// 1: service 2: src expr 3: src 4: dst expr 5: dst
var reDel = mustCompileWithFlexibleSpace(`^route del (\S+)( (\S+)( (\S+))?)?$`)

// route del <svc> tags "<t1>,<t2>,..."
// 1: service 2: tags
var reDelSvcTags = mustCompileWithFlexibleSpace(`^route del (\S+) tags "([^"]*)"$`)

// route del tags "<t1>,<t2>,..."
// 2: tags
var reDelTags = mustCompileWithFlexibleSpace(`^route del tags "([^"]*)"$`)

func parseRouteDel(s string) (*RouteDef, error) {
	if m := reDelSvcTags.FindStringSubmatch(s); m != nil {
		return &RouteDef{Cmd: RouteDelCmd, Service: m[1], Tags: parseTags(m[2])}, nil
	}
	if m := reDelTags.FindStringSubmatch(s); m != nil {
		return &RouteDef{Cmd: RouteDelCmd, Tags: parseTags(m[1])}, nil
	}
	if m := reDel.FindStringSubmatch(s); m != nil {
		return &RouteDef{Cmd: RouteDelCmd, Service: m[1], Src: m[3], Dst: m[5]}, nil
	}
	return nil, errors.New("syntax error: 'route del' invalid")
}

// route weight <svc> <src> weight <w>[ tags "<t1>,<t2>,..."]
// 1: service 2: src 3: weight val 4: tags expr 5: tags val
var reWeightSvc = mustCompileWithFlexibleSpace(`^route weight (\S+) (\S+) weight (\S+)( tags "([^"]*)")?$`)

// route weight <src> weight <w> tags "<t1>,<t2>,..."
// 1: src 2: weight val 3: tags val
var reWeightSrc = mustCompileWithFlexibleSpace(`^route weight (\S+) weight (\S+) tags "([^"]*)"$`)

func parseRouteWeight(s string) (*RouteDef, error) {
	if m := reWeightSvc.FindStringSubmatch(s); m != nil {
		w, err := parseWeight(m[3])
		return &RouteDef{
			Cmd:     RouteWeightCmd,
			Service: m[1],
			Src:     m[2],
			Weight:  w,
			Tags:    parseTags(m[5]),
		}, err
	}
	if m := reWeightSrc.FindStringSubmatch(s); m != nil {
		w, err := parseWeight(m[2])
		return &RouteDef{
			Cmd:    RouteWeightCmd,
			Src:    m[1],
			Weight: w,
			Tags:   parseTags(m[3]),
		}, err
	}
	return nil, errors.New("syntax error: 'route weight' invalid")
}

func mustCompileWithFlexibleSpace(re string) *regexp.Regexp {
	return regexp.MustCompile(strings.Replace(re, " ", "\\s+", -1))
}

func parseWeight(s string) (float64, error) {
	if s == "" {
		return 0, nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0, errors.New("syntax error: weight value invalid")
	}
	return f, nil
}

func parseTags(s string) []string {
	if s == "" {
		return nil
	}
	tags := strings.Split(s, ",")
	for i, t := range tags {
		tags[i] = strings.TrimSpace(t)
	}
	return tags
}

func parseOpts(s string) map[string]string {
	if s == "" {
		return nil
	}
	m := make(map[string]string)
	for _, f := range strings.Fields(s) {
		p := strings.SplitN(f, "=", 2)
		if len(p) == 1 {
			m[f] = ""
		} else {
			m[p[0]] = p[1]
		}
	}
	return m
}

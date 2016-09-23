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

func ParseNew(in string) (defs []*RouteDef, err error) {
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

// route add <svc> <src> <dst> weight <w> tags "<t1>,<t2>,..."
// route add <svc> <src> <dst> weight <w>
// route add <svc> <src> <dst> tags "<t1>,<t2>,..."
// route add <svc> <src> <dst>
func parseRouteAdd(s string) (d *RouteDef, err error) {
	// 1: service 2: src 3: dst 4: weight expr 5: weight val 6: tags expr 7: tags val 8: opts expr 9: opts val
	re := mustCompileWithFlexibleSpace(`^route add (\S+) (\S+) (\S+)( weight (\S+))?( tags "([^"]*)")?( opts "([^"]*)")?$`)

	m := re.FindStringSubmatch(s)
	if m == nil {
		return nil, errors.New("syntax error: 'route add' invalid")
	}

	d = new(RouteDef)
	d.Cmd = RouteAddCmd
	d.Service = m[1]
	d.Src = m[2]
	d.Dst = m[3]
	d.Weight, err = parseWeight(m[5])
	d.Tags = parseTags(m[7])
	d.Opts = parseOpts(m[9])

	return
}

// route del <svc> <src> <dst>
// route del <svc> <src>
// route del <svc>
func parseRouteDel(s string) (d *RouteDef, err error) {
	// 1: service 2: src expr 3: src 4: dst expr 5: dst
	re := mustCompileWithFlexibleSpace(`^route del (\S+)( (\S+)( (\S+))?)?$`)

	m := re.FindStringSubmatch(s)
	if m == nil {
		return nil, errors.New("syntax error: 'route del' invalid")
	}

	d = new(RouteDef)
	d.Cmd = RouteDelCmd
	d.Service = m[1]
	d.Src = m[3]
	d.Dst = m[5]

	return
}

// route weight <svc> <src> weight <w> tags "<t1>,<t2>,..."'
// route weight <svc> <src> weight <w>'
// route weight <src> weight <w> tags "<t1>,<t2>,..."'
func parseRouteWeight(s string) (d *RouteDef, err error) {
	// 1: service 2: src 3: weight val 4: tags expr 5: tags val
	reSvc := mustCompileWithFlexibleSpace(`^route weight (\S+) (\S+) weight (\S+)( tags "([^"]*)")?$`)
	// 1: src 2: weight val 3: tags val
	reSrc := mustCompileWithFlexibleSpace(`^route weight (\S+) weight (\S+) tags "([^"]*)"$`)

	d = new(RouteDef)
	if m := reSvc.FindStringSubmatch(s); m != nil {
		d.Cmd = RouteWeightCmd
		d.Service = m[1]
		d.Src = m[2]
		d.Weight, err = parseWeight(m[3])
		d.Tags = parseTags(m[5])
		return d, nil
	}
	if m := reSrc.FindStringSubmatch(s); m != nil {
		d.Cmd = RouteWeightCmd
		d.Src = m[1]
		d.Weight, err = parseWeight(m[2])
		d.Tags = parseTags(m[3])
		return d, nil
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

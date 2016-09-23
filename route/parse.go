package route

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// Parse loads a routing table from a set of route commands.
//
// The commands are parsed in order and order matters.
// Deleting a route that has not been created yet yields
// a different result than the other way around.
//
// Route commands can have the following form:
//
// route add <svc> <src> <dst> weight <w> tags "<t1>,<t2>,..."
//   - Add route for service svc from src to dst and assign weight and tags
//
// route add <svc> <src> <dst> weight <w>
//   - Add route for service svc from src to dst and assign weight
//
// route add <svc> <src> <dst> tags "<t1>,<t2>,..."
//   - Add route for service svc from src to dst and assign tags
//
// route add <svc> <src> <dst>
//   - Add route for service svc from src to dst
//
// route del <svc> <src> <dst>
//   - Remove route matching svc, src and dst
//
// route del <svc> <src>
//   - Remove all routes of services matching svc and src
//
// route del <svc>
//   - Remove all routes of service matching svc
//
// route weight <svc> <src> weight <w> tags "<t1>,<t2>,..."
//   - Route w% of traffic to all services matching svc, src and tags
//
// route weight <src> weight <w> tags "<t1>,<t2>,..."
//   - Route w% of traffic to all services matching src and tags
//
// route weight <svc> <src> weight <w>
//   - Route w% of traffic to all services matching svc and src
//
// route weight service host/path weight w tags "tag1,tag2"
//   - Route w% of traffic to all services matching service, host/path and tags
//
//     w is a float > 0 describing a percentage, e.g. 0.5 == 50%
//     w <= 0: means no fixed weighting. Traffic is evenly distributed
//     w > 0: route will receive n% of traffic. If sum(w) > 1 then w is normalized.
//     sum(w) >= 1: only matching services will receive traffic
//
//    Note that the total sum of traffic sent to all matching routes is w%.
//
func ParseOldNoOpts(s string) ([]*RouteDef, error) {
	p := &parser{}
	if err := p.parse(strings.NewReader(s)); err != nil {
		return nil, err
	}
	return p.defs, nil
}

type parser struct {
	lineNumber int
	line       string
	defs       []*RouteDef
}

type cmdFn func(s string) error

func (p *parser) parse(r io.Reader) error {
	cmds := map[string]cmdFn{
		`^route\s+add `:    p.routeAdd,
		`^route\s+del `:    p.routeDel,
		`^route\s+weight `: p.routeWeight,
	}

	sc := bufio.NewScanner(r)
	for sc.Scan() {
		p.lineNumber++
		p.line = strings.TrimSpace(sc.Text())
		if p.line == "" || strings.HasPrefix(p.line, "#") {
			continue
		}
		var fn cmdFn
		for cmd := range cmds {
			re := regexp.MustCompile(cmd)
			if re.MatchString(p.line) {
				fn = cmds[cmd]
				break
			}
		}
		if fn == nil {
			return p.syntaxError()
		}
		if err := fn(p.line); err != nil {
			return err
		}
	}
	return nil
}

var (
	// route add <svc> <src> <dst> weight <w> tags "<t1>,<t2>,..."
	routeAddSvcWeightTags = regexp.MustCompile(`^route\s+add\s+(\S+)\s+(\S+)\s+(\S+)\s+weight\s+(\S+)\s+tags\s+"([^"]*)"$`)

	// route add <svc> <src> <dst> weight <w>
	routeAddSvcWeight = regexp.MustCompile(`^route\s+add\s+(\S+)\s+(\S+)\s+(\S+)\s+weight\s+(\S+)$`)

	// route add <svc> <src> <dst> tags "<t1>,<t2>,..."
	routeAddSvcTags = regexp.MustCompile(`^route\s+add\s+(\S+)\s+(\S+)\s+(\S+)\s+tags\s+"([^"]*)"$`)

	// route add <svc> <src> <dst>
	routeAddSvc = regexp.MustCompile(`^route\s+add\s+(\S+)\s+(\S+)\s+(\S+)$`)
)

func (p *parser) routeAdd(s string) error {
	var svc, src, dst string
	var tags []string
	var w float64
	var err error

	// test most to least specific
	if m := routeAddSvcWeightTags.FindStringSubmatch(s); m != nil {
		svc, src, dst, tags = m[1], m[2], m[3], parseTags(m[5])
		w, err = parseWeight(m[4])
	} else if m := routeAddSvcWeight.FindStringSubmatch(s); m != nil {
		svc, src, dst = m[1], m[2], m[3]
		w, err = parseWeight(m[4])
	} else if m := routeAddSvcTags.FindStringSubmatch(s); m != nil {
		svc, src, dst, tags = m[1], m[2], m[3], parseTags(m[4])
	} else if m := routeAddSvc.FindStringSubmatch(s); m != nil {
		svc, src, dst = m[1], m[2], m[3]
	} else {
		err = p.syntaxError()
	}

	if err != nil {
		return err
	}

	p.defs = append(p.defs, &RouteDef{Cmd: RouteAddCmd, Service: svc, Src: src, Dst: dst, Weight: w, Tags: tags})
	//p.t.AddRoute(svc, src, dst, w, tags)
	return nil
}

var (
	// route del <svc> <src> <dst>
	routeDelSvcSrcDst = regexp.MustCompile(`^route\s+del\s+(\S+)\s+(\S+)\s+(\S+)$`)

	// route del <svc> <src>
	routeDelSvcSrc = regexp.MustCompile(`^route\s+del\s+(\S+)\s+(\S+)$`)

	// route del <svc>
	routeDelSvc = regexp.MustCompile(`^route\s+del\s+(\S+)$`)
)

func (p *parser) routeDel(s string) error {
	var svc, src, dst string
	var err error

	// test most to least specific
	if m := routeDelSvcSrcDst.FindStringSubmatch(s); m != nil {
		svc, src, dst = m[1], m[2], m[3]
	} else if m := routeDelSvcSrc.FindStringSubmatch(s); m != nil {
		svc, src = m[1], m[2]
	} else if m := routeDelSvc.FindStringSubmatch(s); m != nil {
		svc = m[1]
	} else {
		err = p.syntaxError()
	}

	if err != nil {
		return err
	}

	p.defs = append(p.defs, &RouteDef{Cmd: RouteDelCmd, Service: svc, Src: src, Dst: dst})
	//p.t.DelRoute(svc, src, dst)
	return nil
}

var (
	// route weight <svc> <src> weight <w> tags "<t1>,<t2>,..."'
	routeWeightSvcSrcTags = regexp.MustCompile(`^route\s+weight\s+(\S+)\s+(\S+)\s+weight\s+(\S+)\s+tags\s+"([^"]*)"$`)

	// route weight <src> weight <w> tags "<t1>,<t2>,..."'
	routeWeightSrcTags = regexp.MustCompile(`^route\s+weight\s+(\S+)\s+weight\s+(\S+)\s+tags\s+"([^"]+)"$`)

	// route weight <svc> <src> weight <w>'
	routeWeightSvcSrc = regexp.MustCompile(`^route\s+weight\s+(\S+)\s+(\S+)\s+weight\s+(\S+)$`)
)

func (p *parser) routeWeight(s string) error {
	var svc, src string
	var tags []string
	var w float64
	var err error

	// test most to least specific
	if m := routeWeightSvcSrcTags.FindStringSubmatch(s); m != nil {
		svc, src, tags = m[1], m[2], parseTags(m[4])
		w, err = parseWeight(m[3])
	} else if m := routeWeightSvcSrc.FindStringSubmatch(s); m != nil {
		svc, src = m[1], m[2]
		w, err = parseWeight(m[3])
	} else if m := routeWeightSrcTags.FindStringSubmatch(s); m != nil {
		src, tags = m[1], parseTags(m[3])
		w, err = parseWeight(m[2])
	} else {
		err = p.syntaxError()
	}

	if err != nil {
		return err
	}

	p.defs = append(p.defs, &RouteDef{Cmd: RouteWeightCmd, Service: svc, Src: src, Weight: w, Tags: tags})
	//p.t.AddRouteWeight(svc, src, w, tags)
	return nil
}

func (p *parser) syntaxError() error {
	return fmt.Errorf("route: line %d: syntax error in %s", p.lineNumber, p.line)
}

func (p *parser) errorf(msg string, args ...string) error {
	return fmt.Errorf("route: line %d: %s", p.lineNumber, fmt.Sprintf(msg, args))
}

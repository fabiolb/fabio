package route

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
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
func Parse(r io.Reader) (Table, error) {
	p := &parser{t: make(Table)}
	if err := p.parse(r); err != nil {
		return nil, err
	}
	return p.t, nil
}

// ParseFile loads a routing table from a file.
func ParseFile(path string) (Table, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(f)
}

// ParseString loads a routing table from a string.
func ParseString(s string) (Table, error) {
	return Parse(strings.NewReader(s))
}

type parser struct {
	t          Table
	lineNumber int
	line       string
}

type cmdFn func(s string) error

func (p *parser) parse(r io.Reader) error {
	cmds := map[string]cmdFn{
		"route add ":    p.routeAdd,
		"route del ":    p.routeDel,
		"route weight ": p.routeWeight,
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
			if strings.HasPrefix(p.line, cmd) {
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
	routeAddSvcWeightTags = regexp.MustCompile(`^route add (\S+) (\S+) (\S+) weight (\S+) tags "([^"]*)"$`)

	// route add <svc> <src> <dst> weight <w>
	routeAddSvcWeight = regexp.MustCompile(`^route add (\S+) (\S+) (\S+) weight (\S+)$`)

	// route add <svc> <src> <dst> tags "<t1>,<t2>,..."
	routeAddSvcTags = regexp.MustCompile(`^route add (\S+) (\S+) (\S+) tags "([^"]*)"$`)

	// route add <svc> <src> <dst>
	routeAddSvc = regexp.MustCompile(`^route add (\S+) (\S+) (\S+)$`)
)

func (p *parser) routeAdd(s string) error {
	var svc, src, dst string
	var tags []string
	var w float64
	var err error

	// test most to least specific
	if m := routeAddSvcWeightTags.FindStringSubmatch(s); m != nil {
		svc, src, dst, tags = m[1], m[2], m[4], strings.FieldsFunc(m[5], splitSeparator)
		w, err = p.parseWeight(m[3])
	} else if m := routeAddSvcWeight.FindStringSubmatch(s); m != nil {
		svc, src, dst = m[1], m[2], m[3]
		w, err = p.parseWeight(m[4])
	} else if m := routeAddSvcTags.FindStringSubmatch(s); m != nil {
		svc, src, dst, tags = m[1], m[2], m[3], strings.FieldsFunc(m[4], splitSeparator)
	} else if m := routeAddSvc.FindStringSubmatch(s); m != nil {
		svc, src, dst = m[1], m[2], m[3]
	} else {
		err = p.syntaxError()
	}

	if err != nil {
		return err
	}
	p.t.AddRoute(svc, src, dst, w, tags)
	return nil
}

func splitSeparator(c rune) bool {
	return c == ','
}

var (
	// route del <svc> <src> <dst>
	routeDelSvcSrcDst = regexp.MustCompile(`^route del (\S+) (\S+) (\S+)$`)

	// route del <svc> <src>
	routeDelSvcSrc = regexp.MustCompile(`^route del (\S+) (\S+)$`)

	// route del <svc>
	routeDelSvc = regexp.MustCompile(`^route del (\S+)$`)
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

	p.t.DelRoute(svc, src, dst)
	return nil
}

var (
	// route weight <svc> <src> weight <w> tags "<t1>,<t2>,..."'
	routeWeightSvcSrcTags = regexp.MustCompile(`^route weight (\S+) (\S+) weight (\S+) tags "([^"]*)"$`)

	// route weight <src> weight <w> tags "<t1>,<t2>,..."'
	routeWeightSrcTags = regexp.MustCompile(`^route weight (\S+) weight (\S+) tags "([^"]+)"$`)

	// route weight <svc> <src> weight <w>'
	routeWeightSvcSrc = regexp.MustCompile(`^route weight (\S+) (\S+) weight (\S+)$`)
)

func (p *parser) routeWeight(s string) error {
	var svc, src string
	var tags []string
	var w float64
	var err error

	// test most to least specific
	if m := routeWeightSvcSrcTags.FindStringSubmatch(s); m != nil {
		svc, src, tags = m[1], m[2], strings.FieldsFunc(m[4], splitSeparator)
		w, err = p.parseWeight(m[3])
	} else if m := routeWeightSvcSrc.FindStringSubmatch(s); m != nil {
		svc, src = m[1], m[2]
		w, err = p.parseWeight(m[3])
	} else if m := routeWeightSrcTags.FindStringSubmatch(s); m != nil {
		src, tags = m[1], strings.FieldsFunc(m[3], splitSeparator)
		w, err = p.parseWeight(m[2])
	} else {
		err = p.syntaxError()
	}

	if err != nil {
		return err
	}

	p.t.AddRouteWeight(svc, src, w, tags)
	return nil
}

func (p *parser) parseWeight(s string) (float64, error) {
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, p.errorf("invalid weight: %s", s)
	}
	return n, nil
}

func (p *parser) syntaxError() error {
	return fmt.Errorf("route: line %d: syntax error in %s", p.lineNumber, p.line)
}

func (p *parser) errorf(msg string, args ...string) error {
	return fmt.Errorf("route: line %d: %s", p.lineNumber, fmt.Sprintf(msg, args))
}

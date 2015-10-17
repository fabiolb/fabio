package route

import (
	"bufio"
	"fmt"
	"io"
	"os"
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
//   route add service host/path targetURL
//   - Add a new route for host/path to targetURL
//
//   route del service
//   - Remove all routes for service
//
//   route del service host/path
//   - Remove all routes for host/path for this service only
//
//   route del service host/path targetURL
//    - Remove only this route
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

type cmdMap map[string]func(args []string) error

func (p *parser) parse(r io.Reader) error {
	cmds := cmdMap{"route": p.route}

	sc := bufio.NewScanner(r)
	for sc.Scan() {
		p.lineNumber++
		p.line = strings.TrimSpace(sc.Text())
		if p.line == "" || strings.HasPrefix(p.line, "#") {
			continue
		}
		if err := p.call(cmds, strings.Split(p.line, " ")); err != nil {
			return err
		}
	}
	return nil
}

func (p *parser) call(cmds cmdMap, args []string) error {
	cmd, args := args[0], args[1:]
	fn, ok := cmds[cmd]
	if !ok {
		return p.syntaxError()
	}
	if err := fn(args); err != nil {
		return err
	}
	return nil
}

func (p *parser) syntaxError() error {
	return fmt.Errorf("route: line %d: syntax error in %s", p.lineNumber, p.line)
}

func (p *parser) errorf(msg string, args ...string) error {
	return fmt.Errorf("route: line %d: %s", p.lineNumber, fmt.Sprintf(msg, args))
}

// route implements the 'route' command.
func (p *parser) route(args []string) error {
	cmds := cmdMap{
		"add":    p.routeAdd,
		"del":    p.routeDel,
		"weight": p.routeWeight,
	}
	return p.call(cmds, args)
}

// routeAdd implements 'route add <svc> <prefix> <target> [weight <weight>] [tags "tag1,tag2,..."]'
func (p *parser) routeAdd(args []string) error {
	var service, prefix, target string
	var weight float64
	var tags []string
	var err error

	switch len(args) {
	case 3:
		service, prefix, target = args[0], args[1], args[2]

	case 5:
		service, prefix, target = args[0], args[1], args[2]
		switch args[3] {
		case "weight":
			if weight, err = p.parseWeight(args[3:]); err != nil {
				return err
			}
		case "tags":
			if tags, err = p.parseTags(args[3:]); err != nil {
				return err
			}
		default:
			return p.syntaxError()
		}

	case 7:
		service, prefix, target = args[0], args[1], args[2]
		if weight, err = p.parseWeight(args[3:]); err != nil {
			return err
		}
		if tags, err = p.parseTags(args[5:]); err != nil {
			return err
		}

	default:
		return p.syntaxError()
	}

	p.t.AddRoute(service, prefix, target, weight, tags)
	return nil
}

// routeDel implements 'route del service [prefix [target]]''
func (p *parser) routeDel(args []string) error {
	var service, prefix, target string
	switch len(args) {
	case 1:
		service = args[0]
	case 2:
		service, prefix = args[0], args[1]
	case 3:
		service, prefix, target = args[0], args[1], args[2]
	default:
		return p.syntaxError()
	}

	p.t.DelRoute(service, prefix, target)
	return nil
}

// routeWeight implements 'route weight <svc> <prefix> weight <weight> tags "tag1,tag2,..."'
func (p *parser) routeWeight(args []string) error {
	var service, prefix string
	var weight float64
	var tags []string
	var err error

	switch len(args) {
	case 6:
		service, prefix = args[0], args[1]
		if weight, err = p.parseWeight(args[2:]); err != nil {
			return err
		}
		if tags, err = p.parseTags(args[4:]); err != nil {
			return err
		}

	default:
		p.syntaxError()
	}

	p.t.AddRouteWeight(service, prefix, weight, tags)
	return nil
}

func (p *parser) parseWeight(args []string) (float64, error) {
	if args[0] != "weight" || len(args) < 2 {
		return 0, p.syntaxError()
	}
	n, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return 0, p.errorf("invalid weight: %s", args[1])
	}
	return n, nil
}

func (p *parser) parseTags(args []string) ([]string, error) {
	if args[0] != "tags" || len(args) < 2 {
		return nil, p.syntaxError()
	}
	tags := args[1]
	if !strings.HasPrefix(tags, `"`) || !strings.HasPrefix(tags, `"`) {
		return nil, p.syntaxError()
	}
	return strings.Split(tags[1:len(tags)-1], ","), nil
}

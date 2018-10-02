package route

import (
	"github.com/gobwas/glob"
	"sync"

	"fmt"
	"github.com/gobwas/glob/match"
)

// GlobCache implements an LRU cache for compiled glob patterns.
type GlobCache struct {
	// m maps patterns to compiled glob matchers.
	m sync.Map

	// l contains the added patterns and serves as an LRU cache.
	// l has a fixed size and is initialized in the constructor.
	l []string

	// h is the first element in l.
	h int

	// n is the number of elements in l.
	n int
}

func NewGlobCache(size int) *GlobCache {
	return &GlobCache{
		l: make([]string, size),
	}
}

// Get returns the compiled glob pattern if it compiled without
// error. Otherwise, the function returns nil. If the pattern
// is not in the cache it will be added.
func (c *GlobCache) Get(pattern string) (glob.Glob, error) {
	// fast path
	if glb, ok := c.m.Load(pattern); ok {
		//Type Assert the returned interface{}
		switch t := glb.(type) {
		case match.Text:
			return glb.(match.Text), nil
		case match.Nothing:
			return glb.(match.Nothing), nil
		case match.Any:
			return glb.(match.Any), nil
		case match.AnyOf:
			return glb.(match.AnyOf), nil
		case match.BTree:
			return glb.(match.BTree), nil
		case match.Contains:
			return glb.(match.Contains), nil
		case match.EveryOf:
			return glb.(match.EveryOf), nil
		case match.List:
			return glb.(match.List), nil
		case match.Max:
			return glb.(match.Max), nil
		case match.Min:
			return glb.(match.Min), nil
		case match.Prefix:
			return glb.(match.Prefix), nil
		case match.PrefixSuffix:
			return glb.(match.PrefixSuffix), nil
		case match.PrefixAny:
			return glb.(match.PrefixAny), nil
		case match.Range:
			return glb.(match.Range), nil
		case match.Row:
			return glb.(match.Row), nil
		case match.Single:
			return glb.(match.Single), nil
		case match.Suffix:
			return glb.(match.Suffix), nil
		case match.SuffixAny:
			return glb.(match.SuffixAny), nil
		case match.Super:
			return glb.(match.Super), nil
		default:
			return nil, fmt.Errorf("[ERROR] Invalid type match in glob compare (%s)", t)
		}

	}

	// try to compile pattern
	glbCompiled, err := glob.Compile(pattern)
	if err != nil {
		return nil, err
	}

	// if the LRU buffer is not full just append
	// the element to the buffer.
	if c.n < len(c.l) {
		c.m.Store(pattern, glbCompiled)
		c.l[c.n] = pattern
		c.n++
		return glbCompiled, nil
	}

	// otherwise, remove the oldest element and move
	// the head. Note that once the buffer is full
	// (c.n == len(c.l)) it will never become smaller
	// again.
	// TODO add logging for cache full - How will this impact performance
	c.m.Delete(c.l[c.h])
	c.m.Store(pattern, glbCompiled)
	c.l[c.h] = pattern
	c.h = (c.h + 1) % c.n
	return glbCompiled, nil
}

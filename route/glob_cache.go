package route

import (
	"github.com/gobwas/glob"
	"sync"
)

// GlobCache implements an LRU cache for compiled glob patterns.
type GlobCache struct {
	mu sync.RWMutex

	// m maps patterns to compiled glob matchers.
	m map[string]glob.Glob

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
		m: make(map[string]glob.Glob, size),
		l: make([]string, size),
	}
}

// Get returns the compiled glob pattern if it compiled without
// error. Otherwise, the function returns nil. If the pattern
// is not in the cache it will be added.
func (c *GlobCache) Get(pattern string) (glob.Glob, error) {
	// fast path with read lock
	c.mu.RLock()
	g := c.m[pattern]
	c.mu.RUnlock()
	if g != nil {
		return g, nil
	}

	// slow path with write lock
	c.mu.Lock()
	defer c.mu.Unlock()

	// check again to handle race condition
	g = c.m[pattern]
	if g != nil {
		return g, nil
	}

	// try to compile pattern
	g, err := glob.Compile(pattern)
	if err != nil {
		return nil, err
	}

	// if the LRU buffer is not full just append
	// the element to the buffer.
	if c.n < len(c.l) {
		c.m[pattern] = g
		c.l[c.n] = pattern
		c.n++
		return g, nil
	}

	// otherwise, remove the oldest element and move
	// the head. Note that once the buffer is full
	// (c.n == len(c.l)) it will never become smaller
	// again.
	delete(c.m, c.l[c.h])
	c.m[pattern] = g
	c.l[c.h] = pattern
	c.h = (c.h + 1) % c.n
	return g, nil
}

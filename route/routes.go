package route

// routes stores a list of routes usually for a single host.
type routes []*route

// find returns the route with the given path and returns nil if none was found.
func (rt routes) find(path string) *route {
	for _, r := range rt {
		if r.path == path {
			return r
		}
	}
	return nil
}

// sort by path in reverse order (most to least specific)
func (rt routes) Len() int           { return len(rt) }
func (rt routes) Swap(i, j int)      { rt[i], rt[j] = rt[j], rt[i] }
func (rt routes) Less(i, j int) bool { return rt[j].path < rt[i].path }

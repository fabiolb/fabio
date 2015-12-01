package route

// Routes stores a list of routes usually for a single host.
type Routes []*Route

// find returns the route with the given path and returns nil if none was found.
func (rt Routes) find(path string) *Route {
	for _, r := range rt {
		if r.Path == path {
			return r
		}
	}
	return nil
}

// sort by path in reverse order (most to least specific)
func (rt Routes) Len() int           { return len(rt) }
func (rt Routes) Swap(i, j int)      { rt[i], rt[j] = rt[j], rt[i] }
func (rt Routes) Less(i, j int) bool { return rt[j].Path < rt[i].Path }

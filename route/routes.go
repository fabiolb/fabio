package route

import "slices"

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

func (rt Routes) findMethods(path string, methods []string) *Route {
	for _, r := range rt {
		if r.Path == path && slices.Equal(r.Methods, methods) {
			return r
		}
	}
	return nil
}

func (rt Routes) findAll(path string) Routes {
	var result Routes
	for _, r := range rt {
		if r.Path == path {
			result = append(result, r)
		}
	}
	return result
}

// sort by path in reverse order (most to least specific)
func (rt Routes) Len() int      { return len(rt) }
func (rt Routes) Swap(i, j int) { rt[i], rt[j] = rt[j], rt[i] }
func (rt Routes) Less(i, j int) bool {
	if rt[i].Path != rt[j].Path {
		return rt[i].Path > rt[j].Path
	}
	// More specific method sets precede wider ones. The stable sort
	// retains configuration order for equally specific sets.
	if len(rt[i].Methods) == 0 {
		return false
	}
	if len(rt[j].Methods) == 0 {
		return true
	}
	return len(rt[i].Methods) < len(rt[j].Methods)
}

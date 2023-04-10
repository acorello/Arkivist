package pathset

import "golang.org/x/exp/maps"

type PathSet map[string]struct{}

func New() PathSet {
	return PathSet{}
}

func (I PathSet) Add(a string) {
	I[a] = struct{}{}
}

func (I PathSet) Entries() []string {
	return maps.Keys(I)
}

func (I PathSet) IsEmpty() bool {
	return len(I) == 0
}

func (I PathSet) Contains(f string) bool {
	_, found := I[f]
	return found
}

func (I PathSet) Intersection(o PathSet) PathSet {
	if len(I) == 0 || len(o) == 0 {
		return PathSet{}
	}
	commonFiles := PathSet{}
	for f := range I {
		if o.Contains(f) {
			commonFiles.Add(f)
		}
	}
	return commonFiles
}

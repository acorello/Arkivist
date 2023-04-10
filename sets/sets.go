package sets

import "golang.org/x/exp/maps"

type Set[Element comparable] map[Element]struct{}

func New[Element comparable]() Set[Element] {
	return Set[Element]{}
}

func (I Set[E]) Add(a E) {
	I[a] = struct{}{}
}

func (I Set[E]) Entries() []E {
	return maps.Keys(I)
}

func (I Set[E]) IsEmpty() bool {
	return len(I) == 0
}

func (I Set[E]) Contains(f E) bool {
	_, found := I[f]
	return found
}

func (I Set[E]) Intersection(o Set[E]) Set[E] {
	commonFiles := New[E]()
	if len(I) == 0 || len(o) == 0 {
		return commonFiles
	}
	for f := range I {
		if o.Contains(f) {
			commonFiles.Add(f)
		}
	}
	return commonFiles
}

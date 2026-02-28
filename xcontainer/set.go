package xcontainer

import (
	"iter"
	"slices"
)

type Set[K comparable] map[K]struct{}

func ToSet[K comparable](items []K) Set[K] {
	set := make(Set[K], len(items))
	for _, item := range items {
		set.Add(item)
	}
	return set
}

func (set Set[K]) Add(values ...K) {
	for _, value := range values {
		set[value] = struct{}{}
	}
}

func (set Set[K]) Has(value K) bool {
	_, ok := set[value]
	return ok
}

func (set Set[K]) Remove(values ...K) {
	for _, value := range values {
		delete(set, value)
	}
}

func (set Set[K]) Union(sets ...Set[K]) Set[K] {
	result := make(Set[K])
	for _, other := range append(sets, set) {
		for item := range other.All() {
			result.Add(item)
		}
	}
	return result
}

func (set Set[K]) Intersection(sets ...Set[K]) Set[K] {
	result := make(Set[K])
	all := append(sets, set)
outer:
	for value := range set.Union(sets...) {
		for _, set := range all {
			if !set.Has(value) {
				continue outer
			}
		}
		result.Add(value)
	}
	return result
}

func (set Set[K]) All() iter.Seq[K] {
	return func(yield func(K) bool) {
		for value := range set {
			if !yield(value) {
				return
			}
		}
	}
}

func (set Set[K]) Collect() []K {
	return slices.Collect(set.All())
}

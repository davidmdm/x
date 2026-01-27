package xsync

import (
	"iter"
	"sync"
)

type Map[K comparable, V any] sync.Map

func (m *Map[K, V]) Load(key K) (V, bool) {
	result, ok := (*sync.Map)(m).Load(key)
	if !ok {
		var zero V
		return zero, ok
	}
	return result.(V), ok
}

func (m *Map[K, V]) Store(key K, value V) {
	(*sync.Map)(m).Store(key, value)
}

func (m *Map[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	result, ok := (*sync.Map)(m).LoadOrStore(key, value)
	return result.(V), ok
}

func (m *Map[K, V]) CompareAndSwap(key K, old, new V) (swapped bool) {
	return (*sync.Map)(m).CompareAndSwap(key, old, new)
}

func (m *Map[K, V]) CompareAndDelete(key K, old V) (deleted bool) {
	return (*sync.Map)(m).CompareAndDelete(key, old)
}

func (m *Map[K, V]) Clear() {
	(*sync.Map)(m).Clear()
}

func (m *Map[K, V]) Swap(key K, value V) (previous V, loaded bool) {
	result, loaded := (*sync.Map)(m).Swap(key, value)
	if !loaded {
		var zero V
		return zero, loaded
	}
	return result.(V), loaded
}

func (m *Map[K, V]) LoadAndDelete(key K) (value V, loaded bool) {
	result, ok := (*sync.Map)(m).LoadAndDelete(key)
	if !ok {
		var zero V
		return zero, ok
	}
	return result.(V), ok
}

func (m *Map[K, V]) Delete(key K) {
	(*sync.Map)(m).Delete(key)
}

func (m *Map[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		(*sync.Map)(m).Range(func(key any, value any) bool {
			return yield(key.(K), value.(V))
		})
	}
}

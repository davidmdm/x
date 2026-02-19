package xiter

import "iter"

func Join[T any](seqs ...iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, seq := range seqs {
			for value := range seq {
				if !yield(value) {
					return
				}
			}
		}
	}
}

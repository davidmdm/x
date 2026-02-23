package xcontainer

import "strings"

type PrefixTree[T any] struct {
	key      string
	children []PrefixTree[T]
	indices  []byte
	value    *T
}

func (node *PrefixTree[T]) Insert(key string, value T) {
	// By checking if the key is empty, we enable storing values on the empty root of the tree,
	// which is great for creating a default for any submatch, including lookup of empty strings.
	// Also, the key will be empty if we matched a child node exactly, which let's us reset the value
	// of that node.
	if key == "" {
		node.value = &value
		return
	}

	var child *PrefixTree[T]
	for i, char := range node.indices {
		if char == key[0] {
			child = &node.children[i]
			break
		}
	}

	// Since there's no child that matches the key, we can directly create the
	// child on this node using the entire key.
	if child == nil {
		node.append(PrefixTree[T]{key: key, value: &value})
		return
	}

	cp := commonPrefixLength(key, child.key)

	// the key contains the entire child node's key, so we can insert the value onto this child
	if cp == len(child.key) {
		child.Insert(key[cp:], value)
		return
	}

	// We create a node that represents the common prefix betwen the child and key
	commonNode := PrefixTree[T]{key: key[:cp]}

	// If the key is the prefix, then we need to give it a value, else we append the rest of the key, as a child.
	if len(key) == cp {
		commonNode.value = &value
	} else {
		commonNode.append(PrefixTree[T]{key: key[cp:], value: &value})
	}

	childValue := *child
	childValue.key = childValue.key[cp:]

	commonNode.append(childValue)

	*child = commonNode
}

func (node *PrefixTree[T]) append(child PrefixTree[T]) {
	node.children = append(node.children, child)
	node.indices = append(node.indices, child.key[0])
}

func (node PrefixTree[T]) GetLongestSubmatch(key string) (result T) {
Outer:
	for {
		var ok bool

		key, ok = strings.CutPrefix(key, node.key)
		if !ok {
			return
		}

		if node.value != nil {
			result = *node.value
		}

		if key == "" {
			return
		}

		for i, b := range node.indices {
			if b == key[0] {
				node = node.children[i]
				continue Outer
			}
		}

		return
	}
}

func commonPrefixLength(a, b string) (i int) {
	minimum := min(len(a), len(b))
	for i = 0; i < minimum; i++ {
		if a[i] != b[i] {
			break
		}
	}
	return
}

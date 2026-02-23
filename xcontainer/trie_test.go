package xcontainer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTrie(t *testing.T) {
	t.Run("get longest submatch", func(t *testing.T) {
		var trie PrefixTree[string]

		trie.Insert("start_of_the", "middle")
		trie.Insert("start_of_alt", "salt")
		trie.Insert("alt", "alt")
		trie.Insert("start", "start")
		trie.Insert("start_of_the_key", "end")

		require.Equal(t, "", trie.GetLongestSubmatch("st"))
		require.Equal(t, "start", trie.GetLongestSubmatch("start"))
		require.Equal(t, "start", trie.GetLongestSubmatch("start_of"))
		require.Equal(t, "middle", trie.GetLongestSubmatch("start_of_the"))
		require.Equal(t, "middle", trie.GetLongestSubmatch("start_of_the_ke"))
		require.Equal(t, "end", trie.GetLongestSubmatch("start_of_the_key"))
		require.Equal(t, "end", trie.GetLongestSubmatch("start_of_the_keys"))
		require.Equal(t, "", trie.GetLongestSubmatch("a"))
		require.Equal(t, "alt", trie.GetLongestSubmatch("alt"))
		require.Equal(t, "salt", trie.GetLongestSubmatch("start_of_alternate"))
	})

	t.Run("reset the key", func(t *testing.T) {
		var trie PrefixTree[string]
		trie.Insert("hello", "world")
		trie.Insert("hello", "universe")
		require.Equal(t, "universe", trie.GetLongestSubmatch("hello"))
	})

	t.Run("set value on root node", func(t *testing.T) {
		var trie PrefixTree[string]
		trie.Insert("", "default")
		trie.Insert("hello", "world")
		require.Equal(t, "default", trie.GetLongestSubmatch(""))
		require.Equal(t, "default", trie.GetLongestSubmatch("other"))
		require.Equal(t, "world", trie.GetLongestSubmatch("hello there!"))
	})
}

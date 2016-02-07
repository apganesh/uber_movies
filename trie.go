package main

import "strings"

// Node for Trie Prefix Tree
type TNode struct {
	ids    []int
	isword bool
	cnodes map[string]*TNode
}

type Trie struct {
	root *TNode
}

func NewTrie() *Trie {
	t := &Trie{}
	t.root = nil
	return t
}

type JsonMovieRes struct {
	Moviename string
}

// Struct to collect all unique movie ids
type IntSet struct {
	m map[int]bool
}

func NewIntSet() *IntSet {
	s := &IntSet{}
	s.m = make(map[int]bool)
	return s
}

func (s *IntSet) add(v int) {
	s.m[v] = true
}

func (s *IntSet) delete(v int) {
	delete(s.m, v)
}

func (s *IntSet) contains(v int) bool {
	_, b := s.m[v]
	return b
}

// Add a movie name which could have many strings
func (trie *Trie) addMovie(s string, id int) {
	s = strings.ToLower(s)

	res := strings.Fields(s)

	strings.ToLower(s)

	if trie.root == nil {
		tmp := new(TNode)
		tmp.isword = false
		tmp.cnodes = make(map[string]*TNode)
		trie.root = tmp
	}

	for i := range res {
		trie.root.addString(res[i], id)
	}
}

// Return the indices of the movie which match the pattern "s"
func (trie *Trie) getMovieIndices(s string) []int {

	var res = NewIntSet()
	s = strings.ToLower(s)

	words := strings.Fields(s)

	for i := range words {
		var n *TNode = trie.root.findNode(words[i])
		if n != nil {
			n.collectPrefix(res)
		}
	}
	var indices []int
	for i, _ := range res.m {
		indices = append(indices, i)
	}
	return indices
}

// TrieNode high level APIs

// return the TNode for a given prefix
func (tnode *TNode) findNode(s string) *TNode {
	cur := tnode
	if len(s) == 0 {
		return cur
	}
	for _, c := range s {
		if cur.cnodes[string(c)] == nil {
			return nil
		}
		cur = cur.cnodes[string(c)]
	}
	return cur
}

// Collect the indices which match the prefix (from the tnode)
func (tnode *TNode) collectPrefix(res *IntSet) {

	if tnode.isword == true {
		for _, v := range tnode.ids {
			res.add(v)
		}
	}

	for _, tn := range tnode.cnodes {
		tn.collectPrefix(res)
	}
}

// Add string to to the trie associated with id
func (tnode *TNode) addString(s string, id int) {
	cur := tnode
	if len(s) == 0 {
		return
	}
	for _, c := range s {
		if cur.cnodes[string(c)] == nil {
			tmp := new(TNode)
			tmp.isword = false
			tmp.cnodes = make(map[string]*TNode)
			cur.cnodes[string(c)] = tmp
		}
		cur = cur.cnodes[string(c)]
	}

	cur.isword = true
	cur.ids = append(cur.ids, id)
}

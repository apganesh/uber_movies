package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/context"
	"googlemaps.github.io/maps"
)

// Node for Trie Prefix Tree
type TNode struct {
	ids    []int
	isword bool
	cnodes map[string]*TNode
}

type LatLng struct {
	Lat float64
	Lng float64
}
type Location struct {
	Address string
	Funfact string
	Latlng  LatLng
}

type Locations []Location

type Record struct {
	Id       int
	Title    string
	Director string
	Released string
	Actors   []string
	//Funfacts []string
	Spots Locations
}

type Trie struct {
	root         *TNode
	movieid_map  map[string]int
	records      map[int]*Record
	geocache_map map[int]bool // Just to cache the records which have the geocoding done
}

var (
	gmapClient *maps.Client
)

func NewTrie() *Trie {
	t := &Trie{}
	t.root = nil
	t.movieid_map = make(map[string]int)
	t.records = make(map[int]*Record)
	//t.locationid_map = make(map[int]Locations)
	t.geocache_map = make(map[int]bool)
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

// Trie highlevel APIs
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

func (trie *Trie) getMovies(s string) []JsonMovieRes {
	var movies []JsonMovieRes
	if trie.root == nil {
		return movies
	}
	if s == "*" && len(s) == 1 {
		for k, _ := range trie.movieid_map {
			movies = append(movies, JsonMovieRes{k})
		}
		return movies
	}

	var res = NewIntSet()
	s = strings.ToLower(s)

	words := strings.Fields(s)

	for i := range words {
		var n *TNode = trie.root.findNode(words[i])
		if n != nil {
			n.collectPrefix(res)
		}
	}

	for i, _ := range res.m {
		movies = append(movies, JsonMovieRes{trie.records[i].Title})
	}

	return movies
}

func (rec *Record) addLocation(s string, ff string) {
	rec.Spots = append(rec.Spots, Location{s, ff, LatLng{0.0, 0.0}})
}

func initGoogleMapsClient() {
	var apiKey string
	// old key
	//apiKey = "AIzaSyCYy0Pt6UolytUOtxbdFdGkUA3iao0UkrA"
	apiKey = "AIzaSyAIqTV-SYMywiTDDtroFH3GicCmGyRLyng"
	var err error

	if apiKey != "" {
		gmapClient, err = maps.NewClient(maps.WithAPIKey(apiKey))
	} else {
		fmt.Println("Please specify an API Key, or Client ID and Signature.")
		return
	}
	if err != nil {
		fmt.Printf("Got an error while creating client")
		return
	}

}

func getStreetAddress(address string) LatLng {

	address = address + " San Francisco CA"

	r := &maps.GeocodingRequest{
		Address: address,
	}

	resp, err := gmapClient.Geocode(context.Background(), r)
	if err != nil || len(resp) == 0 {
		return LatLng{0.0, 0.0}
	}

	return LatLng{resp[0].Geometry.Location.Lat, resp[0].Geometry.Location.Lng}
}

func (trie *Trie) getMovieLocations(s string) *Record {
	id := trie.movieid_map[s]
	r := trie.records[id]
	_, ok := trie.geocache_map[id]
	if ok {
		return r
	}

	trie.updateLocations(r)
	trie.geocache_map[id] = true
	return r
}

func (trie *Trie) updateLocations(r *Record) {

	chs := make(chan bool, len(r.Spots))
	recs := 0
	for i, _ := range r.Spots {
		recs++
		if recs%9 == 0 {
			time.Sleep(1 * time.Second)
		}

		go func(r *Record, ind int) {
			xx := getStreetAddress(r.Spots[ind].Address)
			// Updating the lat and lng from the geocode service
			r.Spots[ind].Latlng.Lat = xx.Lat
			r.Spots[ind].Latlng.Lng = xx.Lng
			chs <- true
		}(r, i)
	}

	for range r.Spots {
		<-chs
	}
}

func (trie *Trie) readJsonFile() error {

	type ObjectType struct {
		Title    string `json:"title"`
		Location string `json:"locations"`
		A1       string `json:"actor_1"`
		A2       string `json:"actor_2"`
		A3       string `json:"actor_3"`
		Dir      string `json:"director"`
		Rel      string `json:"release_year"`
		Ff       string `json:"fun_facts"`
	}

	jsonurl := "https://data.sfgov.org/resource/yitu-d5am.json"
	//jsonurl := "https://data.sfgov.org/resource/yitu-d5an.json"

	response, err := http.Get(jsonurl)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	decoder := json.NewDecoder(response.Body)

	var elems []ObjectType
	decoder.Decode(&elems)

	index := 1
	fmt.Printf("Total records read: %d\n", len(elems))
	for _, p := range elems {
		if _, ok := trie.movieid_map[p.Title]; !ok {
			if p.Location == "" {
				continue
			}
			var acts []string
			if p.A1 != "" {
				acts = append(acts, p.A1)
			}
			if p.A2 != "" {
				acts = append(acts, p.A2)
			}
			if p.A3 != "" {
				acts = append(acts, p.A3)
			}

			newRec := &Record{Id: index, Title: p.Title, Director: p.Dir, Released: p.Rel, Actors: acts}

			newRec.addLocation(p.Location, p.Ff)

			trie.records[index] = newRec
			trie.movieid_map[p.Title] = index
			trie.addMovie(p.Title, index)
			index = index + 1
		} else {
			newRec := trie.records[trie.movieid_map[p.Title]]
			newRec.addLocation(p.Location, p.Ff)
		}
	}

	initGoogleMapsClient()
	return nil
}

// TrieNode high level APIs

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

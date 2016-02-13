package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"uber_movies/Godeps/_workspace/src/golang.org/x/net/context"
	"uber_movies/Godeps/_workspace/src/googlemaps.github.io/maps"
)

var (
	//mdb        *MovieDB
	gmapClient *maps.Client
)

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
	Spots    Locations
}

type MovieDB struct {
	mt           *Trie
	movieid_map  map[string]int
	records_map  map[int]*Record
	geocache_map map[int]bool // Just to cache the records which have the geocoding done
}

func NewMovieDB() *MovieDB {

	db := &MovieDB{}
	db.mt = &Trie{}
	db.movieid_map = make(map[string]int)
	db.records_map = make(map[int]*Record)
	db.geocache_map = make(map[int]bool)
	return db
}

func (rec *Record) addLocation(s string, ff string) {
	rec.Spots = append(rec.Spots, Location{s, ff, LatLng{0.0, 0.0}})
}

// Initialize the Google maps client
func initGoogleMapsClient() error {
	var apiKey string
	// old key
	//apiKey = "AIzaSyCYy0Pt6UolytUOtxbdFdGkUA3iao0UkrA"
	apiKey = "AIzaSyAIqTV-SYMywiTDDtroFH3GicCmGyRLyng"
	var err error
	gmapClient, err = maps.NewClient(maps.WithAPIKey(apiKey))

	if err != nil {
		fmt.Printf("Got an error while creating client")
		return err
	}
	return nil
}

// Get all the movie names in moviedb which prefix matches with sstring
func (m *MovieDB) getMovieNames(sstring string) []JsonMovieRes {
	var movies []JsonMovieRes

	if sstring == "*" && len(sstring) == 1 {
		for k, _ := range m.movieid_map {
			movies = append(movies, JsonMovieRes{k})
		}
		return movies
	}

	res := m.mt.getMovieIndices(sstring)
	for _, k := range res {
		movies = append(movies, JsonMovieRes{m.records_map[k].Title})
	}

	return movies
}

// Get the LatLng for a given address search string
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

// Get all the movie locations for a provided address string
func (m *MovieDB) getMovieLocations(s string) *Record {
	id := m.movieid_map[s]
	r := m.records_map[id]
	_, ok := m.geocache_map[id]
	if ok {
		return r
	}

	m.updateLocations(r)
	m.geocache_map[id] = true
	return r
}

// Update the LatLng for each address in a given Record
func (m *MovieDB) updateLocations(r *Record) {

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

//Read the main json file from the centtral server
func (m *MovieDB) readJsonFile() error {
	e := initGoogleMapsClient()
	if e != nil {
		return e
	}

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

	// Good json file
	jsonurl := "https://data.sfgov.org/resource/yitu-d5am.json"

	// Bad json file (for testing purposes)
	//jsonurl := "https://data.sfgov.org/resource/yatu-d5an.json"

	response, err := http.Get(jsonurl)


	if err != nil || response.StatusCode == 404 {
		return errors.New("Cannot find JsonFile")
	}

	defer response.Body.Close()

	decoder := json.NewDecoder(response.Body)

	var elems []ObjectType
	decoder.Decode(&elems)

	index := 1
	fmt.Printf("Total records read: %d\n", len(elems))
	for _, p := range elems {
		if _, ok := m.movieid_map[p.Title]; !ok {
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

			m.records_map[index] = newRec
			m.movieid_map[p.Title] = index
			m.mt.addMovie(p.Title, index)
			index = index + 1
		} else {
			newRec := m.records_map[m.movieid_map[p.Title]]
			newRec.addLocation(p.Location, p.Ff)
		}
	}

	return nil
}

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

// Json file for the movies data
// https://data.sfgov.org/resource/yitu-d5am.json

var (
	trie *Trie
)

func MoviesHandler(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	search_query := vars["searchWord"]

	var res []JsonMovieRes

	if search_query != "" {
		res = trie.getMovies(search_query)
	}

	json.NewEncoder(rw).Encode(res)
}

func LocationsHandler(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	movie_name := vars["movieName"]
	res := trie.getMovieLocations(movie_name)
	json.NewEncoder(rw).Encode(res)
}

func DefaultMoviesHandler(rw http.ResponseWriter, req *http.Request) {
	var res []JsonMovieRes
	json.NewEncoder(rw).Encode(res)
}

func initializeMovieServer() {
	trie = NewTrie()
	e := trie.readJsonFile()
	if e != nil {
		panic(e)
	}
}

func main() {

	r := mux.NewRouter()
	r.HandleFunc("/Movies/{searchWord}", MoviesHandler)
	r.HandleFunc("/Movies/", DefaultMoviesHandler)
	r.HandleFunc("/Locations/{movieName}", LocationsHandler)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./html/")))
	http.Handle("/", r)

	initializeMovieServer()

	var port = os.Getenv("PORT")
	if port == "" {
		port = "4748"
	}
	fmt.Println("Listening movie server on: ", port)
	err := http.ListenAndServe(":"+port, r)

	if err != nil {
		panic(err)
	}

}

package main

import (
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

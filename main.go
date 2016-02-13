package main

import (
	"fmt"
	"net/http"
	"os"

	"uber_movies/Godeps/_workspace/src/github.com/gorilla/mux"
)

// Json file for the movies data
// https://data.sfgov.org/resource/yitu-d5am.json

var (
	trie *Trie
	mdb  *MovieDB
)

func initializeMovieServer() error {
	mdb = NewMovieDB()
	e := mdb.readJsonFile()
	if e != nil {
		return e
	}
	return nil
}

func main() {

	r := mux.NewRouter()

	e := initializeMovieServer()
  if e != nil {
		panic(e)
  }

	r.HandleFunc("/Movies/{searchWord}", MoviesHandler)
	r.HandleFunc("/Movies/", DefaultMoviesHandler)
	r.HandleFunc("/Locations/{movieName}", LocationsHandler)
	//r.PathPrefix("/").Handler(http.FileServer(http.Dir("./html/")))
	if e != nil {
		r.HandleFunc("/", ErrorHandler)
	} else {
		r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("html/"))))
	}

	http.Handle("/", r)

	var port = os.Getenv("PORT")
	if port == "" {
		port = "4749"
	}
	fmt.Println("Listening movie server on: ", port)
	err := http.ListenAndServe(":"+port, r)

	if err != nil {
		panic(err)
	}

}

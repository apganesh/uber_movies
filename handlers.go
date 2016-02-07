package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
)

func DefaultHandler(rw http.ResponseWriter, req *http.Request) {
	body, _ := ioutil.ReadFile("html/index.html")
	fmt.Fprint(rw, string(body))
}
func ErrorHandler(rw http.ResponseWriter, req *http.Request) {
	body, _ := ioutil.ReadFile("errors/index.html")
	fmt.Fprint(rw, string(body))
}

// Handler for returning movies for a given pattern/patterns
func MoviesHandler(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	search_query := vars["searchWord"]

	var res []JsonMovieRes

	if search_query != "" {
		res = mdb.getMovieNames(search_query)
	}

	json.NewEncoder(rw).Encode(res)
}

// Handler for returning locations for a given movie
func LocationsHandler(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	movie_name := vars["movieName"]
	res := mdb.getMovieLocations(movie_name)
	json.NewEncoder(rw).Encode(res)
}

// Default handler for an empty movie string
func DefaultMoviesHandler(rw http.ResponseWriter, req *http.Request) {
	var res []JsonMovieRes
	json.NewEncoder(rw).Encode(res)
}

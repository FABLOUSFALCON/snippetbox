package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// Define a home handler function which writes a byte slice containing
// "Hello form Snippetbox" as the response body.
func home(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("Hello from Snippetbox"))
	if err != nil {
		log.Println("Got an error in home handler.", err)
	}
}

func snippetView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	msg := fmt.Sprintf("Display a specific snippet with id: %d", id)
	if _, err := w.Write([]byte(msg)); err != nil {
		log.Println("Got an error in snippetView handler.", err)
	}
}

func snippetCreate(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("Display a form for creating a new snippet..."))
	if err != nil {
		log.Println("Got an error in snippetCreate handler", err)
	}
}

func snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	if _, err := w.Write([]byte("Save a new snippet...")); err != nil {
		return
	}
}

func main() {
	// Use the http.NewServeMux() funciton to initialize a new new servemux, then
	// register the home function as the handler for the "/" URL pattern.
	mux := http.NewServeMux()
	mux.HandleFunc("GET	/{$}", home)
	mux.HandleFunc("GET	/snippet/view/{id}", snippetView)
	mux.HandleFunc("GET	/snippet/create", snippetCreate)
	mux.HandleFunc("POST	/snippet/create", snippetCreatePost)
	// Print a log message to say that the server is starting.
	log.Print("starting server on :4000")

	// Use the http.ListenAndServe() function to start a new web server. We pass in
	// two parameters: the TCP network address to listen on (in this case ":4000")
	// and the servemux we just created. If http.ListenAndServe() returns an error
	// we use the log.Fatal() function to log the error message and exit. Note
	// that any error returned by http.ListenAndServe() is always non-nil.
	err := http.ListenAndServe(":4000", mux)
	log.Fatal(err)
}

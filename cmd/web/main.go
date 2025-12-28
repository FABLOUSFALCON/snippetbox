package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")

	flag.Parse()

	// Use the http.NewServeMux() funciton to initialize a new new servemux, then
	// register the home function as the handler for the "/" URL pattern.
	mux := http.NewServeMux()

	// Adding FileServe to serve the static files.
	fs := http.FileServer(http.Dir("./ui/static/"))
	// Adding the Handler to serve static files.
	mux.Handle("GET /static/", http.StripPrefix("/static", fs))

	mux.HandleFunc("GET	/{$}", home)
	mux.HandleFunc("GET	/snippet/view/{id}", snippetView)
	mux.HandleFunc("GET	/snippet/create", snippetCreate)
	mux.HandleFunc("POST	/snippet/create", snippetCreatePost)
	// Print a log message to say that the server is starting.
	log.Printf("starting server on %s", *addr)

	// Use the http.ListenAndServe() function to start a new web server. We pass in
	// two parameters: the TCP network address to listen on (in this case ":4000")
	// and the servemux we just created. If http.ListenAndServe() returns an error
	// we use the log.Fatal() function to log the error message and exit. Note
	// that any error returned by http.ListenAndServe() is always non-nil.
	err := http.ListenAndServe(*addr, mux)
	log.Fatal(err)
}

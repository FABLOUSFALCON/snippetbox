package main

import (
	"net/http"

	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	// Use the http.NewServeMux() funciton to initialize a new new servemux, then
	// register the home function as the handler for the "/" URL pattern.
	mux := http.NewServeMux()

	// Adding FileServe to serve the static files.
	fs := http.FileServer(http.Dir("./ui/static/"))
	// Adding the Handler to serve static files.
	mux.Handle("GET /static/", http.StripPrefix("/static", fs))

	mux.HandleFunc("GET	/{$}", app.home)
	mux.HandleFunc("GET	/snippet/view/{id}", app.snippetView)
	mux.HandleFunc("GET	/snippet/create/", app.snippetCreate)
	mux.HandleFunc("POST	/snippet/create", app.snippetCreatePost)

	standard := alice.New(app.recoverPanic, app.logRequest, commonHeaders)

	return standard.Then(mux)
}

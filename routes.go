package main

import "net/http"

// Handler implements ServeHTTP(ResponseWriter, *Request)
// mux.Handle requires a Handler
// mux.HandleFunc requires pattern and handler func(ResponseWriter, *Request)

// Routes sets up all the routes for the server
func (s *WgServer) Routes() {
	s.mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		s.renderTemplatePage("index.html.tmpl", s).ServeHTTP(w, r)
	}))

	s.mux.Handle("/static/", gzipHandler(http.StripPrefix("/static/", http.FileServer(http.Dir("static")))))
	s.mux.Handle("/api/", s.handleAPI())
}

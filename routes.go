package main

import "net/http"

// Handler interface implements ServeHTTP(ResponseWriter, *Request)
// mux.Handle requires a Handler
// HandleFunc requires pattern and handler func(ResponseWriter, *Request)

func (s *Server) Routes() {
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, "Page not found", http.StatusNotFound)
			return
		}
		s.renderTemplatePage("index.html", nil).ServeHTTP(w, r)
	})
}

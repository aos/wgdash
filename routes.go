package main

import "net/http"

// Handler implements ServeHTTP(ResponseWriter, *Request)
// mux.Handle requires a Handler
// mux.HandleFunc requires pattern and handler func(ResponseWriter, *Request)

func (s *WgServer) Routes() {
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		s.renderTemplatePage("index.html.tmpl", nil).ServeHTTP(w, r)
	})
}

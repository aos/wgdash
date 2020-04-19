package main

// Handler interface implements ServeHTTP(ResponseWriter, *Request)
// mux.Handle requires a Handler
// HandleFunc requires pattern and handler func(ResponseWriter, *Request)

func (s *Server) Routes() {
	s.mux.Handle("/", s.renderTemplatePage("index.html", nil))
}

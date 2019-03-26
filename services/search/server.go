package search

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

const port = 6666

type link struct {
	Data interface{} `json:"data"`
	Meta interface{} `json:"meta"`
}

func (s *Service) searchHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")

	links, err := s.searcher.Search(r.Context(), q)
	if err != nil {
		writeJSONResponse(w, map[string]string{"error": err.Error()})
		return
	}

	res := make([]*link, len(links))

	for i, l := range links {
		var d interface{}
		json.Unmarshal(l.Data, &d)
		res[i] = &link{
			Data: d,
			Meta: l.Meta,
		}
	}

	writeJSONResponse(w, res)
}

func (s *Service) runHTTPServer() {

	r := mux.NewRouter()
	r.HandleFunc("/search", s.searchHandler).Methods("GET")

	log.Printf("Starting server on port %d\n", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), r)
}

// helpers

func writeJSONResponse(w http.ResponseWriter, data interface{}) {
	b, err := json.Marshal(data)
	if err != nil {
		log.Printf("ERROR: could not write JSON to response: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "could not write JSON to response")
		return
	}

	w.Header().Set("content-type", "application/json")
	w.Write(b)

}

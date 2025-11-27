package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
	"encoding/json"
	"log"
)

type apiConfig struct{
		fileserverHits atomic.Int32
	}

func healthHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Add("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(200)
	writer.Write([]byte("OK"))
}

func (cfg *apiConfig) hitsHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Add("Content-Type", "text/html; charset=utf-8")
	writer.WriteHeader(200)
	hits := int(cfg.fileserverHits.Load())
	hitsHTML := fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, hits)
	writer.Write([]byte(hitsHTML))
}

func (cfg *apiConfig) resetHitsHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Add("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(200)
	cfg.fileserverHits.Store(0)
}


func chirpValidationHandler(w http.ResponseWriter, r *http.Request){
    type parameters struct {
        // these tags indicate how the keys in the JSON should be mapped to the struct fields
        // the struct fields must be exported (start with a capital letter) if you want them parsed
        Body string `json:"body"`
    }

    decoder := json.NewDecoder(r.Body)
    params := parameters{}
    err := decoder.Decode(&params)
    if err != nil {
        // an error will be thrown if the JSON is invalid or has the wrong types
        // any missing fields will simply have their values in the struct set to their zero value
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
    }
	if len(params.Body) > 140 {
		type returnError struct {
        // the key will be the name of struct field unless you give it an explicit JSON tag
        Error bool `json:"error"`
    	}
		w.Header().Set("Content-Type", "application/json")
	}

    type returnVals struct {
        // the key will be the name of struct field unless you give it an explicit JSON tag
        Valid bool `json:"valid"`
    }
    respBody := returnVals{
        Valid: true,
    }
    dat, err := json.Marshal(respBody)
	if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
	}
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(200)
    w.Write(dat)
}


func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}


func main() {
	fmt.Println("Hello World")
	cfg := &apiConfig{}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/healthz", healthHandler)
	mux.HandleFunc("GET /admin/metrics", cfg.hitsHandler)
	mux.HandleFunc("POST /admin/reset", cfg.resetHitsHandler)
	mux.HandleFunc("POST /api/validate_chirp", chirpValidationHandler)
	handler := http.StripPrefix("/app", http.FileServer(http.Dir("./app/")))
	mux.Handle("/app/", cfg.middlewareMetricsInc(handler))
	server := http.Server{
		Addr: ":8080",
		Handler: mux,
	}
	server.ListenAndServe()
}
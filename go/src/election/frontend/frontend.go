package main

import (
	"log"
	"net/http"
)

// Redirect function to point to HTTPS page.
func redir(w http.ResponseWriter, req *http.Request) {
	log.Print(req)
	http.Redirect(w, req, "https://localhost:1443"+req.RequestURI, http.StatusMovedPermanently)
}

// Generic handler for a generic page.
func handler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("This is an example server.\n"))
}

func main() {
	http.HandleFunc("/", handler)
	log.Printf("About to listen on 1443. Go to https://127.0.0.1:1443/")
	go func() {
		if err := http.ListenAndServeTLS(":1443", "cert.pem", "key.pem", nil); err != nil {
			log.Fatal(err)
		}
	}()

	if err := http.ListenAndServe(":8000", http.HandlerFunc(redir)); err != nil {
		log.Fatal(err)
	}
}

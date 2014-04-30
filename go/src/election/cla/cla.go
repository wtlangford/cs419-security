package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"log"
	"net/http"
)

// Map of username <-> secrets
var secret map[string]string = map[string]string{
	"a":    "1",
	"b":    "2",
	"asdf": "1234",
}

// Map of username <-> validation numbers
var validation map[string]string = make(map[string]string)

// Generic handler for a generic page.
func handler() string {
	return "hello world"
}

type Registration struct {
	Name   string `form:"name"`
	Secret string `form:"secret"`
}

// RegisterUser takes a user provided registration, then returns a random validation
// number if their (name, secret) pair is correct. If a number was already requested,
// then the previous number is returned again.
func RegisterUser(reg Registration) (int, []byte) {
	log.Println("Received:", reg)
	if reg.Secret == secret[reg.Name] {
		if validation[reg.Name] == "" {
			// Generate random 1024 bit number
			b := make([]byte, 128)
			_, err := rand.Read(b)
			if err != nil {
				log.Println("error:", err)
				return 500, []byte("Error generating validation number.")
			}
			validation[reg.Name] = base64.StdEncoding.EncodeToString(b)
		}

		// Return JSON response with base64 encoded validation number
		res, _ := json.Marshal(map[string]string{"Validation": validation[reg.Name]})
		return 200, res
	}
	return 403, []byte("User does not exist.")
}

func main() {
	m := martini.Classic()
	m.Post("/register", binding.Bind(Registration{}), RegisterUser)
	m.Get("/", handler)
	log.Println("About to listen on 1444. Go to https://localhost:1444/")
	log.Fatal(http.ListenAndServeTLS(":1444", "cert.pem", "key.pem", m))
}

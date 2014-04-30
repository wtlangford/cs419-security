package main

import (
	"log"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"net/http"
)

var db map[string]string = map[string]string {
	"a": "1",
	"b": "2",
}

// Generic handler for a generic page.
func handler() string {
	return "hello world"
}

type Registration struct {
	Name string `form:"name"`
	Secret string `form:"secret"`
}

func RegisterUser(reg Registration) string{
	log.Println(reg)
	return "asdf"
}

func main() {
	m := martini.Classic()
	m.Post("/register", binding.Bind(Registration{}), RegisterUser)
	m.Get("/", handler)
	log.Println("About to listen on 1444. Go to https://localhost:1444/")
	log.Fatal(http.ListenAndServeTLS(":1444", "cert.pem", "key.pem", m))
}

package main

import (
	"fmt"
	"net/http"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"sync"
)

type CTF struct{
	sync.RWMutex
	validationNumbers map[string]bool
	votes map[string] []string
}

type FormPost struct {
	ValNum string `form:"vn"`
	Id string `form:"id"`
	Vote string `form:"vote"`
}


var ctf CTF = CTF{validationNumbers:make(map[string]bool), votes:make(map[string] []string)}
var choices []string = []string{"tacocat","racecar","radar","civic"}

func stringInSlice(a string, list []string) bool {
    for _, b := range list {
        if b == a {
            return true
        }
    }
    return false
}

func addValidationNumber(params FormPost) string {
	vn := params.ValNum
	ctf.Lock()
	_, ok := ctf.validationNumbers[vn]
	if ok {
		ctf.Unlock()
		return "BAD!"
	}
	ctf.validationNumbers[vn] = true
	str := fmt.Sprint(ctf.validationNumbers)
	ctf.Unlock()
	return str
}

func vote(params FormPost) string {
	vn := params.ValNum
	id := params.Id
	vote := params.Vote

	ctf.Lock()
	if v,ok := ctf.validationNumbers[vn]; ok == false {
		ctf.Unlock()
		return fmt.Sprint("This vn does not exist...",vn,"x",params,"\n")
	} else if  v == false {
		ctf.Unlock()
		return "This vn has already voted..."
	} else if !stringInSlice(vote,choices) {
		ctf.Unlock()
		return "Invalid vote"
	}

	ctf.votes[vote] = append(ctf.votes[vote],id)
	ctf.validationNumbers[vn] = false
	res := fmt.Sprint(ctf.votes[vote])
	ctf.Unlock()
	return res
}

func main() {
	m := martini.Classic()
	m.Post("/vn",binding.Bind(FormPost{}),addValidationNumber)
	m.Post("/vote",binding.Bind(FormPost{}),vote)
	m.Get("/results",func() string {
		ctf.RLock()
		str := fmt.Sprint(ctf.votes)
		ctf.RUnlock()
		return str
	})

	m.Get("/",func() string {
		return "Martini up!"
	})
	http.ListenAndServe("wlangford.net:4000",m)
}

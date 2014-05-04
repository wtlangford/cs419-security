package main

import (
	"bytes"
	"crypto/rsa"
	"crypto/tls"
	"election/common"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

type CTF struct {
	sync.RWMutex
	// Map of VN -> Can the user vote
	validationNumbers map[string]bool
	// Map of unique user IDs -> VN
	ids								map[string]string
	// Map of candidates -> list of voters
	votes             map[string][]string
}

type FormPost struct {
	ValNum string `form:"vn"`
	Sig    string `form:"sig"`
	Id     string `form:"id"`
	Vote   string `form:"vote"`
}

var ctf CTF = CTF {
	validationNumbers: make(map[string]bool),
	votes: make(map[string][]string),
	ids: make(map[string]string),
}
var choices []string = []string{"tacocat", "racecar", "radar", "civic"}
var claKey *rsa.PublicKey

func addValidationNumber(params FormPost) (int, string) {
	vn := params.ValNum
	sig := params.Sig

	rawSig, _ := base64.StdEncoding.DecodeString(sig)
	if err := common.VerifySig([]byte(vn), rawSig, claKey); err != nil {
		log.Println(err)
		return 400, "Bad Request"
	} else {
		log.Println("VN validated")
	}

	ctf.Lock()

	_, ok := ctf.validationNumbers[vn]
	if ok {
		ctf.Unlock()
		return 500, "BAD!"
	}
	ctf.validationNumbers[vn] = true
	str := fmt.Sprint(ctf.validationNumbers)
	ctf.Unlock()
	return 200, str
}

func vote(params FormPost) (int, string) {
	vn := params.ValNum
	id := params.Id
	vote := params.Vote

	ctf.Lock()
	if _, ok := ctf.ids[params.Id]; ok {
		ctf.Unlock()
		return 400, "ID is already in use"
	} else {
		ctf.ids[params.Id] = vn
	}
	if v, ok := ctf.validationNumbers[vn]; ok == false {
		ctf.Unlock()
		return 401, fmt.Sprint("This vn does not exist...", vn, "x", params, "\n")
	} else if v == false {
		ctf.Unlock()
		return 403, "This vn has already voted..."
	} else if !common.StringInSlice(vote, choices) {
		ctf.Unlock()
		return 400, "Invalid vote"
	}

	ctf.votes[vote] = append(ctf.votes[vote], id)
	ctf.validationNumbers[vn] = false
	res := fmt.Sprint(ctf.votes[vote])
	ctf.Unlock()
	return 200, res
}

type Results struct {
	Votes map[string][]string
	Voters []string
}

func getResults() []byte {
	// Get list of validation numbers
	vn := make([]string, len(ctf.validationNumbers))
	for key, _ := range ctf.validationNumbers {
		vn = append(vn, key)
	}

	// Send VNs to CLA, get back list of voters
	payload, _ := json.Marshal(map[string][]string{"payload":vn})
	t := &http.Transport {
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: t}
	resp, err := client.Post("https://localhost:1444/voters", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var voters []string
	if err := json.Unmarshal(body, &voters); err != nil {
		log.Println("Unmarshal:", err)
	}

	// Return results
	results := Results{ctf.votes, voters}
	retval, err := json.Marshal(results)
	if err != nil {
		log.Println("Marshal: ", err)
	}
	return retval
}

func main() {
	var err error
	if claKey, err = common.ReadPublicKey("cla-rsa.pub"); err != nil {
		log.Fatal(err)
	}
	m := martini.Classic()

	m.Post("/vn", binding.Bind(FormPost{}), addValidationNumber)
	m.Post("/vote", binding.Bind(FormPost{}), vote)
	m.Get("/results", getResults)

	m.Get("/", func() string {
		return "Martini up!"
	})
	http.ListenAndServeTLS(":4000", "cert.pem", "key.pem", m)
}

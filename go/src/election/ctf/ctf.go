package main

import (
	"bufio"
	"bytes"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"election/common"
	"encoding/json"
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
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
var stopped bool = false
var choices []string = []string{"tacocat", "racecar", "radar", "civic"}
var claKey *rsa.PublicKey
var privateKey *rsa.PrivateKey
var certPool *x509.CertPool = x509.NewCertPool()

func addValidationNumber(params FormPost) (int, string) {
	if stopped {
		return 401, "Voting ended"
	}
	vn := params.ValNum
	sig := params.Sig

	if err := common.VerifySig([]byte(vn), sig, claKey); err != nil {
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
	if stopped {
		return 401, "Voting ended"
	}
	vn := params.ValNum
	id := params.Id
	vote := params.Vote

	ctf.Lock()
	if _, ok := ctf.ids[params.Id]; ok {
		ctf.Unlock()
		return 400, "ID is already in use"
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

	ctf.ids[params.Id] = vn
	ctf.votes[vote] = append(ctf.votes[vote], id)
	ctf.validationNumbers[vn] = false
	ctf.Unlock()
	return 200, "Vote accepted"
}

type Results struct {
	Votes map[string][]string
	Voters []string
}

func getResults() (int, []byte) {
	if !stopped {
		return 401, []byte("Voting not ended")
	}
	// Get list of validation numbers
	vn := []string{}
	for key, notVoted := range ctf.validationNumbers {
		if !notVoted {
			vn = append(vn, key)
		}
	}

	// Send VNs to CLA, get back list of voters
	payload, _ := json.Marshal(vn)
	sig, err := common.SignData(payload,privateKey)
	if err != nil {
		return 500, []byte("Could not sign data")
	}
	payload, _ = json.Marshal(map[string]string{"payload":string(payload),"sig":sig})
	t := &http.Transport {
		TLSClientConfig: &tls.Config{RootCAs:certPool},
	}
	client := &http.Client{Transport: t}
	resp, err := client.Post("https://cla.wlangford.net:1444/voters", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Println(err)
		return 500, []byte("Could not retrieve voters")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var voters []string
	if err := json.Unmarshal(body, &voters); err != nil {
		log.Println("Unmarshal:", err)
	}

	for _,votes := range ctf.votes {
		sort.Strings(votes)
	}
	// Return results
	results := Results{ctf.votes, voters}
	retval, err := json.MarshalIndent(results," ","  ")
	if err != nil {
		log.Println("Marshal: ", err)
	}
	return 200, retval
}

func endTest() {
	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("Can't read from stdin.  Cry.  A lot.")
		}
		log.Println(line)
		if line == "kill\n" {
			stopped = true
			log.Println("Shut it down.")
		}
	}


}

func main() {
	var err error
	if claKey, err = common.ReadPublicKey("cla-rsa.pub"); err != nil {
		log.Fatal(err)
	}
	pemFile, err := ioutil.ReadFile("/var/www/CA/certs/cacert.crt")
	if err != nil {
		log.Fatal(err)
	}
	privateKey, err = common.ReadPrivateKey("ctf-rsa")
	if err != nil {
		log.Fatal(err)
	}

	if certPool.AppendCertsFromPEM(pemFile) == false {
		log.Fatal("Could not read root ca")
	}

	m := martini.Classic()

	m.Post("/vn", binding.Bind(FormPost{}), addValidationNumber)
	m.Post("/vote", binding.Bind(FormPost{}), vote)
	m.Get("/results", getResults)
	go endTest()

	m.Get("/", func() string {
		return "Martini up!"
	})
	http.ListenAndServeTLS("ctf.wlangford.net:4000", "cert.pem", "key.pem", m)
}

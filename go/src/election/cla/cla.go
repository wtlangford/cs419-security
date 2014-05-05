package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"election/common"
	"encoding/base64"
	"encoding/json"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sort"
)

// Map of username <-> secrets
var secret map[string]string = map[string]string{
	"a":    "1",
	"b":    "2",
	"asdf": "1234",
}

// Map of username -> validation numbers
var validation map[string]string = make(map[string]string)
// Map of validation numbers -> usernames
var voter map[string]string = make (map[string]string)

// Security Stuff
var privKey *rsa.PrivateKey
var certPool *x509.CertPool = x509.NewCertPool()

// Generic handler for a generic page.
func handler() string {
	return "hello world"
}

type Registration struct {
	Name   string `form:"name"`
	Secret string `form:"secret"`
}

type ValidationNumbers struct {
	Payload	[]string
}

// RegisterUser takes a user provided registration, then returns a random validation
// number if their (name, secret) pair is correct. If a number was already requested,
// then the previous number is returned again.
func RegisterUser(reg Registration) (int, []byte) {
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
			voter[validation[reg.Name]] = reg.Name
			SendToCTF(validation[reg.Name])
		}

		// Return JSON response with base64 encoded validation number
		res, _ := json.Marshal(map[string]string{"Validation": validation[reg.Name]})
		return 200, res
	}
	return 403, []byte("User does not exist.")
}

// SendToCTF sends the VN # payload to the CTF server.
func SendToCTF(payload string) {
	sig := common.SignData([]byte(payload), privKey)
	log.Println(sig)

	// Need to ignore self-signed cert. Signatures will be used to confirm identity
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{RootCAs:certPool},
	}
	client := &http.Client{Transport: tr}
	if _, err := client.PostForm("https://ctf.wlangford.net:4000/vn",
		url.Values{"vn": {payload}, "sig": {sig}}); err != nil {
		log.Println(err)
	}
}

// GetVotingUsers returns a list of users who voted
func GetVotingUsers(data ValidationNumbers) (int, []byte){
	voters := []string{}
	for _, vn := range data.Payload {
		if voter[vn] != "" {
			voters = append(voters, voter[vn])
		}
	}
	// Sort users so there is no particular ordering of who voted
	sort.Strings(voters)
	retval, err := json.Marshal(voters)
	if err != nil {
		log.Fatal(err)
	}
	return 200, retval
}

func main() {
	var err error
	if privKey, err = common.ReadPrivateKey("cla-rsa"); err != nil {
		log.Fatal(err)
	}
	pemFile, err := ioutil.ReadFile("/var/www/CA/certs/cacert.crt")
	if err != nil {
		log.Fatal(err)
	}
	certPool.AppendCertsFromPEM(pemFile)


	m := martini.Classic()
	m.Post("/register", binding.Bind(Registration{}), RegisterUser)
	m.Post("/voters", binding.Bind(ValidationNumbers{}), GetVotingUsers)
	m.Get("/", handler)
	log.Println("About to listen on 1444. Go to https://localhost:1444/")
	log.Fatal(http.ListenAndServeTLS("cla.wlangford.net:1444", "cert.pem", "key.pem", m))
}

package common

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"log"
)

func SignData(payload []byte, k *rsa.PrivateKey) (string, error) {
	hash := sha512.New()
	hash.Write(payload)
	crypt, err := rsa.SignPKCS1v15(rand.Reader, k, crypto.SHA512, hash.Sum(nil))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(crypt), nil
}

func VerifySig(message []byte, sig string, k *rsa.PublicKey) error {
	h := sha512.New()
	h.Write(message)
	d := h.Sum(nil)
	sigData,err := base64.StdEncoding.DecodeString(sig)
	if err != nil {
		return err
	}
	return rsa.VerifyPKCS1v15(k, crypto.SHA512, d, sigData)
}

func ReadPublicKey(path string) (*rsa.PublicKey, error) {
	// Read in keys
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(buf)
	log.Println(block.Type)
	pubkey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	switch t := pubkey.(type) {
	case *rsa.PublicKey:
		return t, nil
	default:
		return nil, errors.New("unknown key type")
	}
}

func ReadPrivateKey(path string) (*rsa.PrivateKey, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(buf)
	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return privKey, nil
}

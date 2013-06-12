package main

import (
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/voxelbrain/goptions"

	"code.google.com/p/goauth2/oauth/jwt"
	"code.google.com/p/google-api-go-client/googleapi/transport"
	storage "code.google.com/p/google-api-go-client/storage/v1beta1"
)

const (
	keyFile = "key.pem"
)

var (
	options = struct {
		ClientId string        `goptions:"-c, --client-id, description='ClientID', obligatory"`
		KeyFile  *os.File      `goptions:"-k, --key, description='PEM file containing the private key', obligatory"`
		Bucket   string        `goptions:"-b, --bucket, description='Name of bucket', obligatory"`
		Help     goptions.Help `goptions:"-h, --help, description='Show this help'"`
	}{}
)

// 607613198303.apps.googleusercontent.com
func main() {
	goptions.ParseAndFail(&options)
	key, err := readKey(options.KeyFile)
	if err != nil {
		log.Fatalf("Could not read keyfile: %s", err)
	}

	token := &jwt.Token{
		Key: key,
		ClaimSet: &jwt.ClaimSet{
			Iss:   options.ClientId,
			Scope: "https://www.googleapis.com/auth/devstorage.read_only",
			Aud:   "https://accounts.google.com/o/oauth2/token",
		},
	}
	c := &http.Client{}
	o, err := token.Assert(c)
	if err != nil {
		log.Fatalf("Could not get OAuth token: %s", err)
	}

	c.Transport = &transport.APIKey{Key: o.AccessToken}
	service, err := storage.New(c)
	if err != nil {
		log.Fatalf("Could not use storage API: %s", err)
	}
	objs, err := service.Objects.List(options.Bucket).Do()
	if err != nil {
		log.Fatalf("Could not list bucket %s: %s", options.Bucket, err)
	}
	for _, obj := range objs.Items {
		fmt.Printf("%s/%s\n", options.Bucket, obj.Name)
	}
}

func readKey(r io.ReadCloser) ([]byte, error) {
	defer r.Close()
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return data, nil

	b, _ := pem.Decode(data)
	if b == nil {
		return nil, fmt.Errorf("No key found in %s", keyFile)
	}
	return b.Bytes, nil
}

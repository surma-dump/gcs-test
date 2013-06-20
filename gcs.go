package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/voxelbrain/goptions"

	"code.google.com/p/goauth2/oauth"
	"code.google.com/p/goauth2/oauth/jwt"
	storage "code.google.com/p/google-api-go-client/storage/v1beta2"
)

var (
	options = struct {
		ClientId string        `goptions:"-c, --client-id, description='ClientID', obligatory"`
		KeyFile  *os.File      `goptions:"-k, --key, description='PEM file containing the private key', obligatory", ro`
		Bucket   string        `goptions:"-b, --bucket, description='Name of bucket', obligatory"`
		Help     goptions.Help `goptions:"-h, --help, description='Show this help'"`
	}{}
)

func main() {
	goptions.ParseAndFail(&options)
	defer options.KeyFile.Close()
	pemBytes, err := ioutil.ReadAll(options.KeyFile)
	if err != nil {
		log.Fatalf("Could not read keyfile: %s", err)
	}

	token := jwt.NewToken(options.ClientId, storage.DevstorageRead_writeScope, pemBytes)
	// token.ClaimSet.Aud = aud
	c := &http.Client{}
	oauthToken, err := token.Assert(c)
	if err != nil {
		log.Fatalf("Could not get OAuth token: %s", err)
	}

	c.Transport = &oauth.Transport{
		Token: oauthToken,
	}
	service, err := storage.New(c)
	if err != nil {
		log.Fatalf("Could not use storage API: %s", err)
	}
	objs, err := service.Objects.List(options.Bucket).Do()
	if err != nil {
		log.Fatalf("Could not list content of bucket %s: %s", options.Bucket, err)
	}
	for _, obj := range objs.Items {
		log.Printf("%s/%s: %s\n", options.Bucket, obj.Name, obj.SelfLink)
	}

	data := strings.NewReader("Some Data")
	newObj, err := service.Objects.Insert(options.Bucket, &storage.Object{}).Name("gcs-test").Media(data).Do()
	if err != nil {
		log.Fatalf("Could not create new object: %s", err)
	}
	log.Printf("Uploaded to %s", newObj.SelfLink)
}

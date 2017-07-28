package main

import (
	"fmt"
	"log"

	"github.com/googleapis/gnostic/plugins/gnostic-go-generator/examples/googleauth"
	"github.com/googleapis/gnostic/plugins/gnostic-go-generator/examples/v3.0/urlshortener/urlshortener"
)

func main() {
	client, err := googleauth.NewOAuth2Client("https://www.googleapis.com/auth/urlshortener")
	if err != nil {
		log.Fatalf("Error building OAuth client: %v", err)
	}
	fmt.Printf("client: %+v\n", client)

	path := "https://www.googleapis.com/urlshortener/v1" // this should be generated
	c := urlshortener.NewClient(path, client)

	fmt.Printf("OK %+v\n", c)

	// get
	if true {
		response, err := c.Urlshortener_Url_Get("FULL", "https://goo.gl/mUnLia")
		fmt.Printf("\nGET\n")
		fmt.Printf("response = %+v\n", response.Default)

		fmt.Printf("response = %s %s\n", response.Default.Id, response.Default.LongUrl)
		fmt.Printf("err = %+v\n", err)
	}

	// list
	if true {
		fmt.Printf("\nLIST\n")
		response, err := c.Urlshortener_Url_List("", "")
		fmt.Printf("response = %+v\n", response.Default)
		fmt.Printf("err = %+v\n", err)
		for _, item := range response.Default.Items {
			fmt.Printf("Id=%s LongUrl=%s\n", item.Id, item.LongUrl)
		}
	}

	// insert
	if false {
		fmt.Printf("\nINSERT\n")
		var url urlshortener.Url
		url.LongUrl = "https://github.com/googleapis/gnostic/wiki"
		response, err := c.Urlshortener_Url_Insert(&url)
		fmt.Printf("response = %+v\n", response.Default)
		fmt.Printf("err = %+v\n", err)
	}
}

package main

import (
	"log"
	"net/http"
	"os"

	"github.com/TonkyH/imageMagick/go/pubsub"
)

func main() {
	http.HandleFunc("/", pubsub.HelloPubSub)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

package pubsub

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/TonkyH/imageMagick/go/imagemagick"
)

type PubSubMessage struct {
	Message struct {
		Data []byte `json:"data,omitempty"`
		ID   string `json:"id"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}

// HelloPubSub received and processes a Pub/Sub push message
func HelloPubSub(w http.ResponseWriter, r *http.Request) {
	var m PubSubMessage
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("io.ReadAll: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &m); err != nil {
		log.Printf("json.Unmarshal: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var e imagemagick.GCSEvent
	if err := json.Unmarshal(m.Message.Data, &e); err != nil {
		log.Printf("json.Unmarshal: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if e.Name == "" || e.Bucket == "" {
		log.Printf("invalid GCSEvent: expected name and bucket")
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if err := imagemagick.BlurOffensiveImages(r.Context(), e); err != nil {
		log.Printf("imagemagick.BlurOffensiveImages: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

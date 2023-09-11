package main

import (
	"bytes"
	"encoding/json"
	"github.com/joho/godotenv"
	"github.com/mmcdole/gofeed"
	"log"
	"net/http"
	"os"
	"time"
)

type DiscordMessage struct {
	Content string `json:"content"`
}

var lastUpdate time.Time

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	http.HandleFunc("/pollRSS", pollRSSHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func pollRSSHandler(w http.ResponseWriter, r *http.Request) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(os.Getenv("SUBSTACK_RSS_URL"))
	if err != nil {
		log.Printf("Error fetching RSS: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if feed.UpdatedParsed.After(lastUpdate) {
		lastUpdate = *feed.UpdatedParsed
		sendToDiscord("New content available on Substack!")
	}
	w.Write([]byte("Checked RSS feed!"))
}

func sendToDiscord(content string) {
	webhookURL := os.Getenv("DISCORD_WEBHOOK_URL")

	msg := &DiscordMessage{
		Content: content,
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshalling message: %v", err)
		return
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Printf("Error sending message to Discord: %v", err)
		return
	}
	defer resp.Body.Close()
}

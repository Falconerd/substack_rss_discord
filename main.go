package main

import (
	"bytes"
	"encoding/json"
	"github.com/joho/godotenv"
	"github.com/mmcdole/gofeed"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type DiscordMessage struct {
	Content string `json:"content"`
}

var lastUpdate time.Time

const lastUpdateFile = "lastUpdate.txt"

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	loadData, err := ioutil.ReadFile(lastUpdateFile)
	if err == nil {
		lastUpdate, _ = time.Parse(time.RFC3339, string(loadData))
	}

	ticker := time.NewTicker(1 * time.Hour)
	for {
		select {
		case <-ticker.C:
			checkForUpdates()
		}
	}
}

func checkForUpdates() {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(os.Getenv("SUBSTACK_RSS_URL"))
	if err != nil {
		log.Printf("Error fetching RSS: %v", err)
		return
	}

	if feed.UpdatedParsed.After(lastUpdate) {
		lastUpdate = *feed.UpdatedParsed
		sendToDiscord("New content available on Substack!")

		ioutil.WriteFile(lastUpdateFile, []byte(lastUpdate.Format(time.RFC3339)), 0644)
	}
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

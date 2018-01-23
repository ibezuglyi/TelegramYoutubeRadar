package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/youtube/v3"
)

var (
	query      = flag.String("query", "Google", "Search term")
	maxResults = flag.Int64("max-results", 50, "Max YouTube results")
)

type YoutubeFinder struct {
	service *youtube.Service
}

//ctor for yf
func New(key string) *YoutubeFinder {
	flag.Parse()

	client := &http.Client{
		Transport: &transport.APIKey{Key: developerKey},
	}
	service, err := youtube.New(client)
	if err != nil {
		log.Fatalf("Error creating new YouTube client: %v", err)
	}
	return &YoutubeFinder{service}
}

func (yf *YoutubeFinder) searchListByLocation(latlon string, q string) []string {
	part := "id,snippet"
	call := yf.service.Search.List(part)
	call = call.Location(latlon)
	call = call.LocationRadius("10mi")
	call = call.SafeSearch("none")
	call = call.VideoCaption("any")
	call = call.VideoDefinition("any")
	call = call.VideoDuration("any")
	if q != "" {
		call = call.Q(q)
	}
	now := time.Now().Add(-200 * time.Minute)
	formattedTime := now.Format("2006-01-02T15:04:05Z")
	call = call.PublishedAfter(formattedTime)
	call = call.Type("video")
	call = call.MaxResults(*maxResults)
	response, err := call.Do()
	if err != nil {
		panic(err)
	}
	videos := make(map[string]string)
	for _, item := range response.Items {
		switch item.Id.Kind {
		case "youtube#video":
			videos[item.Id.VideoId] = item.Snippet.Title
		}
	}
	var urls []string
	for id := range videos {
		url := fmt.Sprintf("http://youtube.com/watch?v=%v", id)
		urls = append(urls, url)
	}
	return urls
}

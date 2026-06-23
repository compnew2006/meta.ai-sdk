// Example: image + video generation.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/smart-studio/metaai-go"
)

func main() {
	client, err := metaai.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	img, err := client.GenerateImage(ctx, "a serene mountain lake at sunrise", "LANDSCAPE", 1)
	if err != nil {
		log.Printf("image gen error: %v", err)
	} else {
		fmt.Printf("image: success=%v status=%s urls=%v\n", img.Success, img.Status, img.URLs)
	}

	vid, err := client.GenerateVideo(ctx, "ocean waves at sunset")
	if err != nil {
		log.Printf("video gen error: %v", err)
	} else {
		fmt.Printf("video: success=%v status=%s urls=%v\n", vid.Success, vid.Status, vid.URLs)
	}
}

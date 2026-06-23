// Example: minimal chat. Run with:
//
//	META_AI_DATR=... META_AI_ECTO_1_SESS=... META_AI_ACCESS_TOKEN=ecto1:... \
//	  go run ./examples/simple
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/smart-studio/metaai-go"
)

func main() {
	client, err := metaai.NewClient() // reads META_AI_* env vars / .env
	if err != nil {
		log.Fatal(err)
	}
	if !client.IsAuthed() {
		log.Fatal("not authed: set META_AI_DATR / META_AI_ECTO_1_SESS / META_AI_ACCESS_TOKEN")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Non-streaming chat.
	reply, err := client.Chat(ctx, "Reply with exactly: PONG1", nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("reply:", reply)

	// Streaming chat (range over the channel).
	for chunk := range client.StreamChat(ctx, "What is 2+2?", nil) {
		if chunk.Err != nil {
			log.Fatal(chunk.Err)
		}
		fmt.Print(chunk.Text)
	}
	fmt.Println()

	_ = os.Args
}

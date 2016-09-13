package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

func main() {
	config := readConfig()
	stream := openTwitterStream(config)

	attachMessageHandlers(stream)
	keepRunning(stream)
}

// Set twitter stream handlers
func attachMessageHandlers(stream *twitter.Stream) {
	demux := twitter.NewSwitchDemux()

	demux.Tweet = func(tweet *twitter.Tweet) {
		if len(tweet.Entities.Urls) != 0 {
			fmt.Println(tweet.Text)
			for _, url := range tweet.Entities.Urls {
				fmt.Printf("- %#v\n", url.ExpandedURL)
			}
			fmt.Println("---")
		}
	}

	//demux.StreamDisconnect = func(disconnect *twitter.StreamDisconnect) {
	//}
	//demux.StallWarning = func(warning *twitter.StallWarning) {
	//}

	// Receive messages until stopped or stream quits
	go demux.HandleChan(stream.Messages)
}

func openTwitterStream(config Config) *twitter.Stream {
	oauthConfig := oauth1.NewConfig(config.ConsumerKey, config.ConsumerSecret)
	token := oauth1.NewToken(config.AccessToken, config.AccessSecret)
	httpClient := oauthConfig.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	fmt.Println("Opening sample stream...")
	streamParams := &twitter.StreamSampleParams{
		StallWarnings: twitter.Bool(true),
	}
	stream, err := client.Streams.Sample(streamParams)
	if err != nil {
		log.Fatal(err)
	}

	return stream
}

// Wait for SIGINT and SIGTERM (HIT CTRL-C)
func keepRunning(stream *twitter.Stream) {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)
	stream.Stop()
}

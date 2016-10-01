package stream

import (
	"log"

	"../brain"
	"../config"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

type stream struct {
	c   chan *brain.UrlFetch
	s   *twitter.Stream
	log *log.Logger
}

func NewStream(c chan *brain.UrlFetch, log *log.Logger) *stream {
	s := &stream{c: c, log: log}
	s.createTwitterStream()
	return s
}

// Set twitter stream handlers
func (s *stream) Start() {
	demux := twitter.NewSwitchDemux()

	demux.Tweet = func(tweet *twitter.Tweet) {
		if tweet.Entities != nil && len(tweet.Entities.Urls) != 0 {
			for _, u := range tweet.Entities.Urls {
				if len(u.ExpandedURL) > 0 {
					s.c <- brain.NewUrlFetch(u.ExpandedURL)
				}
			}
		}
	}

	go demux.HandleChan(s.s.Messages)
}

func (s *stream) Stop() {
	s.s.Stop()
}

func (s *stream) createTwitterStream() {
	c := config.Twitter
	oauthConfig := oauth1.NewConfig(c.ConsumerKey, c.ConsumerSecret)
	token := oauth1.NewToken(c.AccessToken, c.AccessSecret)
	httpClient := oauthConfig.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	log.Println("Opening sample stream...")
	streamParams := &twitter.StreamSampleParams{
		StallWarnings: twitter.Bool(true),
	}
	stream, err := client.Streams.Sample(streamParams)
	if err != nil {
		log.Fatal(err)
	}

	s.s = stream
}

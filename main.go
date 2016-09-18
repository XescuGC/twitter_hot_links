package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

type url struct {
	count   int
	created time.Time
	score   float64
	url     string
}

type UrlsByScore []url

func (u UrlsByScore) Len() int           { return len(u) }
func (u UrlsByScore) Swap(i, j int)      { u[i], u[j] = u[j], u[i] }
func (u UrlsByScore) Less(i, j int) bool { return u[i].score > u[j].score }

var urls = make(map[string]*url)

func main() {
	config := readConfig()
	stream := openTwitterStream(config)

	tickChan := time.NewTicker(time.Minute * 1).C
	signalChannel := make(chan os.Signal)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	attachMessageHandlers(stream)

	for {
		select {
		case <-signalChannel:
			fmt.Println("Stoping ...")
			stream.Stop()
			showCollectedData()
			return
		case <-tickChan:
			fmt.Println("Printing state ...")
			showCollectedData()
		}
	}
}

// Set twitter stream handlers
func attachMessageHandlers(stream *twitter.Stream) {
	demux := twitter.NewSwitchDemux()

	demux.Tweet = func(tweet *twitter.Tweet) {
		if tweet.Entities != nil && len(tweet.Entities.Urls) != 0 {
			for _, u := range tweet.Entities.Urls {
				if val, ok := urls[u.ExpandedURL]; ok {
					val.count++
				} else {
					urls[u.ExpandedURL] = &url{count: 0, created: time.Now(), url: u.ExpandedURL}
				}
			}
		}
	}

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

func showCollectedData() {
	urlsByScore := make([]url, 1)
	for _, v := range urls {
		if v.count > 1 {
			power := math.Pow(time.Now().Sub(v.created).Minutes(), 1.8)
			v.score = float64(v.count) / power
			urlsByScore = append(urlsByScore, *v)
		}
	}
	sort.Sort(UrlsByScore(urlsByScore))
	for i, v := range urlsByScore {
		if i == 100 {
			break
		}
		fmt.Printf("%f %d %s\n", v.score, v.count, v.url)
	}
}

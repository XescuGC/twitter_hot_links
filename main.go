package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
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

type urlFetch struct {
	url     string
	fetched bool
}

func NewUrlFetch(url string) *urlFetch {
	return &urlFetch{url: url, fetched: false}
}

type UrlsByScore []url

func (u UrlsByScore) Len() int           { return len(u) }
func (u UrlsByScore) Swap(i, j int)      { u[i], u[j] = u[j], u[i] }
func (u UrlsByScore) Less(i, j int) bool { return u[i].score > u[j].score }

var urls = make(map[string]*url)
var timeout = time.Duration(10 * time.Second)
var client = http.Client{
	Timeout: timeout,
}

func main() {
	config := readConfig()
	stream := openTwitterStream(config)

	tickChan := time.NewTicker(time.Minute * 1).C
	signalChannel := make(chan os.Signal)
	tweetChannel := make(chan *urlFetch)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	attachMessageHandlers(stream, tweetChannel)

	for {
		select {
		case <-signalChannel:
			fmt.Println("Stoping ...")
			stream.Stop()
			showCollectedData()
			//close(tickChan)
			close(tweetChannel)
			return
		case t := <-tweetChannel:
			if t.fetched {
				storeTweet(t)
			} else {
				go fetch(t, tweetChannel)
			}
		case <-tickChan:
			fmt.Println("Printing state ...")
			showCollectedData()
		}
	}
}

func fetch(t *urlFetch, c chan *urlFetch) {
	resp, err := client.Get(t.url)
	if err != nil {
		//panic(err)
		fmt.Printf("http.Get => %v\n", err.Error())
	} else {

		t.url = resp.Request.URL.String()
		//fmt.Println(t.url)
		t.fetched = true
		c <- t
	}
}

func storeTweet(t *urlFetch) {
	//fmt.Println("Storeing: %#v", t.url)
	if val, ok := urls[t.url]; ok {
		val.count++
	} else {
		urls[t.url] = &url{count: 0, created: time.Now(), url: t.url}
	}
}

// Set twitter stream handlers
func attachMessageHandlers(stream *twitter.Stream, c chan *urlFetch) {
	demux := twitter.NewSwitchDemux()

	demux.Tweet = func(tweet *twitter.Tweet) {
		//fmt.Println("in")
		if tweet.Entities != nil && len(tweet.Entities.Urls) != 0 {
			for _, u := range tweet.Entities.Urls {
				//fmt.Println("Twitter url %#v", u.URL)
				//stringUrl, _ := u.(string)
				c <- NewUrlFetch(u.ExpandedURL)
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

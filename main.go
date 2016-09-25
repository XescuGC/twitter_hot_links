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

	"./config"
	"./schema"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/golang/groupcache/lru"
)

type url struct {
	count   int
	created time.Time
	score   float64
	url     string
}

func (u *url) CalculateScore() {
	power := math.Pow(time.Now().Sub(u.created).Minutes(), 1.8)
	u.score = float64(u.count) / power
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

var (
	urls    = make(map[string]*url)
	timeout = time.Duration(10 * time.Second)
	client  = http.Client{
		Timeout: timeout,
	}
	cache = lru.New(300)
)

func main() {
	stream := openTwitterStream()

	tickChan := time.NewTicker(time.Minute * 1).C
	flushChan := time.NewTicker(time.Minute * 1).C
	signalChannel := make(chan os.Signal)
	tweetChannel := make(chan *urlFetch)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	attachMessageHandlers(stream, tweetChannel)

	for {
		select {
		case <-flushChan:
			fmt.Println("Flushing ...")
			go flush()
		case <-signalChannel:
			fmt.Println("Stoping ...")
			stream.Stop()
			showCollectedData()
			//close(tickChan)
			close(tweetChannel)
			return
		case t := <-tweetChannel:
			if _, ok := urls[t.url]; ok || t.fetched {
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

func flush() {
	for _, v := range urls {
		v.CalculateScore()
		t := schema.NewTweetUrl(v.url, v.count, v.created, v.score)
		err := t.Save()
		if err != nil {
			panic(err)
		}
	}
}

func fetch(t *urlFetch, c chan *urlFetch) {
	defer func() {
		if t.fetched {
			c <- t
		}
	}()
	if value, ok := cache.Get(t.url); ok {
		u := value.(string)
		t.url = u
		t.fetched = true
	} else {
		resp, err := client.Get(t.url)
		if err != nil {
			fmt.Printf("http.Get => %v\n", err.Error())
		} else {
			defer resp.Body.Close() // It fixes an error with http
			u := resp.Request.URL.String()
			cache.Add(t.url, u)
			t.fetched = true
			t.url = u
		}
	}
}

func storeTweet(t *urlFetch) {
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
		if tweet.Entities != nil && len(tweet.Entities.Urls) != 0 {
			for _, u := range tweet.Entities.Urls {
				c <- NewUrlFetch(u.ExpandedURL)
			}
		}
	}

	go demux.HandleChan(stream.Messages)
}

func openTwitterStream() *twitter.Stream {
	c := config.Twitter
	oauthConfig := oauth1.NewConfig(c.ConsumerKey, c.ConsumerSecret)
	token := oauth1.NewToken(c.AccessToken, c.AccessSecret)
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

func scoreUrls() []url {
	urlsByScore := make([]url, 1)
	for _, v := range urls {
		if v.count > 1 {
			v.CalculateScore()
			urlsByScore = append(urlsByScore, *v)
		}
	}
	return urlsByScore
}

func showCollectedData() {
	urlsByScore := scoreUrls()
	sort.Sort(UrlsByScore(urlsByScore))
	for i, v := range urlsByScore {
		if i == 100 {
			break
		}
		fmt.Printf("%f %d %s\n", v.score, v.count, v.url)
	}
}

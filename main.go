package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"./brain"
	"./config"
	"./logger"
	"./schema"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/golang/groupcache/lru"
)

var (
	timeout    = time.Duration(60 * time.Second)
	httpClient = http.Client{
		Timeout: timeout,
	}
	urlRedirCache  = lru.New(300)
	numberOfWokers = 30
	mutex          = &sync.Mutex{}
)

func main() {
	logFile := logger.NewLogFile("hot-links.log")
	streamLog := logFile.WithNamespace("Stream")
	brainLog := logFile.WithNamespace("Brain")
	resolverLog := logFile.WithNamespace("Resolver")
	storeLog := logFile.WithNamespace("Store")

	stream := openTwitterStream(streamLog)

	tickChan := time.NewTicker(time.Minute * 1).C
	storeChan := time.NewTicker(time.Minute * 5).C
	signalChannel := make(chan os.Signal)
	tweetChannel := make(chan *brain.UrlFetch)
	jobsChannel := make(chan *brain.UrlFetch)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	attachMessageHandlers(stream, tweetChannel)

	for i := 1; i < numberOfWokers; i++ {
		go fetchWorker(jobsChannel, tweetChannel, resolverLog)
	}

	for {
		select {
		case <-storeChan:
			go flush(storeLog)

		case t := <-tweetChannel:
			if t.Fetched || brain.Knows(t.Url) {
				brain.Store(t)
			} else {
				go func() { jobsChannel <- t }()
			}

		case <-tickChan:
			brain.Dump(brainLog)

		case <-signalChannel:
			log.Println("Stoping ...")
			stream.Stop()
			//Ensure to save brain state
			close(tweetChannel)
			return
		}
	}
}

func flush(log *log.Logger) {
	log.Println("Storing...")
	for _, v := range brain.Urls() {
		v.CalculateScore()
		t := schema.NewTweetUrl(v.Url, v.Count, v.Created, v.Score)
		err := t.Save()
		if err != nil {
			panic(err)
		}
	}
}

// Resolve redirect and send the real url to the "brain"
func fetchWorker(jobs <-chan *brain.UrlFetch, c chan<- *brain.UrlFetch, log *log.Logger) {
	for j := range jobs {
		//mutex.Lock()
		value, ok := urlRedirCache.Get(j.Url)
		//mutex.Unlock()
		if ok {
			u := value.(string)
			j.Url = u
			j.Fetched = true
			c <- j
		} else {
			resp, err := httpClient.Get(j.Url)
			if err != nil {
				log.Printf("%v\n", err.Error())
			} else {
				defer resp.Body.Close() // It fixes an error with http
				u := resp.Request.URL.String()
				mutex.Lock()
				urlRedirCache.Add(j.Url, u)
				mutex.Unlock()
				j.Fetched = true
				j.Url = u
				c <- j
			}
		}
	}
}

// Set twitter stream handlers
func attachMessageHandlers(stream *twitter.Stream, c chan *brain.UrlFetch) {
	demux := twitter.NewSwitchDemux()

	demux.Tweet = func(tweet *twitter.Tweet) {
		if tweet.Entities != nil && len(tweet.Entities.Urls) != 0 {
			for _, u := range tweet.Entities.Urls {
				c <- brain.NewUrlFetch(u.ExpandedURL)
			}
		}
	}

	go demux.HandleChan(stream.Messages)
}

func openTwitterStream(log *log.Logger) *twitter.Stream {
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

	return stream
}

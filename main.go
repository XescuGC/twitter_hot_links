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
	"./logger"
	"./schema"
	"./stream"
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

	tickChan := time.NewTicker(time.Minute * 1).C
	storeChan := time.NewTicker(time.Minute * 5).C
	signalChannel := make(chan os.Signal)
	tweetChannel := make(chan *brain.UrlFetch)
	jobsChannel := make(chan *brain.UrlFetch)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	stream := stream.NewStream(tweetChannel, streamLog)
	stream.Start()

	//attachMessageHandlers(stream, tweetChannel)

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
			//TODO: Ensure to save brain state
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
				log.Printf("%v for %#v\n", err.Error(), j)
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

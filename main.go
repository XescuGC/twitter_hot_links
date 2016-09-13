package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

//var (
//c redis.Conn
//)

var config struct {
	ConsumerKey    string `json:"consumer_key"`
	ConsumerSecret string `json:"consumer_secret"`
	AccessToken    string `json:"access_token"`
	AccessSecret   string `json:"access_secret"`
}

func main() {

	configFile, err := os.Open("config.json")
	if err != nil {
		fmt.Println("opening config file", err.Error())
	}

	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&config); err != nil {
		fmt.Println("parsing config file", err.Error())
	}

	//c, _ = redis.Dial("tcp", ":6379")
	//if err != nil {
	//log.Println("Redis error: %#v", err)
	//}
	//defer c.Close()

	log.Println("Config: %#v", config)
	oauthConfig := oauth1.NewConfig(config.ConsumerKey, config.ConsumerSecret)
	token := oauth1.NewToken(config.AccessToken, config.AccessSecret)

	// OAuth1 http.Client will automatically authorize Requests
	httpClient := oauthConfig.Client(oauth1.NoContext, token)

	// Twitter Client
	client := twitter.NewClient(httpClient)

	// Convenience Demux demultiplexed stream messages
	demux := twitter.NewSwitchDemux()
	demux.Tweet = func(tweet *twitter.Tweet) {
		//if len(tweet.Entities.Media) != 0 {
		for _, url := range tweet.Entities.Urls {
			//c.Do("MULTI")
			//c.Do("INCR", "twl:"+url.ExpandedURL)
			//c.Do("EXPIRE", "twl:"+url.ExpandedURL, 3600)
			//c.Do("SADD", "twl:urls", url.ExpandedURL)
			//c.Do("EXEC")
			fmt.Printf("URL => %#v\n", url.ExpandedURL)
		}
		//}
		fmt.Println(tweet.Text)
		fmt.Println()
	}
	//demux.DM = func(dm *twitter.DirectMessage) {
	//fmt.Println(dm.SenderID)
	//}
	//demux.Event = func(event *twitter.Event) {
	//fmt.Printf("%#v\n", event)
	//}

	fmt.Println("Starting Stream...")

	// FILTER
	streamParams := &twitter.StreamSampleParams{
		StallWarnings: twitter.Bool(true),
	}
	stream, err := client.Streams.Sample(streamParams)
	if err != nil {
		log.Fatal(err)
	}

	// Receive messages until stopped or stream quits
	go demux.HandleChan(stream.Messages)

	// Wait for SIGINT and SIGTERM (HIT CTRL-C)
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	//defer func() {
	//<-ch
	log.Println(<-ch)
	fmt.Println("Stopping Stream...")
	//}()

	stream.Stop()

	//http.HandleFunc("/", handlerHome)
	//http.ListenAndServe(":3001", nil)
}

//type URL struct {
//Url, Count string
//}

//func handlerHome(w http.ResponseWriter, r *http.Request) {
//results, _ := c.Do("SORT", "twl:urls", "BY", "twl:*", "LIMIT", "0", "10")
//casted_results := results.([]interface{})
////urls := make([]string)
//urls := []URL{}

//for _, url := range casted_results {
//var u string
//url := url.([]uint8)
//for _, str := range url {
////log.Printf("INT => %#v - %#v", str, int(str))
////log.Printf("%#v", strconv.Itoa(int(str)))
////u += strconv.Itoa(int(str))
//u += string(str)
//}
//count, _ := c.Do("GET", "twl:"+u)
//log.Printf("%#v", count)
//byte_count, _ := count.([]byte)
//log.Printf("%#v", string(byte_count))
//urls = append(urls, URL{u, string(byte_count)})
////log.Printf("%#v", u)
//}

////log.Printf("%#v", urls)
////log.Printf("%#v", u[0])
////log.Println(reflect.TypeOf(urls))

//t := template.New("base.html")
//pwd, _ := os.Getwd()
//t = template.Must(t.ParseFiles(
//path.Join(pwd, "templates/base.html"),
//filepath.Join(pwd, "templates", "home.html"),
//))

//t.Execute(w, urls)
//}

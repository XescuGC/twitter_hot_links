package main

import (
	"fmt"

	"../schema"
	"github.com/xescugc/twitter_hot_links/config" // TODO: Check why one import works and the other no
	//"../config"
)

var mapping = `{
	"mappings":{` + schema.TweetUrlMapping() + `} }`

func main() {
	_, _ = config.ElasticSearch.DeleteIndex("trending_url").Do()
	createIndex, err := config.ElasticSearch.CreateIndex("trending_url").BodyString(mapping).Do()

	if err != nil {
		panic(err)
	}

	if !createIndex.Acknowledged {
		panic("Not Ack")
	}

	fmt.Println("Index Created")
}

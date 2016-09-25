package schema

import (
	"time"

	"../config"
)

type TweetUrl struct {
	Count   int       `json:"count"`
	Created time.Time `jons:"created"`
	Score   float64   `json:"score"`
	Url     string    `json:"url"`
}

func NewTweetUrl(url string, count int, created time.Time, score float64) *TweetUrl {
	return &TweetUrl{Url: url, Created: created, Count: count, Score: score}
}

func (t *TweetUrl) Save() error {
	_, err := config.ElasticSearch.Index().Index("trending_url").Type("tweet_url").BodyJson(t).Do()
	return err
}

func TweetUrlMapping() string {
	return `"tweet_url": {
		"properties": {
			"count":   { "type": "integer" },
			"created": { "type": "date" },
			"score":   { "type": "float" },
			"url":     { "type": "string" }
		}
	}`
}

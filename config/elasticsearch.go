package config

import "github.com/olivere/elastic"

var ElasticSearch *elastic.Client

func init() {
	var err error
	ElasticSearch, err = elastic.NewClient()

	if err != nil {
		panic(err)
	}
}

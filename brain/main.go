package brain

import (
	"log"
	"math"
	"sort"
	"time"
)

type UrlFetch struct {
	Url     string
	Fetched bool
}

type url struct {
	Count   int
	Created time.Time
	Score   float64
	Url     string
}

type UrlsByScore []url

func (u UrlsByScore) Len() int           { return len(u) }
func (u UrlsByScore) Swap(i, j int)      { u[i], u[j] = u[j], u[i] }
func (u UrlsByScore) Less(i, j int) bool { return u[i].Score > u[j].Score }

var urls = make(map[string]*url)

// Remove this shit after factoring all this into a proper struct
func Urls() map[string]*url { return urls }

func NewUrlFetch(url string) *UrlFetch {
	return &UrlFetch{url, false}
}

func (u *url) CalculateScore() {
	power := math.Pow(time.Now().Sub(u.Created).Minutes(), 1.8)
	u.Score = float64(u.Count) / power
}

func Store(t *UrlFetch) {
	if val, ok := Get(t.Url); ok {
		val.Count++
	} else {
		urls[t.Url] = &url{Count: 0, Created: time.Now(), Url: t.Url}
	}
}

func Get(url string) (*url, bool) {
	v, ok := urls[url]
	return v, ok
}
func Knows(url string) bool {
	_, ok := Get(url)
	return ok
}

func scoreUrls() []url {
	urlsByScore := make([]url, 1)
	for _, v := range urls {
		if v.Count > 1 {
			v.CalculateScore()
			urlsByScore = append(urlsByScore, *v)
		}
	}
	return urlsByScore
}

func Dump(log *log.Logger) {
	urlsByScore := scoreUrls()
	sort.Sort(UrlsByScore(urlsByScore))
	for i, v := range urlsByScore {
		if i == 100 {
			break
		}
		log.Printf("%f %d %s\n", v.Score, v.Count, v.Url)
	}
}

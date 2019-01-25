package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/nohtaray/ety-ranking/pkg"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"sync"
)

type IM int

const (
	Roma    IM = 0
	Kana    IM = 1
	English IM = 2
)

func (im IM) String() string {
	return map[IM]string{
		Roma:    "ローマ字",
		Kana:    "かな",
		English: "英語",
	}[im]
}

func getWordNameParamT(im IM) string {
	return map[IM]string{
		Roma:    "trysc.trysc.trysc.std.0",
		Kana:    "trysc.trysc.trysc.kana.1",
		English: "trysc.trysc.trysc.std.2",
	}[im]
}

func fetchChampionScore(im IM) (user string, score int) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatal(err)
	}
	client := &http.Client{Jar: jar}
	// 先週
	// res, err := client.Get(fmt.Sprintf("https://www.e-typing.ne.jp/ranking/index.asp?im=%d&sc=trysc&ct=-1", im))
	// 今週
	res, err := client.Get(fmt.Sprintf("https://www.e-typing.ne.jp/ranking/index.asp?im=%d&sc=trysc", im))
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	if doc.Find("#ranking .ranking>li[class!=head]>.rank").First().Text() != "1位" {
		log.Fatal("HTML がおかしいです。")
	}
	scoreStr := doc.Find("#ranking .ranking>li[class!=head]>.score").First().Text()
	score, err = strconv.Atoi(scoreStr)
	if err != nil {
		log.Fatal(err)
	}
	user = doc.Find("#ranking .ranking>li[class!=head]>.user").First().Text()
	return user, score
}

func fetchWordName(im IM) (wordName string) {
	client := &http.Client{}
	res, err := client.PostForm("https://www.e-typing.ne.jp/parts/cgilib/get_typing_setting.asp", url.Values{
		"d_id": []string{""},
		"t":    []string{getWordNameParamT(im)},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	wordName = doc.Find("title").Text()
	return
}

func tweetRanking(im IM) {
	scoreChan := make(chan int, 1)
	userChan := make(chan string, 1)
	wordChan := make(chan string, 1)
	go func() {
		u, s := fetchChampionScore(im)
		userChan <- u
		scoreChan <- s
	}()
	go func() {
		w := fetchWordName(im)
		wordChan <- w
	}()

	word := <-wordChan
	user := <-userChan
	score := <-scoreChan
	message := fmt.Sprintf("今週の腕試し（%s）の1位は %dpt で %s さんです。（%s）",
		im, score, user, word)
	pkg.Tweet(message)
}

func main() {
	wg := sync.WaitGroup{}
	ims := []IM{Roma, Kana, English}
	for _, im := range ims {
		wg.Add(1)
		go func(im2 IM) {
			defer wg.Done()
			tweetRanking(im2)
		}(im)
	}
	wg.Wait()
}

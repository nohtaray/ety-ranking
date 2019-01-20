package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
)

func GetChampionScore() (user string, score int) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatal(err)
	}
	client := &http.Client{Jar: jar}
	// 先週
	// res, err := client.Get("https://www.e-typing.ne.jp/ranking/index.asp?im=1&sc=trysc&ct=-1")
	// 今週
	res, err := client.Get("https://www.e-typing.ne.jp/ranking/index.asp?im=1&sc=trysc")
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

func GetWordName() (wordName string) {
	client := &http.Client{}
	res, err := client.PostForm("https://www.e-typing.ne.jp/parts/cgilib/get_typing_setting.asp", url.Values{
		"d_id": []string{""},
		"t":    []string{"trysc.trysc.trysc.std.0"},
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

func main() {
	scoreChan := make(chan int, 1)
	userChan := make(chan string, 1)
	wordChan := make(chan string, 1)
	go func() {
		u, s := GetChampionScore()
		userChan <- u
		scoreChan <- s
	}()
	go func() {
		w := GetWordName()
		wordChan <- w
	}()
	fmt.Println(<-wordChan)
	fmt.Println(<-userChan, <-scoreChan)
}

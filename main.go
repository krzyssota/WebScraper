package main

import (
	"concurrent_webscraper/formatter"
	"concurrent_webscraper/webscraper"
	"fmt"
	"runtime"
)

func main() {
	rp := formatter.NewResultPresenter("formatter/stop_words.txt", 1)

	// prepare urls to scrape
	var urls []string
	websensaBlogBaseUrl := "https://websensa.com/en/blog-en/page/"
	pagesCount := 5
	for i := 1; i <= pagesCount; i++ {
		urls = append(urls, websensaBlogBaseUrl+fmt.Sprint(i))
	}

	// create web scraper
	urlChannel := make(chan string, len(urls))
	urlsSent := make(chan int)
	// defer closing owned channels
	defer close(urlChannel)
	defer close(urlsSent)
	resultChannel := make(chan webscraper.ResultWithError, len(urls))
	aggregatedResultsChannel := make(chan map[string]int)
	ws := webscraper.NewWebScraper(urlChannel, resultChannel, aggregatedResultsChannel, urlsSent, runtime.GOMAXPROCS(0))
	go ws.Run()

	// send urls and number of urls sent
	for _, url := range urls {
		urlChannel <- url
	}
	urlsSent <- len(urls)

	// receive partial results
	for res := range resultChannel {
		if err := res.Err; err != nil {
			fmt.Printf("Scraping %v resulted in an error: %v\n", res.Result.Url, err)
		} else {
			fmt.Printf(fmt.Sprintf("words for %v", res.Result.Url)+" %v\n", rp.SortFilterWords(res.Result.Words))
		}
	}

	// receive aggregated results
	aggregatedResult := <-aggregatedResultsChannel
	fmt.Printf("\nAggregated results:\n%v", rp.SortFilterWords(aggregatedResult))
}

package webscraper

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

func createWebScraper(urls []string, gomaxproc int) (ws *WebScraper, urlChannel chan string, urlsSent chan int, resultChannel chan ResultWithError, aggregatedResultsChannel chan map[string]int) {
	urlChannel = make(chan string, len(urls))
	urlsSent = make(chan int)
	resultChannel = make(chan ResultWithError, len(urls))
	aggregatedResultsChannel = make(chan map[string]int)
	ws = NewWebScraper(urlChannel, resultChannel, aggregatedResultsChannel, urlsSent, gomaxproc)
	return ws, urlChannel, urlsSent, resultChannel, aggregatedResultsChannel
}

func timer(text string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s: %v\n", text, time.Since(start))
	}
}

func runWithTimer(text string, ws *WebScraper, urls []string, urlChannel chan string, urlsSent chan int, resultChannel chan ResultWithError, aggregatedResultsChannel chan map[string]int) {
	defer timer(text)()
	go ws.Run()
	for _, url := range urls {
		urlChannel <- url
	}
	urlsSent <- len(urls)
	// drain channels
	for len(resultChannel) > 0 {
		<-resultChannel
	}
	<-aggregatedResultsChannel
	close(urlChannel)
	close(urlsSent)

}

// test scraper using various allowed amount of goroutines running in parallel
func TestVariousGoroutineLimits(t *testing.T) {
	var urls []string
	okopressBaseUrl := "https://oko.press/temat/wybory?page="
	pagesCount := 100
	for i := 1; i <= pagesCount; i++ {
		urls = append(urls, okopressBaseUrl+fmt.Sprint(i))
	}
	wsMaxProc, ch1, ch2, ch3, ch4 := createWebScraper(urls, runtime.GOMAXPROCS(0))
	ws1, ch5, ch6, ch7, ch8 := createWebScraper(urls, 1)
	wsMillion, ch9, ch10, ch11, ch12 := createWebScraper(urls, 1e9)

	runWithTimer("Web scraping 100 pages with 1 goroutine took", ws1, urls, ch5, ch6, ch7, ch8)
	runWithTimer("Web scraping 100 pages with 1_000_000 goroutines took", wsMillion, urls, ch9, ch10, ch11, ch12)
	runWithTimer(fmt.Sprintf("Web scraping 100 pages with runtime.GOMAXPROCS(0)=%d goroutines took", runtime.GOMAXPROCS(0)), wsMaxProc, urls, ch1, ch2, ch3, ch4)
}

package webscraper

import (
	"fmt"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"strings"
	"unicode"
)

type Result struct {
	Url   string
	Words map[string]int
}

type ResultWithError struct {
	Result Result
	Err    error
}

type WebScraper struct {
	urlChannel               <-chan string
	resultChannel            chan<- ResultWithError
	computedResultChannel    chan ResultWithError
	aggregatedResultsChannel chan map[string]int
	cache                    map[string]Result
	urlsSent                 chan int
	sem                      chan int
}

func (w *WebScraper) countWordsInTexts(url string, texts []string) {
	wordCount := make(map[string]int)
	for _, text := range texts {
		// count words in texts that are not whitespace
		if isNotAllWhitespace(text) {
			// divide text by characters that aren't letters
			words := strings.FieldsFunc(text, func(r rune) bool { return !unicode.IsLetter(r) })
			for _, word := range words {
				wordCount[word] += 1
			}
		}
	}
	// pass computed counts to the primary goroutine
	w.computedResultChannel <- ResultWithError{Result{Url: url, Words: wordCount}, nil}
}

// traverse html tree
func (w *WebScraper) parseResponse(url string, responseBody io.ReadCloser) {
	tokenizer := html.NewTokenizer(responseBody)
	// bools keeping track of being within certain html tags
	inBody := false
	// text within those tags is not visible on the website
	inInvisibleTags := map[string]bool{"script": false, "style": false, "noscript": false} // map keeping track if current token is within one of the invisible tags
	texts := make([]string, 0)
	for {
		tokenType := tokenizer.Next()
		token := tokenizer.Token()
		switch tokenType {
		case html.ErrorToken:
			if err := tokenizer.Err(); err != io.EOF {
				fmt.Printf("error token type: %v\n", tokenType)
				w.computedResultChannel <- ResultWithError{Result{Url: url, Words: nil}, err}
			} else { // end of the page, can proceed with counting words
				w.countWordsInTexts(url, texts)
			}
			return
		// certain html tags started/ended
		case html.StartTagToken, html.EndTagToken:
			if token.Data == "body" {
				inBody = !inBody
			} else if tokenWithinInvisibleTag, invisibleTag := inInvisibleTags[token.Data]; invisibleTag { // token is an invisible tag
				inInvisibleTags[token.Data] = !tokenWithinInvisibleTag
			}
		case html.TextToken:
			// text visible on the website
			if !inOneOfInvisibleTags(inInvisibleTags) && inBody {
				texts = append(texts, token.Data)
			}
		}
	}
}

func (w *WebScraper) scrapeUrl(url string) {
	// GET url
	response, err := http.Get(url)
	if err != nil {
		w.computedResultChannel <- ResultWithError{Result: Result{url, nil}, Err: err}
		return
	}
	// defer closing body and sending first error that occurred in scrapeUrl
	defer func(Body io.ReadCloser) {
		if err != nil {
			w.computedResultChannel <- ResultWithError{Result: Result{url, nil}, Err: err}
		}
		if bodyCloseErr := Body.Close(); err == nil && bodyCloseErr != nil {
			w.computedResultChannel <- ResultWithError{Result: Result{url, nil}, Err: err}
		}
	}(response.Body)

	if response.StatusCode != http.StatusOK {
		err = fmt.Errorf("%v responded with %v", url, response.StatusCode)
		return
	}
	if !strings.HasPrefix(response.Header.Get("Content-Type"), "text") {
		err = fmt.Errorf("%v responded with non text-like Content-type: %v", url, response.Header.Get("Content-Type"))
		return
	}

	w.parseResponse(url, response.Body)
}

func NewWebScraper(urlChannel chan string, resultChannel chan ResultWithError, aggregatedResultsChannel chan map[string]int, urlsSent chan int, maxRunningGoroutines int) *WebScraper {
	w := new(WebScraper)
	w.urlChannel = urlChannel
	w.resultChannel = resultChannel
	w.cache = make(map[string]Result)
	w.computedResultChannel = make(chan ResultWithError, 10)
	w.urlsSent = urlsSent
	w.sem = make(chan int, maxRunningGoroutines)
	w.aggregatedResultsChannel = aggregatedResultsChannel
	return w
}

func (w *WebScraper) Run() {
	// defer closing owned channels
	defer close(w.computedResultChannel)
	defer close(w.aggregatedResultsChannel)

	urlsSentCount := -1
	aggregatedResults := make(map[string]int)
	// work until all requested urls were scraped and their word counts were sent
	for responsesSent := 0; urlsSentCount == -1 || responsesSent < urlsSentCount; {
		select {
		case url := <-w.urlChannel:
			if res, cached := w.cache[url]; cached { // return result if cached
				w.resultChannel <- ResultWithError{res, nil}
				responsesSent += 1
			} else { // scrape received url
				/* limited number of concurrently RUNNING goroutines
				more goroutines can be created, but they will be blocked
				and consuming very little resources */
				go func() {
					defer func() { <-w.sem }()
					w.sem <- 1
					w.scrapeUrl(url)
				}()
			}
		case resultWithErr := <-w.computedResultChannel: // receive counts computed by a goroutine
			// save to cache and send to webscraper's user
			w.cache[resultWithErr.Result.Url] = resultWithErr.Result
			w.resultChannel <- resultWithErr
			responsesSent += 1
			// aggregate new result
			for word, count := range resultWithErr.Result.Words {
				aggregatedResults[word] += count
			}
		case urlsSentCount = <-w.urlsSent: // receive how many urls are to be scraped
		}
	}
	close(w.resultChannel) // webscraper's user can now wait for the aggregated results
	w.aggregatedResultsChannel <- aggregatedResults
}

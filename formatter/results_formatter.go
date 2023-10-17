package formatter

import (
	"bufio"
	"fmt"
	"os"
	"sort"
)

type ResultPresenter struct {
	stopWords map[string]bool
	cutoff    int
}

// structure to sort the results
type wordWithCount struct {
	word  string
	count int
}

func (wc wordWithCount) String() string {
	return fmt.Sprintf("%v->%d", wc.word, wc.count)
}

// filter stop words and words that do not appear very often
func (p *ResultPresenter) SortFilterWords(wordCount map[string]int) []wordWithCount {
	wordCountSlice := make([]wordWithCount, 0)
	for word, count := range wordCount {
		if _, isStopWord := p.stopWords[word]; !isStopWord && count >= p.cutoff {
			wordCountSlice = append(wordCountSlice, wordWithCount{word, count})
		}
	}
	sort.Slice(wordCountSlice, func(i, j int) bool {
		return wordCountSlice[i].count >= wordCountSlice[j].count
	}) // sort words by descending count
	return wordCountSlice
}

func getStopWords(filename string) (map[string]bool, error) {
	stopWords := make(map[string]bool)
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return stopWords, err
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		stopWords[scanner.Text()] = true
	}
	if err := scanner.Err(); err != nil {
		return stopWords, err
	}
	return stopWords, nil
}

func NewResultPresenter(stopWordsFilename string, minimalCount int) *ResultPresenter {
	stopWords, err := getStopWords(stopWordsFilename)
	if err != nil {
		fmt.Println("Could not get stop words because of error:", err)
	}
	return &ResultPresenter{stopWords: stopWords, cutoff: minimalCount}
}

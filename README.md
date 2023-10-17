# Concurrent Web Scraper with Caching
Task description (in polish) in task_description.png file.

### Project structure
This repository contains a *concurrent_webscraper* module with *main*, *webscraper* and *formatter* packages.
- main
  - main.go contains a simple usage scenario.
- webscraper
  - webscraper.go is the implementation of the web scraper. 
  - webscraper_utils.go contains some util functions that web scraper uses.
  - webscraper_test.go contains simple performance test.
- formatter
  - results_formatter.go contains some functions to present scraped data in a more readable way.
  - stop_words.txt contains a list of stop words (like "or", "the", "in"), in polish and english, that will be excluded from the presented data.\
  I took it from:
    - https://github.com/bieli/stopwords/blob/master/polish.stopwords.txt
    - https://gist.github.com/sebleier/554280


### Build
Module is built with:
```bash
go build
```

### Run
main.go is run with:
```bash
go run main.go
```

### Test
Performance test is run with:
```bash
cd webscraper && go test
```

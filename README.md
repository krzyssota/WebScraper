# Concurrent Web Scraper with Caching
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
 - stop_words.txt contains a list of stop words (like "or", "the", "in") that will be excluded from the presented data.


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
performance test is run with:
```bash
cd webscraper && go test
```

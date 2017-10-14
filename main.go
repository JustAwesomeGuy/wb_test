/*
Package main reads urls from stdIn and writes to stdOut the count of "Go" substrings for every url response
and the total count. It uses semaphore with capacity set to 5 by default to control number of simultaneous goroutines.
*/
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
)

// urlResult contains name of url, number of "Go" substrings in response of the url and occurred error.
type urlResult struct {
	url   string
	count int
	error error
}

// allResults is storage for urlResult structs. Safe for concurrent usage.
type allResults struct {
	total   int
	results []*urlResult
	mux     sync.Mutex
}

func (counter *urlResult) String() string {
	if counter.error != nil {
		return fmt.Sprintf("%s: Error '%s'", counter.url, counter.error.Error())
	} else {
		return fmt.Sprintf("Count for %s: %d", counter.url, counter.count)
	}
}

// addUrlResult add given urlResult to allResult struct. It is safe to use concurrently.
func (counter *allResults) addUrlResult(res *urlResult) {
	counter.mux.Lock()
	defer counter.mux.Unlock()
	counter.results = append(counter.results, res)
	counter.total += res.count
}

func main() {
	var k = flag.Int("k", 5, "Number of simultaneous goroutines")
	flag.Parse()
	sem := make(chan bool, *k)
	goCount := &allResults{}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if url := scanner.Text(); url != "" {
			sem <- true
			go func() {
				defer func() { <-sem }()
				res := handleUrl(url)
				goCount.addUrlResult(res)
			}()
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
	for _, url := range goCount.results {
		fmt.Println(url)
	}
	fmt.Println("Total:", goCount.total)
}

// handleUrl returns new urlResult for given url with counted number of "Go" substring in response
func handleUrl(url string) *urlResult {
	resp, err := http.Get(url)
	if err != nil {
		return &urlResult{url: url, error: err}
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &urlResult{url: url, error: err}
	}
	count := bytes.Count(body, []byte("Go"))
	return &urlResult{url: url, count: count}
}

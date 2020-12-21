package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var idRegexp = regexp.MustCompile(`<a href="/stables/(?P<ID>[0-9a-zA-Z]+)">`)
var url = "https://portal.mda.maryland.gov/stables"

func main() {
	WriteIdList()
	/*
		resp, err := http.Get(url)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fetch: %v\n", err)
			os.Exit(1)
		}
		b, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "fetch: reading %s: %v\n", url, err)
			os.Exit(1)
		}
		fmt.Printf("%s", b)
	*/
}

func WriteIdList() {
	ch := make(chan string)
	// Start 10 goroutines which will work over pages getting IDs
	for i := 1; i < 11; i++ {
		go processIdPage(i, ch)
	}

	idFile, err := os.Create("./ids.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening id file %v\n", err)
	}

	doneCount := 0
	for {
		id := <-ch
		if id == "" {
			doneCount++
		} else {
			_, err := idFile.WriteString(fmt.Sprintf("%s\n", id))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error writing %s%v\n", id, err)
			}
		}
		if doneCount == 10 {
			close(ch)
			break
		}
	}

	err = idFile.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error closing id file %v\n", err)
	}
	fmt.Printf("Done!\n")
}

func processIdPage(page int, ch chan<- string) {
	for {
		if page > 100 {
			fmt.Fprintf(os.Stderr, "Emergency stopgap at page %d\n", page)
			break
		}
		fullUrl := strings.Join([]string{url, "?page=", strconv.Itoa(page)}, "")
		fmt.Printf("Getting %s\n", fullUrl)
		resp, err := http.Get(fullUrl)
		if err != nil {
			fmt.Fprintf(os.Stderr, "GET err on page %d: %v\n", page, err)
			continue
		}
		b, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ReadAll err on page %d: %v\n", page, err)
			continue
		}
		ids := parseIdsWithReadAll(b)
		if len(ids) < 1 {
			fmt.Fprintf(os.Stderr, "Found no IDs on page %d\n", page)
			break
		}
		for _, id := range ids {
			//fmt.Printf("\tID: %s\n", id)
			ch <- id
		}
		page += 10
	}
	ch <- ""
}

func parseIdsWithReadAll(data []byte) []string {
	// Assumes that we're going to call ioutil.ReadAll on resp.Body
	matches := idRegexp.FindAllSubmatch(data, -1)
	var ids []string
	for _, match := range matches {
		ids = append(ids, fmt.Sprintf("%s", match[1]))
	}
	return ids
}

/* Notes:
1. Let's plan to go in groups of 10 over the /stables?page=x, and stop when we
   get a 404.
2. For each thing on page with pattern like
   <a href="/stables/5ce71cf84d7ef91b0c22a52e">, get that id.
3. I guess we could collect a list of IDs, then move on to page parsing. Let's
   also save all those IDs in a file.
4. For each page, find the "article".
5. Everything in <h4> tags in the article is a column.
   Unfortunately Address looks like:
   <h4>Address:</h4>
   <a href="https://maps.google.com/?q=8900 RACE TRACK ROAD BOWIE, MD, 20715">
    8900 RACE TRACK ROAD<br/>
    BOWIE, MD, 20715<br/>
   </a>
6. To reduce hits, just pull all the article content into a file and we can
   parse that file later.
*/

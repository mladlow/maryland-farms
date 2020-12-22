package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Regexp vars
var (
	idRegexp      = regexp.MustCompile(`<a href="/stables/(?P<ID>[0-9a-zA-Z]+)">`)
	nameRegexp    = regexp.MustCompile(`(?s)<h1>(?P<NAME>.+?)<`)
	articleRegexp = regexp.MustCompile(`(?s)(?P<ARTICLE><article>.+?</article>)`)
	phoneRegexp   = regexp.MustCompile(`Tel: (?P<PHONE>[0-9\-\(\)]+)<`)
	siteRegexp    = regexp.MustCompile(`Website: (?P<SITE>.+?)<`)
	addressRegexp = regexp.MustCompile(`<a href="https://maps.google.com/\?q=(?P<ADDRESS>.+?)">`)
)
var url = "https://portal.mda.maryland.gov/stables"

func main() {
	WriteIdList()
}

func WriteIdList() {
	// This function parses the paginated list of stables and extracts the
	// stable IDs. Stable IDs form the URL of the individual stable pages.
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
	// This function takes a single page containing a list of stables and
	// reads all the IDs off the page.
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
	// This function uses a regexp to get all the IDs off a page read into
	// memory. It assumes that we're going to call ioutil.ReadAll on resp.Body.
	matches := idRegexp.FindAllSubmatch(data, -1)
	var ids []string
	for _, match := range matches {
		ids = append(ids, fmt.Sprintf("%s", match[1]))
	}
	return ids
}

func extractStable(data []byte) (*Stable, error) {
	// This function takes a page with information about a single stable and
	// extracts that information into a struct.

	// Pick out the stable name from the <h1> tags
	matches := nameRegexp.FindAllSubmatch(data, -1)
	if len(matches) != 1 {
		return nil, errors.New(fmt.Sprintf("extractStable: Found %d name(s)\n", len(matches)))
	}
	nameMatch := fmt.Sprintf("%s", matches[0][1])
	nameMatch = strings.TrimSpace(nameMatch)
	stable := Stable{Name: nameMatch}

	// Pull out the article section
	matches = articleRegexp.FindAllSubmatch(data, -1)
	if len(matches) != 1 {
		return nil, errors.New(fmt.Sprintf("extractStable: Found %d article(s)\n", len(matches)))
	}
	article := fmt.Sprintf("%s", matches[0][1])

	// Pull out the address from the article
	strMatches := addressRegexp.FindAllStringSubmatch(article, -1)
	if len(strMatches) != 1 {
		return nil, errors.New(fmt.Sprintf("extractStable: Found %d address(es)\n", len(strMatches)))
	}
	stable.Address = strMatches[0][1]

	// Look for the first website and phone number
	strMatches = phoneRegexp.FindAllStringSubmatch(article, -1)
	if len(strMatches) >= 1 {
		stable.Phone = strMatches[0][1]
	}
	strMatches = siteRegexp.FindAllStringSubmatch(article, -1)
	if len(strMatches) >= 1 {
		stable.Website = strMatches[0][1]
	}
	return &stable, nil
}

type Stable struct {
	Name    string
	Address string
	Phone   string
	Website string
	ID      string
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

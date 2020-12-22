package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// Regexp vars
var (
	idRegexp      = regexp.MustCompile(`<a href="/stables/(?P<ID>[0-9a-zA-Z]+)">`)
	nameRegexp    = regexp.MustCompile(`(?s)<h1>(?P<NAME>.+?)<`)
	articleRegexp = regexp.MustCompile(`(?s)(?P<ARTICLE><article>.+?</article>)`)
	phoneRegexp   = regexp.MustCompile(`Tel: (?P<PHONE>[0-9\-\(\)]+)<`)
	siteRegexp    = regexp.MustCompile(`Website: (?P<SITE>.+?)<`)
	addressRegexp = regexp.MustCompile(`(?s)<a href="https://maps.google.com/\?q=(?P<ADDRESS>.+?)">`)
)

// String vars
var (
	url          = "https://portal.mda.maryland.gov/stables"
	idFileName   = "./ids.txt"
	errFileName  = "./errIds.txt"
	dataFileName = "./data.json"
)

func main() {
	// Use WriteIdList to get a list of stable IDs into ids.txt.
	// Use IterateIdList to use the stable ID list.
	// WriteIdList()
	fileInfo, err := os.Stat(idFileName)
	if os.IsNotExist(err) {
		WriteIdList()
	} else if !fileInfo.IsDir() {
		IterateIdList()
	} else {
		fmt.Fprintf(os.Stderr, "main: Problem with id file\n")
	}
}

func IterateIdList() {
	// Read ids.txt into an array
	// Start 10 goroutines, use atomic int to track index in array
	// Goroutine will fail and add to list of failures if index %9
	// Goroutine will otherwise succeed
	// Write list of failures
	idFile, err := os.Open(idFileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening id file %v\n", err)
		return
	}
	// I read somewhere that this was bad because os.Close can create an error
	// but in such short-lived and non critical software I don't know if I
	// care.
	defer idFile.Close()

	errFile, err := os.Create(errFileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening error file %v\n", err)
		return
	}
	defer errFile.Close()

	dataFile, err := os.Create(dataFileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening data file %v\n", err)
		return
	}
	defer dataFile.Close()

	idCh := make(chan string)
	stableCh := make(chan Stable)

	// Does the order here matter?
	// addIds closes idCh - would not be managable with larger codebase.
	go addIds(idFile, idCh)

	wg := new(sync.WaitGroup)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go processStablePage(idCh, stableCh, wg)
	}

	go func() {
		wg.Wait()
		close(stableCh)
	}()

	for stable := range stableCh {
		if stable.Name == "" {
			errFile.WriteString(fmt.Sprintf("%s\n", stable.ID))
		} else {
			stableJson, err := json.Marshal(stable)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Marshal failed for %s, %v\n", stable.ID, err)
				errFile.WriteString(fmt.Sprintf("%s\n", stable.ID))
				continue
			}
			stableJson = append(stableJson, '\n')
			dataFile.Write(stableJson)
		}
	}
	fmt.Printf("Done!\n")
}

func WriteIdList() {
	// This function parses the paginated list of stables and extracts the
	// stable IDs. Stable IDs form the URL of the individual stable pages.
	ch := make(chan string)
	// Start 10 goroutines which will work over pages getting IDs
	for i := 1; i < 11; i++ {
		go processIdPage(i, ch)
	}

	idFile, err := os.Create(idFileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening id file %v\n", err)
		return
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

func addIds(idFile *os.File, idCh chan<- string) {
	scanner := bufio.NewScanner(idFile)
	for scanner.Scan() {
		idCh <- scanner.Text()
	}
	close(idCh)
}

func processStablePage(idCh <-chan string, stableCh chan<- Stable, wg *sync.WaitGroup) {
	defer wg.Done()

	for id := range idCh {
		stable := Stable{ID: id}

		stableUrl := strings.Join([]string{url, "//", id}, "")
		resp, err := http.Get(stableUrl)
		if err != nil {
			fmt.Fprintf(os.Stderr, "GET err for ID %s: %v\n", id, err)
			stableCh <- stable
			continue
		}
		b, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ReadAll err for ID %s: %v\n", id, err)
			stableCh <- stable
			continue
		}

		// Question here is should we mutate the Stable or return a new Stable?
		// Mutating means we need to make a new Stable to return with the name
		// (which denotes failure and required re-processing) unset.
		err = stable.extractStable(b)
		if err != nil {
			stableCh <- Stable{ID: id}
		} else {
			stableCh <- stable
		}
	}
}

func (stable *Stable) extractStable(data []byte) error {
	// This function takes a page with information about a single stable and
	// extracts that information into a struct. It mutates the input stable.

	// Pick out the stable name from the <h1> tags
	matches := nameRegexp.FindAllSubmatch(data, -1)
	if len(matches) != 1 {
		return errors.New(fmt.Sprintf("extractStable: Found %d name(s)\n", len(matches)))
	}
	nameMatch := fmt.Sprintf("%s", matches[0][1])
	nameMatch = strings.TrimSpace(nameMatch)
	stable.Name = nameMatch

	// Pull out the article section
	matches = articleRegexp.FindAllSubmatch(data, -1)
	if len(matches) != 1 {
		return errors.New(fmt.Sprintf("extractStable: Found %d article(s)\n", len(matches)))
	}
	article := fmt.Sprintf("%s", matches[0][1])

	// Pull out the address from the article
	strMatches := addressRegexp.FindAllStringSubmatch(article, -1)
	if len(strMatches) != 1 {
		return errors.New(fmt.Sprintf("extractStable: Found %d address(es)\n", len(strMatches)))
	}
	stable.Address = strings.ReplaceAll(strMatches[0][1], "\n", " ")

	// Look for the first website and phone number
	strMatches = phoneRegexp.FindAllStringSubmatch(article, -1)
	if len(strMatches) >= 1 {
		stable.Phone = strMatches[0][1]
	}
	strMatches = siteRegexp.FindAllStringSubmatch(article, -1)
	if len(strMatches) >= 1 {
		stable.Website = strMatches[0][1]
	}
	return nil
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

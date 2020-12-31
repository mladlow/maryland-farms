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

/*
Declare and compile a set of global regexp variables to be used while parsing
HTML pages from the Maryland Horse Board's directory.
*/
var (
	idRegexp      = regexp.MustCompile(`<a href="/stables/(?P<ID>[0-9a-zA-Z]+)">`)
	nameRegexp    = regexp.MustCompile(`(?s)<h1>(?P<NAME>.+?)<`)
	articleRegexp = regexp.MustCompile(`(?s)(?P<ARTICLE><article>.+?</article>)`)
	phoneRegexp   = regexp.MustCompile(`Tel: (?P<PHONE>[0-9\-\(\)]+)<`)
	siteRegexp    = regexp.MustCompile(`Website: (?P<SITE>.+?)<`)
	addressRegexp = regexp.MustCompile(`(?s)<a href="https://maps.google.com/\?q=(?P<ADDRESS>.+?)">`)
)

/*
Declare global string variables to be used throughout this code.
*/
var (
	portalUrl    = "https://portal.mda.maryland.gov/stables"
	idFileName   = "./ids.txt"
	errFileName  = "./errIds.txt"
	dataFileName = "./portalData.json"
)

/*
If a list of stable IDs does not exist in this directory, create it by parsing
over the list available on the MHB portal.
If the list does exist, use the IDs to extract information about the stables.
*/
/*
func main() {
	fileInfo, err := os.Stat(idFileName)
	if os.IsNotExist(err) {
		WriteIdList()
	} else if !fileInfo.IsDir() {
		IterateIdList()
	} else {
		fmt.Fprintf(os.Stderr, "main: Problem with id file\n")
	}
}
*/

/*
Read a list of IDs from a text file into a channel. Use a WaitGroup to read
from that channel and send each ID to a function that will GET it and extract
key stable information from the individual stable page.

Finally, this function reads from the WaitGroup channel, marshals Stable structs
to JSON, and writes that JSON to a file.
*/
func IterateIdList() {
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
	// addIds closes idCh - would not be able to mentally track this with larger
	// codebase.
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
			stableJson = append(stableJson, ',', '\n')
			dataFile.Write(stableJson)
		}
	}
	fmt.Printf("Done!\n")
}

/*
WriteIdList iterates over the pages on the MHB portal and extract the stable
IDs. The IDs can later be used to form URLs for individual stable pages.

These IDs are written to a text file, because generating the data for the
actual website is still a pretty manual process.
*/
func WriteIdList() {
	// I didn't use a WaitGroup here because I didn't realize you needed to
	// wait on it in a go routine. Probably should read more about this later.
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

/*
processIdPage GETs a page containing a list of stables and extracts the IDs
that form stable URLs off the page.
*/
func processIdPage(page int, ch chan<- string) {
	for {
		if page > 100 {
			fmt.Fprintf(os.Stderr, "Emergency stopgap at page %d\n", page)
			break
		}
		fullUrl := strings.Join([]string{portalUrl, "?page=", strconv.Itoa(page)}, "")
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
		ids := parseIds(b)
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

/*
parseIds uses a regexp to get all the IDs off a page. It assumes that we're
going to call ioutil.ReadAll on resp.Body. It's one of the things that's pretty
testable, so is its own function.
*/
func parseIds(data []byte) []string {
	matches := idRegexp.FindAllSubmatch(data, -1)
	var ids []string
	for _, match := range matches {
		ids = append(ids, fmt.Sprintf("%s", match[1]))
	}
	return ids
}

/*
addIds reads from the text file of IDs and puts them in a channel. Perhaps for
readability this would be better in-lined where it is called.
*/
func addIds(idFile *os.File, idCh chan<- string) {
	scanner := bufio.NewScanner(idFile)
	for scanner.Scan() {
		idCh <- scanner.Text()
	}
	close(idCh)
}

/*
processStablePage reads from an ID channel, GETs an individual stable page,
and then extracts a Stable from that page. Finally, this function writes the
Stable to an output channel.
*/
func processStablePage(idCh <-chan string, stableCh chan<- Stable, wg *sync.WaitGroup) {
	defer wg.Done()

	for id := range idCh {
		stable := Stable{ID: id}

		stableUrl := strings.Join([]string{portalUrl, "//", id}, "")
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

/*
extractStable was easily testable. It takes a []byte of the individual stable
HTML page and extracts information about the stable from it.
*/
func (stable *Stable) extractStable(data []byte) error {
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
	Lat     float64
	Lng     float64
}

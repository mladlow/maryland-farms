package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

/*
Global string variables.
*/
var (
	geocodeUrl = os.Getenv("GEOCODE_URL")
	api_key    = os.Getenv("GEOCODE_API_KEY")
)

// Probably, this should return an error that's handled by the runner with an
// exit code if the URL is undefined. For now, moving on.
func Geocode(inFN string, outFN string) {
	fmt.Printf("Geolocating using server: %s\n", geocodeUrl)
	fmt.Printf("Reading from %s, writing to %s\n", inFN, outFN)

	dataFile, err := os.Open(inFN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening input file %v\n", err)
		return
	}

	scanner := bufio.NewScanner(dataFile)
	for scanner.Scan() {
		var stable Stable
		if err = json.Unmarshal(scanner.Bytes(), &stable); err != nil {
			fmt.Fprintf(os.Stderr, "Error unmarshaling stable %v\n", err)
			continue
		}
		queryAddress := url.QueryEscape(stable.Address)
		geocodeQuery := strings.Join([]string{geocodeUrl, queryAddress,
			`&key=`, api_key}, "")
		resp, err := http.Get(geocodeQuery)
		fmt.Printf("Query: %s\n", geocodeQuery)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error querying server %v\n", err)
			continue
		}
		b, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading body %v\n", err)
			continue
		}
		fmt.Printf("%s\n", b)
		fmt.Printf("%v\n", json.Valid(b))
		break
	}

	err = dataFile.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error closing input file %v\n", err)
	}
}

/*
extractGeocode manually unmarshals the Google Geocode API response and picks
out the bits we care about. It puts them in the stable object.

extractGeocode updates the address of the Stable to whatever Google thinks the
address should be, which is helpful because there are typos in the source data.
*/
func (stable *Stable) extractGeocode(rawData []byte) error {
	var j interface{}
	if err := json.Unmarshal(rawData, &j); err != nil {
		return errors.New(fmt.Sprintf("extractGeocode: Error unmarshaling geocode of %v\n", stable))
	}
	m := j.(map[string]interface{})
	results := m["results"].([]interface{})
	if len(results) != 1 {
		return errors.New(fmt.Sprintf("extractGeocode: Got %d results: %v\n", len(results), results))
	}
	result := results[0].(map[string]interface{})
	geometry := result["geometry"].(map[string]interface{})
	location := geometry["location"].(map[string]interface{})
	stable.Address = result["formatted_address"].(string)
	stable.Lat = location["lat"].(float64)
	stable.Lng = location["lng"].(float64)
	return nil
}

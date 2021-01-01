package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
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
	defer dataFile.Close()

	// If this is going to make an error let's get it out of the way before
	// doing all the queries.
	outFile, err := os.Create(outFN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening output file %v\n", err)
		return
	}
	defer outFile.Close()

	// Cheating a bit since I know there's 719 stables.
	geocodedStables := make([]Stable, 0, 800)

	dec := json.NewDecoder(dataFile)
	// Read open bracket
	_, err = dec.Token()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading array open %v\n", err)
	}

	for dec.More() {
		var stable Stable
		err := dec.Decode(&stable)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error unmarshaling stable %v\n", err)
			continue
		}
		queryAddress := url.QueryEscape(stable.Address)
		geocodeQuery := strings.Join([]string{geocodeUrl, queryAddress,
			`&key=`, api_key}, "")
		resp, err := http.Get(geocodeQuery)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error querying server for %s %v\n",
				stable.ID, err)
			continue
		}
		b, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading body for %s %v\n",
				stable.ID, err)
			continue
		}
		err = stable.extractGeocode(b)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error with JSON response for %s %v\n",
				stable.ID, err)
			fmt.Fprintf(os.Stderr, "JSON response: %s\n", b)
			continue
		}
		geocodedStables = append(geocodedStables, stable)
		time.Sleep(500 * time.Millisecond)
	}

	enc := json.NewEncoder(outFile)
	err = enc.Encode(geocodedStables)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing JSON to file %v\n", err)
	}

	_, err = dec.Token()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading array close %v\n", err)
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
	if len(results) == 0 {
		return errors.New(fmt.Sprintf("extractGeocode: Got no results: %v\n", results))
	}

	// If there is more than one result, we prefer either the first result or
	// the first result where geometry.location_type is ROOFTOP. This seemed to
	// match the results returned by Google Maps when I searched a couple
	// manually.
	preferredIdx := 0
	for idx, r := range results {
		result := r.(map[string]interface{})
		geometry := result["geometry"].(map[string]interface{})
		locationType := geometry["location_type"].(string)
		if locationType == "ROOFTOP" {
			preferredIdx = idx
			break
		}
	}

	preferredResult := results[preferredIdx].(map[string]interface{})
	geometry := preferredResult["geometry"].(map[string]interface{})
	location := geometry["location"].(map[string]interface{})
	stable.Address = preferredResult["formatted_address"].(string)
	stable.Lat = location["lat"].(float64)
	stable.Lng = location["lng"].(float64)
	return nil
}

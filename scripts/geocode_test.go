package main

import (
	"io/ioutil"
	"testing"
)

func TestGeocodeParser(t *testing.T) {
	data, err := ioutil.ReadFile("test_geocode.json")
	if err != nil {
		t.Error("Couldn't read geocode JSON file")
	}
	var stable Stable
	err = stable.extractGeocode(data)

	if err != nil {
		t.Errorf("Got %v while parsing geocode JSON\n", err)
	}

	if stable.Address != "3400 Damascus Rd, Brookeville, MD 20833, USA" {
		t.Errorf(`Extracted "%s" as address`, stable.Address)
	}
	if stable.Lat != 39.2261497 {
		t.Errorf("Extracted %v as latitude", stable.Lat)
	}
	if stable.Lng != -77.0680746 {
		t.Errorf("Extracted %v as longitude", stable.Lng)
	}
}

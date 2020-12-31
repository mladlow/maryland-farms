package main

import (
	"encoding/json"
	"io/ioutil"
	"testing"
)

func TestIdParser(t *testing.T) {
	data, err := ioutil.ReadFile("test_id_page.html")
	if err != nil {
		t.Error("Couldn't read stable list page")
	}
	ids := parseIds(data)
	if len(ids) != 10 {
		t.Error("Didn't get all the ids!")
	}
	if ids[0] != "5ce71cf74d7ef91b0c22a505" {
		t.Error("Didn't find expected id!")
	}
}

func TestStableParser(t *testing.T) {
	data, err := ioutil.ReadFile("test_stable_page.html")
	if err != nil {
		t.Error("Couldn't read test stable page")
	}
	expected := Stable{
		ID:      "ID",
		Name:    "YELLOW WOOD DRESSAGE INC.",
		Address: "1455 CAYOTS CORNER ROAD CHESAPEAKE CITY, MD, 21915",
		Phone:   "207-749-6458",
		Website: "www.yellowwooddressage.com",
	}
	stableData := Stable{ID: "ID"}
	stableData.extractStable(data)
	// Note to self that Stables should be comparable, however I find it
	// useful in tests to see what didn't parse.
	if expected.Name != stableData.Name {
		t.Errorf("Extracted %s as name", stableData.Name)
	}
	if expected.Address != stableData.Address {
		t.Errorf("Extracted %s as address", stableData.Address)
	}
	if expected.Phone != stableData.Phone {
		t.Errorf("Extracted %s as phone", stableData.Phone)
	}
	if expected.Website != stableData.Website {
		t.Errorf("Extracted %s as website", stableData.Website)
	}
}

func TestStablePoboxParser(t *testing.T) {
	data, err := ioutil.ReadFile("test_stable_pobox_page.html")
	if err != nil {
		t.Error("Couldn't read test stable page")
	}
	expected := Stable{
		ID:      "ID",
		Name:    "FOX QUARTER FARM, LLC",
		Address: "3875 BARK HILL ROAD P.O. Box 600 UNION BRIDGE, MD, 21791",
		Phone:   "410-984-2011",
	}
	stableData := Stable{ID: "ID"}
	err = stableData.extractStable(data)
	if err != nil {
		t.Errorf("Expected success, got %v\n", err)
	}
	if expected.Name != stableData.Name {
		t.Errorf("Extracted %s as name", stableData.Name)
	}
	if expected.Address != stableData.Address {
		t.Errorf("Extracted %s as address", stableData.Address)
	}
	if expected.Phone != stableData.Phone {
		t.Errorf("Extracted %s as phone", stableData.Phone)
	}
	if expected.Website != stableData.Website {
		t.Errorf("Extracted %s as website", stableData.Website)
	}
	_, err = json.Marshal(stableData)
	if err != nil {
		t.Errorf("Got error marshaling resulting json: %v\n", err)
	}
}

func TestStableParserEdges(t *testing.T) {
	testBytes := []byte("<article>blah</article><article>blah2</article>")
	s := Stable{}
	err := s.extractStable(testBytes)
	if err == nil {
		t.Error("Should fail if two articles in data\n")
	}

	testBytes = []byte("")
	err = s.extractStable(testBytes)
	if err == nil {
		t.Error("Should fail if no articles in data\n")
	}
}

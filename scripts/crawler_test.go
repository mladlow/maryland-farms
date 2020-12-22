package main

import (
	"io/ioutil"
	"testing"
)

func TestIdParserWithReadAll(t *testing.T) {
	data, err := ioutil.ReadFile("test_id_page.html")
	if err != nil {
		t.Error("Couldn't read stable list page")
	}
	ids := parseIdsWithReadAll(data)
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
		Name:    "YELLOW WOOD DRESSAGE INC.",
		Address: "1455 CAYOTS CORNER ROAD CHESAPEAKE CITY, MD, 21915",
		Phone:   "207-749-6458",
		Website: "www.yellowwooddressage.com",
	}
	stableData, _ := extractStable(data)
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

func TestStableParserEdges(t *testing.T) {
	testBytes := []byte("<article>blah</article><article>blah2</article>")
	_, err := extractStable(testBytes)
	if err == nil {
		t.Error("Should fail if two articles in data\n")
	}

	testBytes = []byte("")
	_, err = extractStable(testBytes)
	if err == nil {
		t.Error("Should fail if no articles in data\n")
	}
}

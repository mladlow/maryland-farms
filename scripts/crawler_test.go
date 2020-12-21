package main

import (
	"io/ioutil"
	"testing"
)

func TestIdParserWithReadAll(t *testing.T) {
	data, err := ioutil.ReadFile("test_page.html")
	if err != nil {
		t.Error("Couldn't read stable list page")
	}
	ids := parseIdsWithReadAll(data)
	if len(ids) != 10 {
		t.Error("Didn't get all the ids!")
	}
	if ids[0] != "5ce71cf74d7ef91b0c22a505" {
		t.Error("Did't find expected id!")
	}
}

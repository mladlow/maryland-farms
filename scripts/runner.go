package main

import (
	"flag"
	"fmt"
	"os"
)

func helpAndExit() {
	fmt.Println("Expected 'geocode' or 'crawl' subcommands")
	os.Exit(1)
}

func main() {
	geocodeCmd := flag.NewFlagSet("geocode", flag.ExitOnError)
	inFN := geocodeCmd.String("in", "portalData.json",
		"File containing stable data.")
	outFN := geocodeCmd.String("out", "geocoded.json",
		"Target file for geocoded data.")
	crawlerCmd := flag.NewFlagSet("crawl", flag.ExitOnError)

	if len(os.Args) < 2 {
		helpAndExit()
	}

	switch os.Args[1] {
	case "geocode":
		geocodeCmd.Parse(os.Args[2:])
		Geocode(*inFN, *outFN)
	case "crawl":
		fmt.Println("Crawling portal...")
		crawlerCmd.Parse(os.Args[2:])
	default:
		helpAndExit()
	}
}

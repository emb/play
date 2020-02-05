/*
tz takes an argument describing a timezone location and prints some
information about it.

TODO(emb): Add a flag to read form a specific TZ database file
*/
package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() == 0 {
		displayTZ("Local")
	}

	for i := 0; i < flag.NArg(); i++ {
		displayTZ(flag.Arg(i))
	}
}

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s LOCATION [LOCATION...]\n", os.Args[0])
	fmt.Fprintf(flag.CommandLine.Output(), `
Location must be a valid IANA time zone database name such as 'Australia/Melbourne'
`)
}

func displayTZ(name string) {
	loc, err := time.LoadLocation(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing location: %s\n", err)
		return
	}

	zone, offset := time.Now().In(loc).Zone()
	fmt.Printf("%s offset=%s\n", zone, time.Duration(offset)*time.Second) // Use Duration for pretty printing
}

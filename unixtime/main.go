/*
unixtime attempts at converting a string to time.

Every time I have to use the date utility to convert a number to time
I get confused. The GNU date command functions differently than the
BSD date command and all I want is a utility that I pass in something
that might be a datetime that I need to know what it looks like for a
human and a machine.
*/
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() == 0 {
		display(time.Now())
		return
	}
	c, err := convert(flag.Arg(0))
	if err != nil {
		exit(err)
	}
	display(c)
}

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s TIME\n", os.Args[0])
	fmt.Fprintln(flag.CommandLine.Output(), `
If the time is a number that will be treated as a
Unix time in seconds. Otherwise argument is assumed to
be a date time in RFC 3339

Try:
	$ unixtime 1576662079

or 
        $ unixtime 2019-12-19T00:00:00-05:00

When parsing datetime you must provide the entire RFC3339
string to disambiguate zones.
`)
}

func exit(err error) {
	if err == nil {
		return
	}
	fmt.Fprintf(flag.CommandLine.Output(), "Error: %s\n", err)
	os.Exit(1)
}

func convert(in string) (time.Time, error) {
	// Attempt at reading unix time
	n, err := strconv.ParseInt(in, 10, 64)
	if err == nil {
		return time.Unix(n, 0), nil
	}

	// If it is not a unix time then interpret the string as date
	// time.
	t, err := time.Parse(time.RFC3339, in)
	if err != nil {
		return time.Time{}, fmt.Errorf("expecting a unix time in seconds or an RFC3339 date/time: %s", in, err)
	}
	return t, nil
}

func display(t time.Time) {
	fmt.Printf("%s unix=%d\n", t, t.Unix())
}

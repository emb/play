// urlpretty pretty print a huge fucking url that you cannot look at
// to something a little easier to digest.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
)

var jsonFlag = flag.Bool("json", false, "output as json")

func exit(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err)
	}
	os.Exit(1)
}

func usage() {
	fmt.Fprintf(os.Stderr, "USAGE: %s URL [URL]\n\n", os.Args[0])
	flag.PrintDefaults()
}

func reader() io.Reader {
	if flag.NArg() > 0 {
		return strings.NewReader(strings.Join(flag.Args(), "\n"))
	}
	return os.Stdin
}

// pURL an internal representation of url.URL struct. This is mainly
// here to provide a simpler json representation.
type pURL struct {
	Scheme   string
	Host     string
	Path     string
	Fragment string
	Query    map[string][]string
}

func writeJSON(u *pURL, w io.Writer) error {
	return json.NewEncoder(w).Encode(u)
}

func writeText(u *pURL, w io.Writer) error {
	fmt.Fprintln(w, "Scheme:   ", u.Scheme)
	fmt.Fprintln(w, "Host:     ", u.Host)
	fmt.Fprintln(w, "Path:     ", u.Path)
	if len(u.Fragment) > 0 {
		fmt.Fprintln(w, "Fragment: ", u.Fragment)
	}
	if len(u.Query) > 0 {
		fmt.Fprintln(w, "Query:")
		for k, v := range u.Query {
			fmt.Fprintf(w, "\t%s: %s\n", k, v)
		}
	}
	fmt.Fprintln(w, "%%")
	return nil
}

func writer(json bool) func(*pURL, io.Writer) error {
	if json {
		return writeJSON
	}
	return writeText
}

func main() {
	flag.Usage = usage
	flag.Parse()
	scanner := bufio.NewScanner(reader())
	w := writer(*jsonFlag)
	for scanner.Scan() {
		u, err := url.Parse(scanner.Text())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s", err)
			continue
		}
		w(&pURL{u.Scheme, u.Host, u.Path, u.Fragment, u.Query()}, os.Stdout)
	}
	if err := scanner.Err(); err != nil {
		exit(err)
	}
}

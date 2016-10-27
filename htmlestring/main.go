// htmlstring escape/unescape html strings
package main

import (
	"bufio"
	"flag"
	"fmt"
	"html"
	"os"
)

var unescapeFlag bool

func init() {
	flag.BoolVar(&unescapeFlag, "u", false, "unescape html string")
}

func main() {
	flag.Parse()
	stringer := htmlstringer()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fmt.Println(stringer(scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error reading stdin: ", err)
	}
}

func htmlstringer() func(string) string {
	if unescapeFlag {
		return html.UnescapeString
	}
	return html.EscapeString
}

// qp (quoted-pritable) encodes and decodes quoted-printable text.
package main

import (
	"flag"
	"io"
	"mime/quotedprintable"
	"os"
)

var decodeFlag bool

func init() {
	flag.BoolVar(&decodeFlag, "d", false, "docode incoming text (default is false)")
}

func main() {
	flag.Parse()
	io.Copy(writer(), reader())
}

func reader() io.Reader {
	if decodeFlag {
		return quotedprintable.NewReader(os.Stdin)
	}
	return os.Stdin
}

func writer() io.Writer {
	if decodeFlag {
		return os.Stdout
	}
	return quotedprintable.NewWriter(os.Stdout)
}

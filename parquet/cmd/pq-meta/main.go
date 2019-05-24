/*
The pq-meta command displays meta data about a parquet file.

An example usage

	$ pq-meta your-data-file.snappy.parquet
*/
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/emb/play/parquet"
	"github.com/emb/play/parquet/meta"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s PARQUET_FILE\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		exitErr("mast pass a file as an argument\n")
	}

	p, err := parquet.NewParquetFile(os.Args[1])
	if err != nil {
		exitErr("Error: %s\n", err)
	}
	defer p.Close()

	meta, err := p.MetaData()
	if err != nil {
		exitErr("Error: %s\n", err)
	}
	printMetaData(os.Stdout, meta)
}

func exitErr(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(1)
}

func printMetaData(f io.Writer, meta *meta.FileMetaData) {
	fmt.Fprintf(f, "Created By: %s\n", *meta.CreatedBy)
	fmt.Fprintf(f, "Version: %d\n", meta.Version)
	fmt.Fprintf(f, "Rows: %d\n", meta.NumRows)
	fmt.Fprintf(f, "Columns: %d\n", len(meta.Schema))
	fmt.Println("%%")
	printSchema(f, deep(make([]int, 0)), meta.Schema)
}

type deep []int

func (d deep) Push(i int) deep {
	d = append(d, i)
	return d
}

func (d deep) String() string {
	if len(d) == 0 {
		return ""
	}
	last := len(d) - 1
	if d[last] == 0 {
		d = d[:last]
		return d.String()
	}
	d[last]--
	return strings.Repeat("  ", len(d))
}

func printSchema(f io.Writer, prefix deep, fields []*meta.SchemaElement) {
	if len(fields) == 0 {
		return
	}

	field := fields[0]
	if field.Type == nil { // has child elements
		fmt.Fprintf(f, "%s%s:\n", prefix, field.Name)
		printSchema(f, prefix.Push(int(*field.NumChildren)), fields[1:])
		return
	}

	var converted string
	if field.ConvertedType != nil {
		converted = fmt.Sprintf("/%s", *field.ConvertedType)
	}
	fmt.Fprintf(f, "%s%s: %s%s (%s)\n", prefix, field.Name,
		strings.ToLower(field.Type.String()),
		strings.ToLower(converted),
		strings.ToLower(field.RepetitionType.String()))
	printSchema(f, prefix, fields[1:])
}

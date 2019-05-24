// Package meta describes parquet metadata types.
//
// The thrift definition was copied from https://github.com/apache/parquet-format/
package meta

//go:generate thrift -out .. -r --gen go:package=meta parquet.thrift

PARQUET
=======

This is a learning opportunity to understand the ins and outs of the
parquet format. The ultimate idea is to build an awk variant that
works with parquet.

Tools
-----

* [pq-meta](./cmd/pq-meta): A utility to print meta data information
  about parquet files.
  - get it `go get github.com/emb/play/parquet/cmd/pq-meta`
  - use it `$ pq-meta your-file.snappy.parquet`

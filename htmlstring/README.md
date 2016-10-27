# htmlstring

Escape arbitary strings into HTML strings

## Usage

Utility uses `stdin` and `stdout`

```
$ ./htmlestring -h
Usage of ./htmlestring:
  -u	unescape html string
```

### Example

```
$ echo "0 & 1 < 1" | ./htmlestring 
0 &amp; 1 &lt; 1
```

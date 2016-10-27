# README

Pretty print long URLs

## Usage

```
$ urlpretty -help
USAGE: urlpretty URL [URL]

  -json
    	output as json
```

### Example

From standard in

```
$ cat <<EOF | urlpretty 
> http://hello.world/?foo=bar&eek=ook
> http://link/?to=some&really=long&trying=hard&to=be=&long=url
> EOF
Scheme:    http
Host:      hello.world
Path:      /
Query:
	eek: [ook]
	foo: [bar]
%%
Scheme:    http
Host:      link
Path:      /
Query:
	to: [some be=]
	really: [long]
	trying: [hard]
	long: [url]
%%
```



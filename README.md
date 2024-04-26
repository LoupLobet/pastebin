# Pastebin
## Example
### Post a document
```
% pb http://host <file.txt
vh74gk9l0sa4vp18
```
### Get a document
```
% curl http://host/vh74gk9l0sa4vp18
File content !
```

## Compile
`% make`

Binaries: `bin/pbd`, `bin/pb`

## Usages
```
Usage of ./pbd:
  -c string
    	Document name charset (default "abcdefghijklmnopqrstuvwxyz0123456789")
  -d string
    	Root dir (default "docs")
  -l string
    	Listen address (default "127.0.0.1:1488")
  -n int
    	Document name length (default 9)
  -s int
    	Maximum Doc size (default 10000000)
  -t duration
    	Document lifetime (default 168h0m0s)
  -x int
    	Maximum Doc number (default 2000)
```

```
Usage of ./pb:
  -c string
    	Document name charset (default "abcdefghijklmnopqrstuvwxyz0123456789")
  -n int
    	Document name length (default 16)
  -t duration
    	Document lifetime (default 168h0m0s)
```

all: pastebin pastebind

pastebin: cmd/pastebin
	mkdir -p bin
	go build -C $< -o ../../bin/$@

pastebind: cmd/pastebind
	mkdir -p bin
	go build -C $< -o ../../bin/$@

all: pb pbd

pb: cmd/pastebin
	mkdir -p bin
	go build -C $< -o ../../bin/$@

pbd: cmd/pastebind
	mkdir -p bin
	go build -C $< -o ../../bin/$@

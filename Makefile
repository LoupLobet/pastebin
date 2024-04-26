all: pb pbd

pb: cmd/pb
	mkdir -p bin
	go build -C $< -o ../../bin/$@

pbd: cmd/pb
	mkdir -p bin
	go build -C $< -o ../../bin/$@

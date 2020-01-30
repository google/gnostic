
all:
	./COMPILE-PROTOS.sh
	go get ./...
	go install ./...
	cd extensions/sample; make

test:
	go test . -v
	go test ./plugins -v
	go test ./extensions -v
	go test ./apps/petstore-builder -v

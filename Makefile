
all:
	./COMPILE-PROTOS.sh
	go get ./...
	go install ./...
	cd extensions/sample; make

test:
	go test . -v
	cd plugins; go test . -v
	cd extensions; go test . -v
	cd apps/petstore-builder; go test . -v

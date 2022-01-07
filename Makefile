
all:
	go generate ./...
	go get ./...
	go install ./...
	cd extensions/sample; make
	for CMD in `ls cmd` ; do (cd ./cmd/$$CMD; echo "install $$CMD"; go install ./...); done

test:
	# since some tests call separately-built binaries, clear the cache to ensure all get run
	go clean -testcache
	go test ./... -v
	for CMD in `ls cmd` ; do (cd ./cmd/$$CMD; echo "test $$CMD"; go test ./... -v); done

tidy:
	go mod tidy
	for CMD in `ls cmd` ; do (cd ./cmd/$$CMD; echo "tidy $$CMD"; go mod tidy); done

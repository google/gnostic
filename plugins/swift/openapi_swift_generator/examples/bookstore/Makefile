
all:
	rm -f openapi_swift_generator
	ln -s ../../openapi_swift_generator 
	openapic bookstore.json --swift_generator_out=Sources/Bookstore
	swift build

test:
	swift test

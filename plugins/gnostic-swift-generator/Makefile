
BINDIR=.build/x86_64-apple-macosx10.10/debug

all:
	swift build

install: all
	cp $(BINDIR)/gnostic-swift-generator $(GOPATH)/bin/gnostic-swift-generator
	cp $(BINDIR)/gnostic-swift-generator $(GOPATH)/bin/gnostic-swift-client
	cp $(BINDIR)/gnostic-swift-generator $(GOPATH)/bin/gnostic-swift-server

clean:
	rm -rf .build Packages

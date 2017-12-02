
TGT=gnostic-swift-sample

BINDIR=.build/x86_64-apple-macosx10.10/debug

all:
	swift build

install: all
	cp $(BINDIR)/$(TGT) $(GOPATH)/bin/$(TGT)

clean :
	rm -rf Packages
	rm -rf .build
	rm -f Package.pins Package.resolved

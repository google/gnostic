go get github.com/golang/protobuf/protoc-gen-go


# ONE TIME
#
#

############################# FOR E2E TESTING ########################################
# ensure vendorextension proto contract is compiled.
pushd $GOPATH/src/github.com/googleapis/gnostic/extension/extension_data
./COMPILE-EXTENSION-DATA.sh
popd

pushd $GOPATH/src/github.com/googleapis/gnostic/generator 
./INSTALL.sh

pushd $GOPATH/src/github.com/googleapis/gnostic
go install

# ensure tool to generate the vendor extension compiler is installed.
pushd $GOPATH/src/github.com/googleapis/gnostic/extension/extensionc
go build
go install

####################################################################################33



# Now generate sample extension plugins and install them.
#
#
pushd $GOPATH/src/github.com/googleapis/gnostic/extension

    EXTENSION_OUT_DIR="github.com/googleapis/gnostic/extension/sample/generated"
    # For Google Extension Example
    #
    #
    GOOGLE_EXTENSION_SCHEMA="sample/x-google.json"

    extensionc $GOOGLE_EXTENSION_SCHEMA --out_dir_relative_to_gopath_src=$EXTENSION_OUT_DIR

    pushd $GOPATH/src/$EXTENSION_OUT_DIR/openapi_extensions_google/proto
        protoc --go_out=Mgoogle/protobuf/any.proto=github.com/golang/protobuf/ptypes/any:. *.proto
        go install
    popd

    pushd  $GOPATH/src/$EXTENSION_OUT_DIR/openapi_extensions_google
        go install
    popd

    # For IBM Extension Example
    #
    #
    IBM_EXTENSION_SCHEMA="sample/x-ibm.json"

    extensionc $IBM_EXTENSION_SCHEMA --out_dir_relative_to_gopath_src=$EXTENSION_OUT_DIR

    pushd $GOPATH/src/$EXTENSION_OUT_DIR/openapi_extensions_ibm/proto
        protoc --go_out=Mgoogle/protobuf/any.proto=github.com/golang/protobuf/ptypes/any:. *.proto
        go install
    popd

    pushd $GOPATH/src/$EXTENSION_OUT_DIR/openapi_extensions_ibm
        go install
    popd
popd

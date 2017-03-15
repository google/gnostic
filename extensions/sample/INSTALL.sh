go get github.com/golang/protobuf/protoc-gen-go


# ONE TIME
#
#

############################# FOR E2E TESTING ########################################
# ensure vendorextension proto contract is compiled.
pushd $GOPATH/src/github.com/googleapis/gnostic/extensions
./COMPILE-EXTENSION.sh
popd

pushd $GOPATH/src/github.com/googleapis/gnostic/generator
./INSTALL.sh

pushd $GOPATH/src/github.com/googleapis/gnostic
go install


######################################################################################



# Now generate sample extension plugins and install them.
#
#
pushd $GOPATH/src/github.com/googleapis/gnostic/extensions

    EXTENSION_OUT_DIR=$GOPATH/src/"github.com/googleapis/gnostic/extensions/sample/generated"
    # For Google Extension Example
    #
    #
    GOOGLE_EXTENSION_SCHEMA="sample/x-insourcetech.json"

    generator --ext $GOOGLE_EXTENSION_SCHEMA --out_dir=$EXTENSION_OUT_DIR

    pushd $EXTENSION_OUT_DIR/openapi_extensions_insourcetech/proto
        protoc --go_out=Mgoogle/protobuf/any.proto=github.com/golang/protobuf/ptypes/any:. *.proto
        go install
    popd

    pushd  $EXTENSION_OUT_DIR/openapi_extensions_insourcetech
        go install
    popd

    # For IBM Extension Example
    #
    #
    IBM_EXTENSION_SCHEMA="sample/x-nerdwaretech.json"

    generator --ext $IBM_EXTENSION_SCHEMA --out_dir=$EXTENSION_OUT_DIR

    pushd $EXTENSION_OUT_DIR/openapi_extensions_nerdwaretech/proto
        protoc --go_out=Mgoogle/protobuf/any.proto=github.com/golang/protobuf/ptypes/any:. *.proto
        go install
    popd

    pushd $EXTENSION_OUT_DIR/openapi_extensions_nerdwaretech
        go install
    popd
popd

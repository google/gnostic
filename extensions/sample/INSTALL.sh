go get github.com/golang/protobuf/protoc-gen-go


# ONE TIME
#
#

############################# FOR E2E TESTING ########################################
# ensure vendorextension proto contract is compiled.

pushd $GOPATH/src/github.com/googleapis/gnostic/extensions
echo compile gnostic extension protos
./COMPILE-EXTENSION.sh
popd

pushd $GOPATH/src/github.com/googleapis/gnostic/generator
echo install gnostic-generator
go install
popd

pushd $GOPATH/src/github.com/googleapis/gnostic
echo install gnostic
go install
popd


######################################################################################



# Now generate sample extension plugins and install them.
#
#
pushd $GOPATH/src/github.com/googleapis/gnostic/extensions

    EXTENSION_OUT_DIR=$GOPATH/src/github.com/googleapis/gnostic/extensions/sample/generated
    # For SAMPLE_ONE Extension Example
    #
    #
    SAMPLE_ONE_EXTENSION_SCHEMA="sample/x-samplecompanyone.json"

    echo generate samplecompanyone extension
    generator --extension $SAMPLE_ONE_EXTENSION_SCHEMA --out_dir=$EXTENSION_OUT_DIR

    pushd $EXTENSION_OUT_DIR/gnostic-x-samplecompanyone/proto
        protoc --go_out=Mgoogle/protobuf/any.proto=github.com/golang/protobuf/ptypes/any:. *.proto
        go install
    popd

    pushd  $EXTENSION_OUT_DIR/gnostic-x-samplecompanyone
        go install
    popd

    # For SAMPLE_TWO Extension Example
    #
    #
    SAMPLE_TWO_EXTENSION_SCHEMA="sample/x-samplecompanytwo.json"

    echo generate samplecompanyone extension
    generator --extension $SAMPLE_TWO_EXTENSION_SCHEMA --out_dir=$EXTENSION_OUT_DIR

    pushd $EXTENSION_OUT_DIR/gnostic-x-samplecompanytwo/proto
        protoc --go_out=Mgoogle/protobuf/any.proto=github.com/golang/protobuf/ptypes/any:. *.proto
        go install
    popd

    pushd $EXTENSION_OUT_DIR/gnostic-x-samplecompanytwo
        go install
    popd
popd

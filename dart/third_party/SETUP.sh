#!/bin/sh
#
# Download dependencies needed to build the flame tools.
#
if [ ! -d "api-common-protos" ]
then
  git clone https://github.com/googleapis/api-common-protos
else
  echo "Using previous download of third_party/api-common-protos."
fi
#!/bin/bash
# Copyright 2020 Google LLC. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This should point to the .proto files distributed with protoc.
export PROTO_PROTOS="$HOME/local/include"

# This is a local directory containing .proto files used by many APIs.
export ANNOTATIONS="third_party/api-common-protos"

echo "Generating dart support code."
protoc  \
	${PROTO_PROTOS}/google/protobuf/timestamp.proto \
	${PROTO_PROTOS}/google/protobuf/any.proto \
	${PROTO_PROTOS}/google/protobuf/empty.proto \
	--proto_path=${ANNOTATIONS} \
	--proto_path=../../openapiv2 \
	--proto_path=../../openapiv3 \
	--proto_path=../../discovery \
	--proto_path=../../surface \
	--proto_path=../../plugins \
	../../openapiv2/OpenAPIv2.proto \
	../../openapiv3/OpenAPIv3.proto \
	../../discovery/discovery.proto \
	../../surface/surface.proto \
	../../plugins/plugin.proto \
	--dart_out=grpc:lib/generated

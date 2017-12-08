plugin:
	node_modules/.bin/pbjs -t json \
		../../OpenAPIv2/OpenAPIv2.proto \
		../../OpenAPIv3/OpenAPIv3.proto \
		../../discovery/discovery.proto \
		../../surface/surface.proto \
		../plugin.proto \
		> bundle.json
	node_modules/.bin/nexe gnostic-lint-node.js

run:	plugin
	gnostic ../../examples/v2.0/yaml/petstore.yaml --lint-node-out=-

setup:
	npm install protobufjs
	npm install get-stdin
	npm install nexe

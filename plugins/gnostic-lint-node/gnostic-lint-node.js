// import libraries
const protobuf = require("protobufjs");
const getStdin = require('get-stdin');

// import messages
var root = protobuf.Root.fromJSON(require("./bundle.json"));
var Document = root.lookupType("openapi.v2.Document")
var Request = root.lookupType("gnostic.plugin.v1.Request")
var Response = root.lookupType("gnostic.plugin.v1.Response")

getStdin.buffer().then(buffer => {
	var request = Request.decode(buffer);
	console.error('message: %s\n\n', JSON.stringify(request))

	var openapi2 = request.openapi2
	var paths = openapi2.paths.path

	console.error('paths: %s\n\n', paths)

	for (var i in paths) {
		const path = paths[i]
		console.error('path %s\n\n', path.name)
		const getOperation = path.value.get
		console.error('get %s\n\n', JSON.stringify(getOperation))
	}

	var payload = {
		errors: [],
		files: [{
			name: "report.txt",
			data: Buffer.from("testing 123\n", 'utf8')
		}]
	};

	// Verify the payload if necessary (i.e. when possibly incomplete or invalid)
	var errMsg = Response.verify(payload);
	if (errMsg)
		throw Error(errMsg);

	var message = Response.create(payload)
	var buffer = Response.encode(message).finish();
	process.stdout.write(buffer);

}).catch(err => console.error(err));

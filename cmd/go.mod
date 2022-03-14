module github.com/google/gnostic/cmd

go 1.12

require (
	github.com/docopt/docopt-go v0.0.0-20180111231733-ee0de3bc6815
	github.com/flowstack/go-jsonschema v0.1.1
	github.com/golang/protobuf v1.5.2
	github.com/google/gnostic v0.0.0-00010101000000-000000000000
	google.golang.org/genproto v0.0.0-20220314164441-57ef72a4c106
	google.golang.org/protobuf v1.27.1
)

require (
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/stoewer/go-strcase v1.2.0 // indirect
	golang.org/x/net v0.0.0-20210805182204-aaa1db679c0d // indirect
	golang.org/x/text v0.3.6 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
)

replace github.com/google/gnostic => ../

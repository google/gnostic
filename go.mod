module github.com/ASkyFullOfStar/gnosticcyr

go 1.12

require (
	buf.bilibili.co/bapis/bapis-gen-go/gogo/protobuf v0.0.0-master-0.0.20221211061531-cde0bf5f049a
	github.com/docopt/docopt-go v0.0.0-20180111231733-ee0de3bc6815
	github.com/flowstack/go-jsonschema v0.1.1
	github.com/golang/protobuf v1.5.2
	github.com/google/gnostic v0.6.9
	github.com/kr/pretty v0.3.1 // indirect
	github.com/stoewer/go-strcase v1.2.0
	golang.org/x/tools v0.0.0-20210106214847-113979e3529a
	google.golang.org/genproto v0.0.0-20220107163113-42d7afdf6368
	google.golang.org/protobuf v1.27.1
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
)

replace github.com/google/gnostic => github.com/ASkyFullOfStar/gnosticcyr v0.0.0-20221212111828-f7ab087b4633

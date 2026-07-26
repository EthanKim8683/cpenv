module github.com/EthanKim8683/cpenv

go 1.26.3

tool (
	connectrpc.com/connect/cmd/protoc-gen-connect-go
	google.golang.org/protobuf/cmd/protoc-gen-go
)

require (
	connectrpc.com/connect v1.20.0
	github.com/caarlos0/env/v11 v11.4.1
	github.com/spf13/afero v1.15.0
	github.com/stretchr/testify v1.11.1
	go.starlark.net v0.0.0-20260708150628-5395d018f003
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.37.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

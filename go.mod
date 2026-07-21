module github.com/EthanKim8683/cpenv

go 1.26.3

tool (
	connectrpc.com/connect/cmd/protoc-gen-connect-go
	google.golang.org/protobuf/cmd/protoc-gen-go
)

require (
	connectrpc.com/connect v1.20.0
	github.com/caarlos0/env/v11 v11.4.1
	github.com/google/go-jsonnet v0.22.0
	github.com/rs/cors v1.11.1
	github.com/spf13/afero v1.15.0
	github.com/spf13/cobra v1.10.2
	github.com/stretchr/testify v1.11.1
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	go.starlark.net v0.0.0-20260708150628-5395d018f003 // indirect
	golang.org/x/crypto v0.45.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)

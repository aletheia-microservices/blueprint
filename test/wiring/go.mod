module github.com/blueprint-uservices/blueprint/test/wiring

go 1.26.0

require (
	github.com/blueprint-uservices/blueprint/test/workflow v0.0.0
	github.com/stretchr/testify v1.10.0
	golang.org/x/exp v0.0.0-20240525044651-4c93da0ed11d
)

require (
	github.com/blueprint-uservices/blueprint/runtime v0.0.0-20250815220819-bcd4a51069cb // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.mongodb.org/mongo-driver v1.17.6 // indirect
	go.opentelemetry.io/otel v1.32.0 // indirect
	go.opentelemetry.io/otel/metric v1.32.0 // indirect
	go.opentelemetry.io/otel/trace v1.32.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/blueprint-uservices/blueprint/test/workflow => ../workflow

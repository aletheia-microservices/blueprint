package specs

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/plugins/goproc"
	"github.com/blueprint-uservices/blueprint/plugins/http"
	"github.com/blueprint-uservices/blueprint/plugins/linuxcontainer"
	"github.com/blueprint-uservices/blueprint/plugins/grpc"
	"github.com/blueprint-uservices/blueprint/plugins/thrift"
)

func applyGRPCDefaults(spec wiring.WiringSpec, serviceName, procName, ctrName string) string {
	grpc.Deploy(spec, serviceName)
	goproc.CreateProcess(spec, procName, serviceName)
	return linuxcontainer.CreateContainer(spec, ctrName, procName)
}

func applyThriftDefaults(spec wiring.WiringSpec, serviceName, procName, ctrName string) string {
	thrift.Deploy(spec, serviceName)
	goproc.CreateProcess(spec, procName, serviceName)
	return linuxcontainer.CreateContainer(spec, ctrName, procName)
}

func applyHTTPDefaults(spec wiring.WiringSpec, serviceName, procName, ctrName string) string {
	http.Deploy(spec, serviceName)
	goproc.CreateProcess(spec, procName, serviceName)
	return linuxcontainer.CreateContainer(spec, ctrName, procName)
}

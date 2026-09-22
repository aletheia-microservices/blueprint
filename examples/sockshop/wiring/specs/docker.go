package specs

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/carts"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/catalogue"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/frontend"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/orders"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/payment"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/queuemaster"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/shipping"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/user"
	"github.com/blueprint-uservices/blueprint/plugins/cmdbuilder"
	"github.com/blueprint-uservices/blueprint/plugins/goproc"
	"github.com/blueprint-uservices/blueprint/plugins/http"
	"github.com/blueprint-uservices/blueprint/plugins/linuxcontainer"
	"github.com/blueprint-uservices/blueprint/plugins/mongodb"
	"github.com/blueprint-uservices/blueprint/plugins/mysql"
	"github.com/blueprint-uservices/blueprint/plugins/rabbitmq"
	"github.com/blueprint-uservices/blueprint/plugins/grpc"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
)

// A wiring spec that deploys each service into its own Docker container and using gRPC to communicate between services.
// All RPC calls are retried up to 3 times.  RPC clients use a client pool with 10 clients.
// All services are instrumented with OpenTelemetry and traces are exported to Zipkin
// The user, cart, shipping, and orders services using separate MongoDB instances to store their data.
// The catalogue service uses MySQL to store catalogue data.
var Docker = cmdbuilder.SpecOption{
	Name:        "docker",
	Description: "Deploys each service in a separate container with gRPC, and uses mongodb as NoSQL database backends and rabbitmq as the queue backend.",
	Build:       makeDockerRabbitSpec,
}

func applyDockerQueueHandlerDefaults(spec wiring.WiringSpec, serviceName, procName, ctrName string) string {
	goproc.CreateProcess(spec, procName, serviceName)
	return linuxcontainer.CreateContainer(spec, ctrName, procName)
}

func applyDockerDefaults(spec wiring.WiringSpec, serviceName, procName, ctrName string) string {
	grpc.Deploy(spec, serviceName)
	goproc.CreateProcess(spec, procName, serviceName)
	return linuxcontainer.CreateContainer(spec, ctrName, procName)
}

func applyHTTPDefaults(spec wiring.WiringSpec, serviceName, procName, ctrName string) string {
	http.Deploy(spec, serviceName)
	goproc.CreateProcess(spec, procName, serviceName)
	return linuxcontainer.CreateContainer(spec, ctrName, procName)
}

func makeDockerRabbitSpec(spec wiring.WiringSpec) ([]string, error) {
	var containers []string
	var allServices []string

	user_db := mongodb.Container(spec, "user_db")
	cart_db := mongodb.Container(spec, "cart_db")
	order_db := mongodb.Container(spec, "order_db")
	catalogue_db := mysql.Container(spec, "catalogue_db")
	shipdb := mongodb.Container(spec, "ship_db")
	shipqueue := rabbitmq.Container(spec, "ship_queue", "ship_queue")

	allServices = append(allServices, user_db)
	allServices = append(allServices, cart_db)
	allServices = append(allServices, order_db)
	allServices = append(allServices, catalogue_db)
	allServices = append(allServices, shipdb)
	allServices = append(allServices, shipqueue)

	user_service := workflow.Service[user.UserService](spec, "user_service", user_db)
	user_service_ctr := applyDockerDefaults(spec, user_service, "user_service_proc", "user_service_container")
	containers = append(containers, user_service_ctr)

	payment_service := workflow.Service[payment.PaymentService](spec, "payment_service", "500")
	payment_service_ctr := applyDockerDefaults(spec, payment_service, "payment_service_proc", "payment_service_container")
	containers = append(containers, payment_service_ctr)

	cart_service := workflow.Service[carts.CartService](spec, "cart_service", cart_db)
	cart_service_ctr := applyDockerDefaults(spec, cart_service, "cart_service_proc", "cart_service_container")
	containers = append(containers, cart_service_ctr)

	shipping_service := workflow.Service[shipping.ShippingService](spec, "shipping_service", shipqueue, shipdb)
	shipping_service_ctr := applyDockerDefaults(spec, shipping_service, "shipping_service_proc", "shipping_service_container")
	containers = append(containers, shipping_service_ctr)

	queue_service := workflow.Service[queuemaster.QueueMaster](spec, "queue_master", shipqueue, shipping_service)
	queue_service_ctr := applyDockerQueueHandlerDefaults(spec, queue_service, "queue_service_proc", "queue_service_container")
	containers = append(containers, queue_service_ctr)

	order_service := workflow.Service[orders.OrderService](spec, "order_service", user_service, cart_service, payment_service, shipping_service, order_db)
	order_service_ctr := applyDockerDefaults(spec, order_service, "order_service_proc", "order_service_container")
	containers = append(containers, order_service_ctr)

	catalogue_service := workflow.Service[catalogue.CatalogueService](spec, "catalogue_service", catalogue_db)
	catalogue_service_ctr := applyDockerDefaults(spec, catalogue_service, "catalogue_service_proc", "catalogue_service_container")
	containers = append(containers, catalogue_service_ctr)

	frontend_service := workflow.Service[frontend.Frontend](spec, "frontend", user_service, catalogue_service, cart_service, order_service)
	frontend_service_ctr := applyHTTPDefaults(spec, frontend_service, "frontend_service_proc", "frontend_service_container")
	containers = append(containers, frontend_service_ctr)

	return containers, nil
}

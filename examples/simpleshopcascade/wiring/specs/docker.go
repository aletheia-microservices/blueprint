package specs

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/examples/simpleshopcascade/workflow/simpleshop"
	"github.com/blueprint-uservices/blueprint/plugins/cmdbuilder"
	"github.com/blueprint-uservices/blueprint/plugins/mongodb"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
)

var Docker = cmdbuilder.SpecOption{
	Name:        "docker",
	Description: "Deploys each service in a separate container with thrift, and uses mongodb as NoSQL database backends.",
	Build:       makeDockerSpec,
}

func makeDockerSpec(spec wiring.WiringSpec) ([]string, error) {
	var containers []string

	product_db := mongodb.Container(spec, "product_db")
	inventory_db := mongodb.Container(spec, "inventory_db")

	inventory_service := workflow.Service[simpleshop.InventoryService](spec, "inventory_service", inventory_db)
	inventory_service_ctr := applyHTTPDefaults(spec, inventory_service, "inventory_service_proc", "inventory_service_container")
	containers = append(containers, inventory_service_ctr)

	product_service := workflow.Service[simpleshop.ProductService](spec, "product_service", product_db, inventory_service)
	product_service_ctr := applyHTTPDefaults(spec, product_service, "product_service_proc", "product_service_container")
	containers = append(containers, product_service_ctr)

	return containers, nil
}

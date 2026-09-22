# Foo Bar

## Getting started

Prerequisites for this tutorial:
* [thrift compiler](https://thrift.apache.org/download) is installed
* docker is installed

## Compiling the application

To compile the application, we execute `wiring/main.go` and specify which wiring spec to compile. To view options and list wiring specs, run:

```
go run wiring/main.go -h
```

If you encounter errors because of missing modules that are supposed to be replaced by local ones, do:

```zsh
cd wiring
go clean -cache -modcache
export GOFLAGS=-mod=mod
export GOWORK=off
go mod tidy
cd ..
```

The following will compile the `docker` wiring spec to the directory `build`. This will fail if the pre-requisite thrift compiler is not installed.

```
rm -rf build
go run wiring/main.go -w docker -o build
```

## Running the application

To run the application, navigate to `build/docker` and run `docker compose up`. Use flag `--build` to build images if code is changed.

```zsh
docker-compose --env-file build/.env -f build/docker/docker-compose.yml up --build -d
```

In the end, don't forget to remove the containers.

```zsh
docker-compose --env-file build/.env -f build/docker/docker-compose.yml down
```

## Sending HTTP requests (examples)

This application has two services:
- `product_service`, which is the entry point and manages products
- `inventory_service`, which manages the inventory for those products

`ProductService.RegisterProduct` and `ProductService.DeleteProduct` each cascade into a call on `InventoryService` (`AddInventory` and `DeleteInventory` respectively), so a single request to `product_service` results in writes to both `product_db` and `inventory_db`.

Both services are deployed with their own HTTP address, so `inventory_service` can also be reached directly, independently of the cascade coming from `product_service`.

The examples below use the variables from the `.local.env` file to resolve each service's address instead of hardcoding ports.

```zsh
source build/.local.env
```

### 1. Register a product

Registering a product on `product_service` inserts it into `product_db`, and cascades a call to `inventory_service` that inserts the corresponding entry into `inventory_db`.

```zsh
curl "http://$PRODUCT_SERVICE_HTTP_DIAL_ADDR/RegisterProduct?id=sock-1&description=Blue+Sock&amount=10"
```

```zsh
# Read product
curl "http://$PRODUCT_SERVICE_HTTP_DIAL_ADDR/GetProduct?id=sock-1"

# Read the inventory entry created by the write cascade
curl "http://$INVENTORY_SERVICE_HTTP_DIAL_ADDR/GetInventory?id=sock-1"
```

### 2. Manage inventory directly

Since `inventory_service` is independently reachable, it can also be called directly, without going through `product_service`. This allows testing the cascade detection separately from a direct write.

```zsh
curl "http://$INVENTORY_SERVICE_HTTP_DIAL_ADDR/AddInventory?id=sock-2&amount=5"
curl "http://$INVENTORY_SERVICE_HTTP_DIAL_ADDR/GetInventory?id=sock-2"
```

### 3. Delete a product

Deleting a product from `product_service` removes it from `product_db` and triggers a call to `inventory_service` that removes the corresponding entry from `inventory_db`.

```zsh
curl "http://$PRODUCT_SERVICE_HTTP_DIAL_ADDR/DeleteProduct?id=sock-1"
```

```zsh
# (Attempt to) read the deleted product and inventory entry
curl "http://$PRODUCT_SERVICE_HTTP_DIAL_ADDR/GetProduct?id=sock-1"
curl "http://$INVENTORY_SERVICE_HTTP_DIAL_ADDR/GetInventory?id=sock-1"
```

### 4. Delete inventory directly

```zsh
curl "http://$INVENTORY_SERVICE_HTTP_DIAL_ADDR/DeleteInventory?id=sock-2"
```

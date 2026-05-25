# Digota

This is a Blueprint re-implementation of the [digota application](https://github.com/digota/digota).

Version used:
https://github.com/digota/digota/tree/c2a16d5

## Getting started

Prerequisites for this tutorial:
* [thrift compiler](https://thrift.apache.org/download) is installed
* docker is installed

## Running tests

```zsh
cd tests
go test
```

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
docker-compose --env-file build/.env -f build/docker/docker-compose.yml up --build
``` 

If you see Docker complain about missing environment variables, edit the `.env` file in `build/docker` and remove `0.0.0.0:` in all addresses. For example, `PRODUCT_SERVICE_HTTP_BIND_ADDR=0.0.0.0:12349` becomes `PRODUCT_SERVICE_HTTP_BIND_ADDR=12349`.

## Sending HTTP requests (examples)

### Product Service (port 12349)

```zsh
# Create a product
curl "http://localhost:12349/New?name=Widget&active=true&description=A+great+widget&shippable=true"

# Get a product by ID
curl "http://localhost:12349/Get?id=<product-id>"

# List products
curl "http://localhost:12349/List?page=0&limit=10&sort=0"

# Update a product
curl "http://localhost:12349/Update?id=<product-id>&name=Updated+Widget&active=false"

# Delete a product
curl "http://localhost:12349/Delete?id=<product-id>"
```

### SKU Service (port 12351)

```zsh
# Create a SKU (currency: 0=USD, price in cents)
curl "http://localhost:12351/New?name=Widget-SKU&currency=0&price=1999&active=true&parent=<product-id>"

# Get a SKU by ID
curl "http://localhost:12351/Get?id=<sku-id>"

# List SKUs
curl "http://localhost:12351/List?page=0&limit=10&sort=0"

# Update a SKU
curl "http://localhost:12351/Update?id=<sku-id>&price=2499&active=true"

# Delete a SKU
curl "http://localhost:12351/Delete?id=<sku-id>"
```

### Payment Service (port 12347)

```zsh
# Create a charge (currency: 0=USD, total in cents)
curl "http://localhost:12347/NewCharge?currency=0&total=2999&email=user@example.com&statement=Order+Payment&paymentProviderId=0"

# Get a charge by ID
curl "http://localhost:12347/Get?id=<charge-id>"

# List charges
curl "http://localhost:12347/List?page=0&limit=10&sort=0"

# Refund a charge (reason: 0=duplicate, 1=fraudulent, 2=requested_by_customer)
curl "http://localhost:12347/RefundCharge?id=<charge-id>&amount=2999&reason=2"
```

### Order Service (port 12345)

```zsh
# Create an order
curl -g "http://localhost:12345/New?currency=0&email=user@example.com&items=[{\"skuId\":\"<sku-id>\",\"quantity\":1}]&shipping={\"name\":\"John+Doe\",\"address\":{\"line1\":\"123+Main+St\",\"city\":\"New+York\",\"state\":\"NY\",\"country\":\"US\",\"postalCode\":\"10001\"}}&metadata={}"

# Get an order by ID
curl "http://localhost:12345/Get?id=<order-id>"

# List orders
curl "http://localhost:12345/List?page=0&limit=10&sort=0"

# Attempt to return an order
curl "http://localhost:12345/Return?id=<order-id>"
```

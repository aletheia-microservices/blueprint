# SockShop

This is a Blueprint re-implementation of the [SockShop application](https://github.com/ocp-power-demos/sock-shop-demo).

Version used:
https://github.com/ocp-power-demos/sock-shop-demo/tree/807a7ba

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

```zsh
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

```zsh
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

The examples below assume the Frontend service is reachable at `http://localhost:12349` (i.e. `FRONTEND_HTTP_BIND_ADDR=12349` in the `.env` file).

The requests below exemplify: register, log in, browse the catalogue, manage the cart, add an address and card, and finally create and inspect an order.

### 1. Register a new user

```zsh
RESPONSE=$(curl -s "http://localhost:12349/Register?sessionID=&username=alice&password=supersecret&email=alice@example.com&first=Alice&last=Smith") && echo $RESPONSE
export USER_ID=$(echo $RESPONSE | jq -r '.Ret0')
```

### 2. Log in

```zsh
curl "http://localhost:12349/Login?sessionID=&username=alice&password=supersecret"
```

```zsh
# Fetch the logged-in user's details
curl "http://localhost:12349/GetUser?userID=$USER_ID"
```

### 3. Browse the catalogue

The catalogue starts out empty.

```zsh
curl "http://localhost:12349/LoadCatalogueTags"
curl "http://localhost:12349/LoadCatalogueSocks"
```

```zsh
# List items
curl -g "http://localhost:12349/ListItems?tags=[]&order=&pageNum=1&pageSize=10"

# List items filtered by two tags ("blue" and "formal")
curl -g "http://localhost:12349/ListItems?tags=[\"blue\",\"formal\"]&order=&pageNum=1&pageSize=10"

# List all available tags
curl "http://localhost:12349/ListTags"
```

```zsh
# Get first item's ID from the list
RESPONSE=$(curl -s -g "http://localhost:12349/ListItems?tags=[]&order=&pageNum=1&pageSize=10") && echo $RESPONSE
export ITEM_ID=$(echo $RESPONSE | jq -r '.Ret0[0].id')
```

```zsh
# Get a single sock by ID
curl "http://localhost:12349/GetSock?itemID=$ITEM_ID"
```

### 4. Add items to the cart

```zsh
# Add an item to the cart
curl "http://localhost:12349/AddItem?sessionID=$USER_ID&itemID=$ITEM_ID"

# Get the current cart contents
curl "http://localhost:12349/GetCart?sessionID=$USER_ID"

# Update the quantity of an item already in the cart
curl "http://localhost:12349/UpdateItem?sessionID=$USER_ID&itemID=$ITEM_ID&quantity=3"

# Remove an item from the cart
curl "http://localhost:12349/RemoveItem?sessionID=$USER_ID&itemID=$ITEM_ID"

# (Attempt to) Get the current cart contents
curl "http://localhost:12349/GetCart?sessionID=$USER_ID"

# Add an item to the cart (again)
curl "http://localhost:12349/AddItem?sessionID=$USER_ID&itemID=$ITEM_ID"

# Get the current cart contents (again)
curl "http://localhost:12349/GetCart?sessionID=$USER_ID"
```

### 5. Add a shipping address

```zsh
RESPONSE=$(curl -s -g "http://localhost:12349/PostAddress?userID=$USER_ID&address={\"Street\":\"Baker+St\",\"Number\":\"221B\",\"Country\":\"UK\",\"City\":\"London\",\"PostCode\":\"NW1+6XE\"}") && echo $RESPONSE
export ADDRESS_ID=$(echo $RESPONSE | jq -r '.Ret0')
```

```zsh
# Get an address by ID
curl "http://localhost:12349/GetAddress?addressID=$ADDRESS_ID"
```

### 6. Add a payment card

```zsh
RESPONSE=$(curl -s -g "http://localhost:12349/PostCard?userID=$USER_ID&card={\"LongNum\":\"1234123412341234\",\"Expires\":\"04/26\",\"CCV\":\"123\"}") && echo $RESPONSE
export CARD_ID=$(echo $RESPONSE | jq -r '.Ret0')
```

```zsh
# Get a card by ID
curl "http://localhost:12349/GetCard?cardID=$CARD_ID"
```

### 7. Create an order

The `cartID` is the same `sessionID` used to build the cart in step 4.

```zsh
RESPONSE=$(curl -s "http://localhost:12349/NewOrder?userID=$USER_ID&addressID=$ADDRESS_ID&cardID=$CARD_ID&cartID=$USER_ID") && echo $RESPONSE
export ORDER_ID=$(echo $RESPONSE | jq -r '.Ret0.ID')
```

### 8. Inspect orders

```zsh
# Get order by ID
curl "http://localhost:12349/GetOrder?orderID=$ORDER_ID"

# Get all orders created by the user
curl "http://localhost:12349/GetOrders?userID=$USER_ID"
```

### Emptying the cart

```zsh
curl "http://localhost:12349/DeleteCart?sessionID=$USER_ID"
```

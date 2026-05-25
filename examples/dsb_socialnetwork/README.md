# DeathStarBench Social Network

This is a Blueprint re-implementation / translation of the [social-network application](https://github.com/delimitrou/DeathStarBench/tree/master/socialNetwork) from the DeathStarBench microservices benchmark.

Version used:
https://github.com/delimitrou/DeathStarBench/tree/6ecb097

The application provides a mostly-direct translation of the original code. In terms of the APIs and functionality, this implementation is intended to be as close to unmodified from the original as possible.

* [workflow](workflow) contains service implementations
* [tests](tests) has unit tests of the workflow
* [wiring](wiring) configures the application's topology and deployment and is used to compile the application

## Getting started

Prerequisites for this tutorial:
* [protoc](https://grpc.io/docs/protoc-installation/) is installed:
  `go install google.golang.org/protobuf/cmd/protoc-gen-go@latest`
  `go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest`
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

The following will compile the `docker` wiring spec to the directory `build`. This will fail if the pre-requisite protoc plugins are not installed.

```zsh
go run wiring/main.go -w docker -o build
```

## Running the application

To run the application, navigate to `build/docker` and run `docker compose up`. Use flag `--build` to build images if code is changed.

```zsh
docker-compose --env-file build/.env -f build/docker/docker-compose.yml up --build
```

If you see Docker complain about missing environment variables or servers not available (e.g. `rpc error: code = Unknown desc = memcache: no servers configured or available`), edit the `.env` file in `build/docker` and remove `0.0.0.0:` in all addresses. For example, `COMPOSEPOST_SERVICE_THRIFT_BIND_ADDR=0.0.0.0:9000` becomes `COMPOSEPOST_SERVICE_THRIFT_BIND_ADDR=9000`.

## Sending HTTP requests (examples)

All requests go through the `Wrk2APIService` HTTP gateway on port `12367`.

### Register a User

```zsh
curl "http://localhost:12367/Register?firstName=Alice&lastName=Doe&username=alicedoe&password=secret123&userId=1"
curl "http://localhost:12367/Register?firstName=Bob&lastName=Smith&username=bobsmith&password=secret456&userId=2"
```

### Follow and Unfollow

```zsh
# Follow by username
curl "http://localhost:12367/Follow?username=alicedoe&followeeName=bobsmith&userId=0&followeeID=0"

# Follow by user IDs
curl "http://localhost:12367/Follow?username=&followeeName=&userId=1&followeeID=2"

# Unfollow by username
curl "http://localhost:12367/Unfollow?username=alicedoe&followeeName=bobsmith&userId=0&followeeID=0"
```

### Compose a Post

```zsh
# Post with text only (no media)
curl 'http://localhost:12367/ComposePost?userId=1&username=alicedoe&post_type=0&text=Hello+from+the+social+network&media_types=%5B%5D&media_ids=%5B%5D'

# Post with an image
curl 'http://localhost:12367/ComposePost?userId=1&username=alicedoe&post_type=0&text=Check+out+this+photo&media_types=%5B%22png%22%5D&media_ids=%5B101%5D'
```

### Read Timelines

```zsh
# Read home timeline (posts from users that userId follows)
curl "http://localhost:12367/ReadHomeTimeline?userId=2&start=0&stop=10"

# Read user timeline (posts uploaded by userId)
curl "http://localhost:12367/ReadUserTimeline?userId=1&start=0&stop=10"
```

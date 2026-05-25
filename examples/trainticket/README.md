# Train Ticket

This is a Blueprint re-implementation / translation of the [Train Ticket application](https://github.com/FudanSELab/train-ticket/tree/master).

Version used:
https://github.com/FudanSELab/train-ticket/tree/313886e

The application provides a mostly-direct translation of the original code. In terms of the APIs and functionality, this implementation is intended to be as close to unmodified from the original as possible.

* [workflow](workflow) contains service implementations
* [tests](tests) has tests of the workflow
* [wiring](wiring) configures the application's topology and deployment and is used to compile the application

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

The following will compile the `docker` wiring spec to the directory `build`. 

```zsh
go run wiring/main.go -o build -w docker
```

## Running the application

To run the application, navigate to `build/docker` and run `docker compose up`. Use flag `--build` to build images if code is changed.

```zsh
docker-compose --env-file build/.env -f build/docker/docker-compose.yml up --build
``` 

If you see Docker complain about missing environment variables, edit the `.env` file in `build/docker` and remove `0.0.0.0:` in all addresses. For example, `ASSURANCE_SERVICE_HTTP_BIND_ADDR=0.0.0.0:9003` becomes `ASSURANCE_SERVICE_HTTP_BIND_ADDR=9003`.

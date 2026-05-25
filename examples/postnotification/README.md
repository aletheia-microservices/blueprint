# Post Notification

This is a Blueprint re-implementation of the [PostNotification application](https://github.com/Antipode-SOSP23/antipode-post-notification).

Version used:
https://github.com/Antipode-SOSP23/antipode-post-notification/tree/892a8c1

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
docker-compose --env-file build/.env -f build/docker/docker-compose.yml up --build
```

If you see Docker complain about missing environment variables, edit the `.env` file in `build/docker` and remove `0.0.0.0:` in all addresses. For example, `POSTS_DB_BIND_ADDR=0.0.0.0:12347` becomes `POSTS_DB_BIND_ADDR=12347`.

## Sending HTTP requests (examples)

### Upload Service (port 12349)

```zsh
# Upload a post
curl "http://localhost:12349/UploadPost?username=alice&text=Hello+world"

# Delete a post by ID
curl "http://localhost:12349/DeletePost?postID=<post-id>"
```

# DeathStarBench MediaMicroservices

This is a Blueprint re-implementation of the [media-microservices application](https://github.com/delimitrou/DeathStarBench/tree/master/mediaMicroservices).

Version used:
https://github.com/delimitrou/DeathStarBench/tree/6ecb097

* [workflow](workflow) contains service implementations
* [tests](tests) has tests of the workflow
* [wiring](wiring) configures the application's topology and deployment and is used to compile the application

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

If you see Docker complain about missing environment variables, edit the `.env` file in `build/docker` and remove `0.0.0.0:` in all addresses. For example, `MOVIE_ID_CACHE_BIND_ADDR=0.0.0.0:12351` becomes `MOVIE_ID_CACHE_BIND_ADDR=12351`.

## Sending HTTP requests (examples)

All requests go through the `APIService` HTTP gateway on port `12345`.

### Register a Movie

```zsh
# RegisterMovieId
curl "http://localhost:12345/RegisterMovieId?reqID=0&movieID=tt0111161&title=The+Shawshank+Redemption"
curl "http://localhost:12345/RegisterMovieId?reqID=1&movieID=tt0068646&title=The+Godfather"
```

### Movie Info

```zsh
# WriteMovieInfo
curl "http://localhost:12345/WriteMovieInfo?reqID=0&movieID=tt0111161&title=The+Shawshank+Redemption&plotID=1&numRating=0"
curl "http://localhost:12345/WriteMovieInfo?reqID=1&movieID=tt0068646&title=The+Godfather&plotID=2&numRating=0"
```

### Cast Info

```zsh
# WriteCastInfo
curl "http://localhost:12345/WriteCastInfo?reqID=0&castInfoID=cast001&name=Tim+Robbins&gender=Male&intro=American+actor+and+filmmaker"
curl "http://localhost:12345/WriteCastInfo?reqID=1&castInfoID=cast002&name=Morgan+Freeman&gender=Male&intro=American+actor+and+narrator"
```

### Plot

```zsh
# WritePlot
curl "http://localhost:12345/WritePlot?reqID=0&plotID=1&plotText=Two+imprisoned+men+bond+over+a+number+of+years+finding+solace+and+eventual+redemption"
curl "http://localhost:12345/WritePlot?reqID=1&plotID=2&plotText=The+aging+patriarch+of+an+organized+crime+dynasty+transfers+control+to+his+reluctant+son"
```

### User

```zsh
# RegisterUser
curl "http://localhost:12345/RegisterUser?reqID=req001&firstName=John&lastName=Doe&username=johndoe&password=secret123"

# Login
curl "http://localhost:12345/Login?reqID=0&username=johndoe&password=secret123"
```

### Compose a Review

A review is composed by uploading all its parts (each with the same `reqID`). When all 5 components arrive, the review is automatically persisted. Use a fresh `reqID` that has not been used by any prior upload call.

`UploadMovieId` counts as **2 components** internally (movie ID + rating), so exactly 4 curl calls are needed.

```zsh
# generate a unique review ID (count=1)
curl "http://localhost:12345/UploadUniqueId?reqID=2"

# associate the movie and rating (counts as 2 of the 5 components; count=3)
curl "http://localhost:12345/UploadMovieId?reqID=2&title=The+Shawshank+Redemption&rating=5"

# associate the user (count=4)
curl "http://localhost:12345/UploadUserWithUsername?reqID=2&username=johndoe"

# upload the review text (count=5 => review is stored)
curl "http://localhost:12345/UploadText?reqID=2&text=An+absolute+masterpiece+of+cinema"
```

### Read a Page

```zsh
curl "http://localhost:12345/ReadPage?reqID=0&title=The+Shawshank+Redemption&reviewStart=0&reviewStop=10"
```

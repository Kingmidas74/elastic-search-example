FROM golang:1.16.4-alpine AS build_base

RUN apk add --no-cache git

# Set the Current Working Directory inside the container
WORKDIR /tmp/go-sample-app

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

# Build the Go app
RUN go build -o ./out/go-sample-app .


# Start fresh from a smaller image
FROM alpine:3.9
RUN apk add ca-certificates


COPY --from=build_base /tmp/go-sample-app/out/go-sample-app /app/go-sample-app
COPY --from=build_base /tmp/go-sample-app/posts.csv /app/posts.csv

# This container exposes port 8080 to the outside world
EXPOSE 8010

# Run the binary program produced by `go install`
CMD ["/app/go-sample-app"]
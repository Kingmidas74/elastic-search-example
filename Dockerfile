FROM golang:1.16.4-alpine AS build_base

RUN apk add --no-cache git

WORKDIR /tmp/go-sample-app

#RUN export PATH=$(go env GOPATH)/bin:$PATH

COPY go.mod .
COPY go.sum .
RUN go mod download
RUN go get -u github.com/alecthomas/template
RUN go get github.com/swaggo/swag/cmd/swag


COPY . .
RUN swag init -g ./models/app.go
RUN go build -o ./out/go-sample-app .


FROM alpine:3.9
RUN apk add ca-certificates

ENV APP_PORT 8010
EXPOSE 8010

COPY --from=build_base /tmp/go-sample-app/out/go-sample-app /app/go-sample-app
COPY --from=build_base /tmp/go-sample-app/docs /app/docs
COPY --from=build_base /tmp/go-sample-app/posts.csv /app/posts.csv

CMD ["/app/go-sample-app"]
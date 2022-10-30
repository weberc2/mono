FROM golang:1.19

WORKDIR build
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN go build -o /bin/main

WORKDIR run
ENTRYPOINT /bin/main

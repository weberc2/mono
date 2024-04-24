FROM golang:1.22

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY ./cmd/ ./cmd/
RUN ls && go build -o /bin/main ./cmd/gobuilder

WORKDIR /run
ENTRYPOINT /bin/main
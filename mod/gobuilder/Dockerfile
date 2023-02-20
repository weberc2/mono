FROM golang:1.19

ARG TARGET

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY ./cmd/ ./cmd/
RUN ls && go build -o /bin/main ./cmd/${TARGET}

WORKDIR /run
ENTRYPOINT /bin/main
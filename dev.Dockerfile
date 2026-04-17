FROM golang:1.24-alpine AS builder

ARG PORTS
ENV PORTS 8080 40000

RUN go install github.com/go-delve/delve/cmd/dlv@latest

WORKDIR /build
ENV GO111MODULE=on
COPY go.mod .
COPY go.sum .
RUN go mod download
RUN go mod tidy
COPY . .
RUN CGO_ENABLED=0 go build -o ./bin/app

# Build runtime

FROM golang:1.24-alpine
COPY --from=builder /build/bin/app /
COPY --from=builder /go/bin/dlv /
COPY config*.yaml /
EXPOSE ${PORTS}
WORKDIR /
CMD ["/app"]



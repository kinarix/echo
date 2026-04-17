FROM golang:1.23-alpine AS builder

RUN go install github.com/go-delve/delve/cmd/dlv@latest

ARG PORTS
ENV PORTS 8080 40000
WORKDIR /build

ENV GO111MODULE=on
COPY go.mod .
COPY go.sum .
RUN go mod download
RUN go mod tidy
COPY . .
RUN CGO_ENABLED=0 go build -o ./bin/app
RUN pwd && ls -al

# Build runtime

FROM golang:1.23-alpine
COPY --from=builder /build/bin/app /
COPY --from=builder /go/bin/dlv /
COPY config*.yaml /
RUN apk add curl bash
EXPOSE ${PORTS}
WORKDIR /
CMD ["/app"]



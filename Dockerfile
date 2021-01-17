FROM golang:1.15.6-alpine3.12
RUN apk add --no-cache git gcc musl-dev
WORKDIR /build
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o /app /build/cmd/main.go

FROM alpine:3.12
RUN apk add --no-cache ca-certificates tzdata
COPY --from=0 /app /onlinestat
CMD ["/onlinestat"]

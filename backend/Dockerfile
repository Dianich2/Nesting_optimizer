FROM golang:1.24 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o nesting_optimizer ./cmd/api

FROM alpine:3.22
WORKDIR /app
COPY --from=builder /app/nesting_optimizer .
EXPOSE 8082
ENTRYPOINT [ "./nesting_optimizer" ]
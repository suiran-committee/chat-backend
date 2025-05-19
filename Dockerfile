FROM golang:1.22 AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/server

FROM scratch
WORKDIR /app
COPY --from=builder /app/server .
COPY cert.pem .
COPY key.pem .
EXPOSE 8443
ENTRYPOINT ["./server"]

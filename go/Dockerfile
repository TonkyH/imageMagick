FROM golang:1.21-bookworm as builder

WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . ./
# Build the binary.
RUN CGO_ENABLED=0 go build -v -o server

FROM alpine:3
RUN apk add --no-cache imagemagick
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/server .

CMD ["/server"]

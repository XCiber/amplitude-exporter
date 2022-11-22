# builder image
FROM golang:1.19-alpine as builder
RUN apk --no-cache add ca-certificates
RUN mkdir /build
WORKDIR /build
COPY go.* ./
RUN go mod download
COPY . ./
# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o amplitude-exporter

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/amplitude-exporter /app/amplitude-exporter
EXPOSE 8080
WORKDIR /app
ENTRYPOINT ["/app/amplitude-exporter"]
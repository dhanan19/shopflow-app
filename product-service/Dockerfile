FROM golang:1.26-alpine AS build
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o /product-service ./cmd/server
FROM alpine:3.22
RUN adduser -D appuser
COPY --from=build /product-service /product-service
USER appuser
EXPOSE 8080
ENTRYPOINT ["/product-service"]

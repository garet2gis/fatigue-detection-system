# Step 1: Modules caching
FROM golang:1.21-alpine as modules
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

# Step 2: Builder
FROM golang:1.21-alpine as builder
COPY --from=modules /go/pkg /go/pkg
COPY . /model_handler_service
WORKDIR /model_handler_service
RUN CGO_ENABLED=0 GOOS=linux go build -o ./.bin/main ./cmd/app/

# Step 3: Final
FROM alpine:3.19

COPY --from=builder /model_handler_service/.bin/main .
COPY --from=builder /model_handler_service/migrations ./migrations
COPY --from=builder /model_handler_service/thresholds.json ./thresholds.json

ENTRYPOINT ["./main"]

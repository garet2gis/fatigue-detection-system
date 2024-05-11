# Step 1: Modules caching
FROM golang:1.21-alpine as modules
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

# Step 2: Builder
FROM golang:1.21-alpine as builder
COPY --from=modules /go/pkg /go/pkg
COPY . /user_data_service
WORKDIR /user_data_service
RUN CGO_ENABLED=0 GOOS=linux go build -o ./.bin/main ./cmd/app/

# Step 3: Final
FROM alpine:3.19

COPY --from=builder /user_data_service/.bin/main .
COPY --from=builder /user_data_service/migrations ./migrations

ENTRYPOINT ["./main"]

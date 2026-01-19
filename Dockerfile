FROM golang:1.25-alpine AS builder

ARG BIN_NAME="subscription-service"
ARG MAIN_PATH="cmd/main.go"
ARG GO_BUILD_FLAGS="-s -w"
# ARG UPX_FLAGS="--best --lzma"

WORKDIR /app

RUN apk add --no-cache git
# RUN apk add --no-cache upx

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/pressly/goose/v3/cmd/goose@latest
RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY . .

RUN swag init -o docs -d cmd,internal/models,internal/api

RUN CGO_ENABLED=0 go build -ldflags="$GO_BUILD_FLAGS" -o $BIN_NAME $MAIN_PATH
# RUN upx $UPX_FLAGS $BIN_NAME

FROM alpine:latest

ARG BIN_NAME="subscription-service"

WORKDIR /root/

COPY --from=builder /app/$BIN_NAME .
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/example_config.yaml ./config.yaml
COPY --from=builder /app/docs ./docs

EXPOSE 8080

CMD ["./subscription-service"]

FROM golang:1.18-alpine3.17 AS builder
WORKDIR /tmp/pharmacy-user
ADD . .
RUN apk add make && GOPATH=$(pwd)/cache make build

FROM alpine:3.17.2 AS app
RUN apk add bind-tools && apk add curl
COPY --from=builder /tmp/pharmacy-user/bin/pharmacy_user /app/pharmacy_user
COPY ./migrations /app/migrations/
COPY ./configs /app/configs/
WORKDIR /app
ENTRYPOINT ["./pharmacy_user", "-conf", "./configs"]
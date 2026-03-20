FROM golang:1.23-alpine AS build

WORKDIR /app
COPY go.mod ./
COPY cmd ./cmd
COPY templates ./templates
COPY static ./static

RUN go build -o /qiandao ./cmd/server

FROM alpine:3.20

WORKDIR /app
ENV APP_ADDR=:8080

COPY --from=build /qiandao /usr/local/bin/qiandao
COPY templates ./templates
COPY static ./static

RUN mkdir -p /app/data

EXPOSE 8080
CMD ["qiandao"]

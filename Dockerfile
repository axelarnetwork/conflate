FROM golang:1.18-alpine3.16 AS build

WORKDIR /code

COPY . .

ARG VERSION=devel
RUN go build -ldflags "-X main.version=$VERSION" -o /code/releases/conflate ./conflate     

FROM alpine:3.16

WORKDIR /app
COPY --from=build /code/releases/ .

ENTRYPOINT [ "/app/conflate" ]
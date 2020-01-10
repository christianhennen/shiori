FROM golang:1.13-alpine as builder

RUN apk update && apk --no-cache add git build-base
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

# ========== END OF BUILDER ========== #

FROM alpine:latest

RUN apk update && apk --no-cache add dumb-init ca-certificates
COPY --from=builder /app/main /usr/local/bin/shiori

ENV SHIORI_DIR /srv/shiori/
EXPOSE 8080

ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["/usr/local/bin/shiori", "serve", "--ldap", "/srv/shiori/conf/ldap.toml"]

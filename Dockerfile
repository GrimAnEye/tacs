FROM golang:1.22.1 as builder

# Building a static binary without debugging information
WORKDIR /app
COPY . /app

RUN go mod download
RUN \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build -ldflags "-s -w" -v -o tacs .

# Forming a container for launch
FROM scratch

ENV TACS_SCHEME=scheme.yaml \
    TACS_ADDR= \
    TACS_PORT=8080 \
    TACS_CERT= \
    TACS_CERT_KEY= \
    TACS_LDAP_SERVER= \
    TACS_LDAP_PORT=389 \
    TACS_LDAP_CERT= \
    TACS_LDAP_KEY= \
    TACS_LDAP_USER= \
    TACS_LDAP_PASSWORD=

COPY --from=builder /app/tacs /

EXPOSE 8080
CMD ["/tacs"]
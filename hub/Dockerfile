# build:
# docker build -t go-hub -f hub/Dockerfile .

FROM golang:1.17-alpine as app-builder
WORKDIR /go/src/app
COPY ./lib ./lib
COPY ./hub ./hub
# Static build required so that we can safely copy the binary over.
# `-tags timetzdata` embeds zone info from the "time/tzdata" package.
RUN cd ./hub && mkdir -p /go/src/app/www && CGO_ENABLED=0 go install -ldflags '-s -w -extldflags "-static"' -tags timetzdata

FROM scratch
WORKDIR /

# the tls certificates:
# NB: this pulls directly from the upstream image, which already has ca-certificates:
COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# the program:
COPY --from=app-builder /go/bin/hub /

#EXPOSE 8080:8080
ENTRYPOINT ["/hub"]


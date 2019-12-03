FROM golang:alpine

RUN apk add --update bash curl git && \
    curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
WORKDIR /go/src/github.com/analogj/lodestone-publisher/

CMD /go/src/github.com/analogj/lodestone-publisher/build.sh

#GOOS=linux GARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags "-X main.goos=linux -X main.goarch=amd64 -extldflags \"-static\"" -o lodestone-document-processor-linux-amd64 ./cmd/document-processor/document-processor.go && \
#GOOS=linux GARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags "-X main.goos=linux -X main.goarch=amd64 -extldflags \"-static\"" -o lodestone-thumbnail-processor-linux-amd64 ./cmd/thumbnail-processor/thumbnail-processor.go
#
#
#FROM debian:buster-slim as runtime
#
#RUN apt-get update && apt-get install -y bash curl git go-dep libmagickwand-6.q16-dev libreoffice-common
#COPY --from=builder /go/src/github.com/analogj/lodestone-processor/lodestone-document-processor-linux-amd64 /usr/bin/lodestone-document-processor
#COPY --from=builder /go/src/github.com/analogj/lodestone-processor/lodestone-thumbnail-processor-linux-amd64 /usr/bin/lodestone-thumbnail-processor
#
#RUN chmod +x /usr/bin/lodestone-document-processor /usr/bin/lodestone-thumbnail-processor

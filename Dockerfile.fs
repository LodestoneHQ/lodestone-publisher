FROM golang:1.15-buster AS build

RUN apt-get update && apt-get install -y --no-install-recommends bash curl git
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

WORKDIR /go/src/github.com/analogj/lodestone-publisher/

ADD Gopkg.toml ./
ADD Gopkg.lock ./
ADD cmd ./cmd
ADD pkg ./pkg

RUN dep ensure
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o lodestone-fs-publisher ./cmd/fs-publisher/fs-publisher.go

FROM scratch
COPY --from=build /go/src/github.com/analogj/lodestone-publisher/lodestone-fs-publisher /lodestone-fs-publisher
CMD ["/lodestone-fs-publisher"]

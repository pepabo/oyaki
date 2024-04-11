FROM golang:1.21-bullseye AS build

ARG OYAKI_VERSION

WORKDIR /go/src/oyaki
COPY . /go/src/oyaki

RUN apt update && apt install -y libvips-dev
RUN go build -ldflags "-s -w -X main.version=${OYAKI_VERSION}" -o /go/bin/oyaki

FROM debian:bookworm

RUN apt update && apt install -y libvips-dev

COPY --from=build /go/bin/oyaki /

EXPOSE 8080

CMD ["/oyaki"]

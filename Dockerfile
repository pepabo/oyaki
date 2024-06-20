FROM golang:1.22-bullseye AS build

ARG OYAKI_VERSION

WORKDIR /go/src/oyaki
COPY . /go/src/oyaki

RUN CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=${OYAKI_VERSION}" -o /go/bin/oyaki
RUN apt update && apt install -y curl \
 && curl https://storage.googleapis.com/downloads.webmproject.org/releases/webp/libwebp-1.3.1-linux-x86-64.tar.gz --output libwebp.tar.gz \
 && tar vzxf libwebp.tar.gz \
 && mv libwebp-1.3.1-linux-x86-64/bin/cwebp /go/bin/

FROM gcr.io/distroless/static-debian11

COPY --from=build /go/bin/oyaki /
COPY --from=build /go/bin/cwebp /bin/

EXPOSE 8080

CMD ["/oyaki"]

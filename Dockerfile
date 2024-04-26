FROM golang:1.21-bullseye AS build

ARG OYAKI_VERSION

WORKDIR /go/src/oyaki
COPY . /go/src/oyaki

ARG TARGETARCH="arm64"
ENV LIBWEBP_VERSION=1.3.1
ENV BASE_URL=https://storage.googleapis.com/downloads.webmproject.org/releases/webp
RUN apt-get update && apt-get install -y curl tar
RUN if [ "${TARGETARCH}" = "amd64" ]; then \
        ARCH='amd64'; \
    elif [ "${TARGETARCH}" = "arm64" ]; then \
        ARCH='aarch64'; \
    else \
        echo "Unsupported arch: ${TARGETARCH}"; exit 1; \
    fi && \
    curl https://storage.googleapis.com/downloads.webmproject.org/releases/webp/libwebp-1.3.1-linux-${ARCH}.tar.gz --output libwebp.tar.gz \
    tar -zxf libwebp.tar.gz --strip-components=1 -C /go/bin && \
    rm libwebp.tar.gz

RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -ldflags "-s -w -X main.version=${OYAKI_VERSION}" -o /go/bin/oyaki

FROM gcr.io/distroless/static-debian11

COPY --from=build /go/bin/oyaki /
COPY --from=build /go/bin/cwebp /bin/

EXPOSE 8080

CMD ["/oyaki"]

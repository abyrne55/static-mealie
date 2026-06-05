FROM --platform=$BUILDPLATFORM quay.io/hummingbird/go:1.26-builder AS builder
ARG TARGETARCH

WORKDIR /build
COPY go.mod .
COPY . .

RUN CGO_ENABLED=0 GOARCH=$TARGETARCH go build \
    -ldflags "-s -w" \
    -o /static-mealie .

FROM quay.io/hummingbird/static:latest

WORKDIR /

COPY --from=builder /static-mealie /usr/bin/static-mealie

USER 65532

ENTRYPOINT ["/usr/bin/static-mealie"]

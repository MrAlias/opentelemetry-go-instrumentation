FROM  --platform=$BUILDPLATFORM golang:1.23.2-bookworm AS base

RUN apt-get update && apt-get install -y curl clang gcc llvm make libbpf-dev

WORKDIR /app

COPY sdk/ /app/sdk/

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading
# them in subsequent builds if they change
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg \
    go mod download && go mod verify

FROM --platform=$BUILDPLATFORM base AS builder
COPY . .

ARG TARGETARCH
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    GOARCH=$TARGETARCH make build

FROM gcr.io/distroless/base-debian12@sha256:6ae5fe659f28c6afe9cc2903aebc78a5c6ad3aaa3d9d0369760ac6aaea2529c8
COPY --from=builder /app/otel-go-instrumentation /
CMD ["/otel-go-instrumentation"]

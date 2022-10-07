# get newest certificates
FROM --platform=$BUILDPLATFORM alpine:3.16 AS ca-certs
RUN apk add --no-cache ca-certificates
RUN --mount=type=cache,target=/etc/ssl/certs \
    update-ca-certificates 2>/dev/null || true

# zig compiler
FROM --platform=$BUILDPLATFORM ghcr.io/euantorano/zig:master AS zig-env

# build environment
FROM --platform=$BUILDPLATFORM golang:1-alpine AS build

# required arguments(buildx will set target)
ARG VERSION
ARG BUILD_TIME
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

# set working directory
WORKDIR /go/src

# download packages
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg \
    go mod download

# add source
COPY . .

# setup go+zig
COPY --from=zig-env /usr/local/bin/zig /usr/local/bin/zig
ENV PATH="/usr/local/bin/zig:${PATH}"
RUN --mount=type=cache,target=/go/pkg \
    go install github.com/dosgo/zigtool/zigcc@latest && \
    go install github.com/dosgo/zigtool/zigcpp@latest && \
    go env -w GOARM=${TARGETVARIANT##*v}
ENV CC="zigcc" \
    CXX="zigcpp" \
    CGO_ENABLED=0 \
    GOOS="linux" \
    GOARCH=$TARGETARCH

# build binary 
RUN --mount=type=bind,target=. \
    --mount=type=cache,target=/root/.cache/go-build \ 
    --mount=type=cache,target=/go/pkg \
    go build \
    -v \
    -o /bin/omadasitedns

RUN chmod 1001 /bin/omadasitedns

FROM scratch

COPY --from=ca-certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /bin/omadasitedns /omadasitedns

USER 1001

EXPOSE 53

ENTRYPOINT ["/omadasitedns"]
# build environment
FROM --platform=$BUILDPLATFORM ghcr.io/kwitsch/ziggoimg:development AS build

# download packages
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg \
    go mod download

# add source
COPY . .

# build binary 
RUN --mount=type=bind,target=. \
    --mount=type=cache,target=/root/.cache/go-build \ 
    --mount=type=cache,target=/go/pkg \
    go build \
    -v \
    -o /bin/omadasitedns

RUN apk add --no-cache libcap && \
    setcap 'cap_net_bind_service=+ep' /bin/omadasitedns && \
    chown 1001 /bin/omadasitedns && \
    chmod u+x /bin/omadasitedns

FROM scratch

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /bin/omadasitedns /omadasitedns

USER 1001

EXPOSE 53

ENTRYPOINT ["/omadasitedns"]
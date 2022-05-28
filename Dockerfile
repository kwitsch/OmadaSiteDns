FROM ghcr.io/kwitsch/docker-buildimage:main AS build-env

ADD src .
RUN gobuild.sh -o omadasitedns

FROM scratch
COPY --from=build-env /builddir/omadasitedns /omadasitedns

EXPOSE 53

ENTRYPOINT ["/omadasitedns"]
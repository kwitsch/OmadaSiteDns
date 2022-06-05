FROM scratch
COPY /omadasitedns /omadasitedns

EXPOSE 53

ENTRYPOINT ["/omadasitedns"]
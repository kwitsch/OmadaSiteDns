FROM scratch
COPY /etc/ssl/certs /usr/local/share/ca-certificates
COPY /omadasitedns /omadasitedns


EXPOSE 53

ENTRYPOINT ["/omadasitedns"]
FROM golang:alpine AS builder

RUN adduser \    
    --disabled-password \    
    --gecos "" \    
    --home "/nonexistent" \    
    --shell "/sbin/nologin" \    
    --no-create-home \    
    --uid "10001" \    
    "appuser"

COPY /omadasitedns /omadasitedns

RUN chown appuser /omadasitedns && \
    chown :appuser /omadasitedns

FROM scratch
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

COPY --from=builder /omadasitedns /omadasitedns

USER appuser:appuser

EXPOSE 53

ENTRYPOINT ["/omadasitedns"]
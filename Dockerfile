FROM almalinux:10
RUN useradd -u 10001 txlog

FROM scratch
ARG VERSION
ARG BUILD_DATE
LABEL org.opencontainers.image.title="Txlog Server"
LABEL org.opencontainers.image.description="The server component serves as a centralized system that manages the PostgreSQL database server, functioning as a repository for transaction data while efficiently handling incoming information from multiple agent instances throughout the network."
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.created="${BUILD_DATE}"
LABEL org.opencontainers.image.source="https://github.com/txlog/server"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.vendor="rda.run"
COPY --from=0 /etc/passwd /etc/passwd
COPY --from=0 /etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem /etc/ssl/certs/ca-certificates.crt
COPY bin/txlog-server /bin/txlog-server
USER txlog
CMD ["/bin/txlog-server"]

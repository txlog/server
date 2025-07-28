FROM scratch
LABEL org.opencontainers.image.source=https://github.com/txlog/server
LABEL org.opencontainers.image.description="The server component serves as a centralized system that manages the PostgreSQL database server, functioning as a repository for transaction data while efficiently handling incoming information from multiple agent instances throughout the network."
LABEL org.opencontainers.image.licenses=MIT
COPY bin/txlog-server /bin/txlog-server
CMD ["/bin/txlog-server"]

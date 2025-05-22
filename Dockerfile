FROM alpine:3.21
LABEL org.opencontainers.image.source=https://github.com/txlog/server
LABEL org.opencontainers.image.description="The server component serves as a centralized system that manages the PostgreSQL database server, functioning as a repository for transaction data while efficiently handling incoming information from multiple agent instances throughout the network."
LABEL org.opencontainers.image.licenses=MIT
RUN apk upgrade --no-cache
COPY bin/txlog-server /bin/txlog-server
RUN addgroup -S txlog && adduser -S -G txlog txlog
USER txlog
CMD ["/bin/txlog-server"]

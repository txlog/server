FROM alpine:3.21
COPY bin/txlog-server /bin/txlog-server
RUN addgroup -S txlog && \
    adduser -S -G txlog txlog
USER txlog
CMD ["/bin/txlog-server"]

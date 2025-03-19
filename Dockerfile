FROM scratch
COPY bin/txlog-server /bin/txlog-server
RUN addgroup -S txlog && \
    adduser -S -G txlog txlog
USER txlog
CMD ["/bin/txlog-server"]

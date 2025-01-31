FROM ghcr.io/almalinux/9-micro:9
COPY bin/txlog-server /bin/txlog-server
CMD ["/bin/txlog-server"]

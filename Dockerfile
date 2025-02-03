FROM gcr.io/distroless/static-debian12
COPY bin/txlog-server /bin/txlog-server
CMD ["/bin/txlog-server"]

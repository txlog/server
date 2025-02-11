FROM scratch
COPY bin/txlog-server /bin/txlog-server
CMD ["/bin/txlog-server"]

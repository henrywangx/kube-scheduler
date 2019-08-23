FROM alpine:3.7
COPY ./scheduler /bin
COPY ./run.sh /bin

CMD ["/bin/sh", "/bin/run.sh"]
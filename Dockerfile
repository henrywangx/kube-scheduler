FROM alpine:3.7
COPY ./kube-scheduler /bin

ENTRYPOINT ["/bin/scheduler"]
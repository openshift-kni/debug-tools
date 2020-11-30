FROM registry.access.redhat.com/ubi8/ubi-minimal:latest
COPY _output /usr/local/bin
# no tool is more important than others
ENTRYPOINT ["/bin/sh"]

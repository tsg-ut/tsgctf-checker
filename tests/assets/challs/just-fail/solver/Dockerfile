FROM ubuntu:latest

RUN apt-get clean
RUN apt-get update
RUN apt-get install -y python2 curl

COPY exploit /exploit
WORKDIR /

ENTRYPOINT ["/bin/bash", "exploit"]

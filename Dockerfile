FROM golang:1.15-alpine

WORKDIR /tiwengo
ADD . /tiwengo

RUN cd /tiwengo && go build

EXPOSE 6080 9000
CMD ./tiwengo

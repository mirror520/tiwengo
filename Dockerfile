FROM golang:1.14-alpine

RUN apk --no-cache add gcc

WORKDIR /tiwengo
ADD . /tiwengo

RUN cd /tiwengo && go build

EXPOSE 6080
CMD tiwengo

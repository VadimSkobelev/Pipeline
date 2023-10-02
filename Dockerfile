FROM golang:latest AS compiling_stage
RUN mkdir -p /go/src/pipeline
WORKDIR /go/src/pipeline
ADD main.go .
ADD go.mod .
RUN go install .
 
FROM alpine:latest
LABEL version="1.0"
LABEL maintainer="Vadim Skobelev"
WORKDIR /root/
COPY --from=compiling_stage /go/bin/pipeline .
ENTRYPOINT ./pipeline
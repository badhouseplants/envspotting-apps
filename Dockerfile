# Building
FROM golang:1.16.5-alpine3.13  as builder
# create a working directory
WORKDIR /go/src/app
# add source code
COPY . /go/src/app
RUN apk add git --no-cache
RUN go build -o main


FROM  alpine:3.14
WORKDIR /root/
# copy artifacts from builder
COPY --from=builder /go/src/app/main .
# run server
CMD ["./main"]
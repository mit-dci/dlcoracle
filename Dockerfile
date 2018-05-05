
FROM golang:alpine as build

RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh gcc musl-dev
ENV GOROOT=/usr/local/go
RUN go get github.com/awnumar/memguard 
RUN go get github.com/boltdb/bolt
RUN go get github.com/gorilla/mux
RUN go get github.com/gorilla/handlers
RUN rm -rf /usr/local/go/src/github.com/gertjaap/dlcoracle
COPY . /usr/local/go/src/github.com/gertjaap/dlcoracle
WORKDIR /usr/local/go/src/github.com/gertjaap/dlcoracle
RUN go get -v ./...
RUN go build

FROM alpine
RUN apk add --no-cache ca-certificates
COPY --from=build /usr/local/go/src/github.com/gertjaap/dlcoracle/dlcoracle /app/bin/dlcoracle
EXPOSE 3000
RUN mkdir /app/bin/data
WORKDIR /app/bin

CMD ["./dlcoracle"]
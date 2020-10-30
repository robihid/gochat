FROM golang:1.14.10-alpine3.12

RUN mkdir /server
ADD . /server
WORKDIR /server

RUN go mod download

RUN go build -o main .

CMD ["/server/main"]
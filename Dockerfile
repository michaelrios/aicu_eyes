FROM golang:1.10.1-alpine3.7
RUN apk add --no-cache git
RUN apk add --update make
RUN go get github.com/golang/dep/cmd/dep

RUN go get -u -d gocv.io/x/gocv

RUN cd /go/src/gocv.io/x/gocv
RUN make install

RUN mkdir -p /go/src/github.com/michaelrios/aicu_eyes
WORKDIR /go/src/github.com/michaelrios/aicu_eyes

COPY Gopkg.lock Gopkg.toml /go/src/github.com/michaelrios/aicu_eyes/
RUN dep ensure -vendor-only

COPY . /go/src/github.com/michaelrios/aicu_eyes/

RUN go build main.go

EXPOSE 8080
CMD ["./main"]

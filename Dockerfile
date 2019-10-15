FROM golang:alpine as builder

WORKDIR /bin/app
COPY . .
RUN apk add git build-base
RUN go get github.com/gin-gonic/gin
RUN go get github.com/go-redis/redis
RUN go get github.com/satori/go.uuid
RUN go get go.mongodb.org/mongo-driver/mongo
RUN go get go.mongodb.org/mongo-driver/mongo/options

RUN go build -o app *.go

FROM alpine
COPY --from=builder /bin/app/app /bin
CMD ["app"]
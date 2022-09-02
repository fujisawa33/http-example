# 開発環境ホットリロード
FROM golang:1.18.5-buster
WORKDIR /go/src/app

COPY ./. .

RUN go mod download

CMD [ "go", "run", "main.go" ]

FROM golang:1.22.5

WORKDIR /src
COPY . .

RUN go build && touch .env

CMD ["./lcp-v2"]
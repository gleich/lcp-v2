# syntax=docker/dockerfile:1
FROM golang:1.22.5 AS build

WORKDIR /src
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/lcp ./main.go && touch .env

FROM alpine

RUN apk add --no-cache ca-certificates

COPY --from=build /bin/lcp /bin/lcp
COPY --from=build /src/.env ./.env

CMD ["/bin/lcp"]

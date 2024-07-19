FROM golang:1.22.5 AS build

WORKDIR /src
COPY . .

RUN go build -o /bin/lcp

FROM scratch

COPY --from=build /bin/lcp /bin/lcp

CMD ["/bin/lcp"]
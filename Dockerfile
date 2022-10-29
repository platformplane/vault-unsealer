# builder
FROM golang:1-alpine as builder

WORKDIR /src
COPY . .

RUN go mod download
RUN go build -o unsealer


FROM alpine:3

RUN apk add --no-cache tini

WORKDIR /
COPY --from=builder /src/unsealer .

ENTRYPOINT ["tini", "--"]
CMD [ "/unsealer" ]
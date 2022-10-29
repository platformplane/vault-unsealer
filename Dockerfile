# builder
FROM golang:1.19-alpine3.16 as builder

WORKDIR /src
COPY . .

RUN go mod download
RUN go build -o unsealer


FROM alpine:3.16

RUN apk add --no-cache tini

WORKDIR /
COPY --from=builder /src/unsealer .

ENTRYPOINT ["tini", "--"]
CMD [ "/unsealer" ]
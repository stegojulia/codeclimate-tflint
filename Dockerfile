FROM golang:1.18-alpine as builder

LABEL maintainer="Stefano Tenuta <stapps@outlook.it>"

WORKDIR /usr/src/app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . ./
RUN go build

FROM ghcr.io/terraform-linters/tflint

RUN adduser -u 9000 -D app
COPY --from=builder /usr/src/app/codeclimate-tflint /usr/src/app/codeclimate-tflint
RUN chown -R app:app /usr/src/app

USER app

VOLUME /code
WORKDIR /code

ENTRYPOINT [""]
CMD ["/usr/src/app/codeclimate-tflint"]
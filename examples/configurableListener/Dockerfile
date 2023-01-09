FROM golang:1.18

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./
COPY configurableListener.yaml ./

RUN go build -o /wrp-listener-example

EXPOSE 7100

CMD ["/wrp-listener-example"]
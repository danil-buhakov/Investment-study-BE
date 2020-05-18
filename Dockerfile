FROM golang:latest

RUN mkdir /app
WORKDIR /app
COPY . .
ENV CGO_ENABLED 0
RUN go build


CMD ["/app/ethERC20"]

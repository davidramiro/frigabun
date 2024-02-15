FROM golang:1.22

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /frigabun
EXPOSE 9595

CMD ["/frigabun"]
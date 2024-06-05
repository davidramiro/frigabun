FROM golang:1.22 AS build-stage

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /frigabun

FROM gcr.io/distroless/static-debian11 AS build-release-stage

COPY --from=build-stage /frigabun /frigabun

WORKDIR /app
EXPOSE 9595

CMD ["/frigabun"]

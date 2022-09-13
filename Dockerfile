FROM golang:1.18-alpine as build

RUN apk --no-cache add git

RUN mkdir /app
ADD . /app
WORKDIR /app

# get deps
RUN go mod init github.com/cliveyg/poptape-categories
RUN go mod tidy
RUN go mod download

# need these flags or alpine image won't run due to dynamically linked libs in binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-w' -o categories


FROM alpine:latest

RUN mkdir -p /categories
COPY --from=build /app/categories /categories
COPY --from=build /app/.env /categories
WORKDIR /categories

# Make port 8220 available to the world outside this container
EXPOSE 8220

# Run reviews binary when the container launches
CMD ["./categories"]

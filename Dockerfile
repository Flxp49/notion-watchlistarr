FROM golang:1.22-alpine as build
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o /notionwatchlistarr cmd/notionwatchlistarr/docker/main.go
EXPOSE 7879
FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /notionwatchlistarr /notionwatchlistarr
CMD [ "/notionwatchlistarr" ]
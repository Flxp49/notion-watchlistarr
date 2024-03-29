FROM golang:1.22-alpine as builder
WORKDIR /app
COPY . .
RUN go mod download
# RUN go build -ldflags "-H windowsgui" -o /notionwatchlistarrsync cmd/notionwatchlistarrsync/main.go
RUN go build -o /notionwatchlistarrsync cmd/notionwatchlistarrsync/main.go
EXPOSE 7879
# FROM scratch
# COPY --from=build /app/ /app/
CMD [ "/notionwatchlistarrsync" ]
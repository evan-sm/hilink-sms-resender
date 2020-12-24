FROM golang:alpine AS build

RUN apk add --update --no-cache tzdata

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o hilink-sms-resender

FROM scratch

ENV TZ Europe/Moscow

COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/group /etc/group
COPY --from=build /app/hilink-sms-resender /app/hilink-sms-resender


USER 1000:1000

ENTRYPOINT ["/app/hilink-sms-resender"]

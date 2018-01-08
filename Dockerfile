FROM alpine

RUN apk add -U bash
RUN apk --no-cache add ca-certificates && update-ca-certificates
ADD ./cmd/bot/bot /app/

CMD [“/app/bot]

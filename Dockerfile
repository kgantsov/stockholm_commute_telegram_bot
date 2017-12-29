FROM alpine

RUN apk add -U bash
RUN apk --no-cache add ca-certificates && update-ca-certificates
ADD ./stockholm_commute_bot /app/

CMD [“/app/stockholm_commute_bot”]

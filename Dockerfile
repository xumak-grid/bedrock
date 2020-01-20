FROM alpine:3.6

RUN apk add --no-cache --update ca-certificates
COPY ./bin/api /bin/
CMD ["/bin/api"]

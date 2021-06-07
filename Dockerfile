FROM golang:alpine AS build_base

# WE NEED GIT
RUN apk update && apk add --no-cache git

COPY src/ /src
RUN ls src
WORKDIR /src
RUN go install github.com/mitranim/gow@latest
RUN go build -o /api-gateway

CMD ["gow", "run", "."]

#USER appuser
FROM alpine:3.9

RUN mkdir /app
RUN mkdir /etc/certificates

RUN apk add openssl

# GENERATE CERTS
RUN openssl genrsa -out /etc/certificates/fly.rsa 2048
RUN openssl rsa -in /etc/certificates/fly.rsa -pubout > /etc/certificates/fly.rsa.pub

COPY --from=build_base api-gateway /app/api-gateway
COPY --from=build_base /src/model.conf /model.conf
#COPY --from=build_base /etc/certificates/fly.rsa /etc/certificates/fly.rsa 
#COPY --from=build_base /etc/certificates/fly.rsa.pub /etc/certificates/fly.rsa.pub


# CHANGE RIGHTS ON CERTS
RUN adduser -S  -D appuser
RUN chown appuser /etc/certificates/fly.rsa && chown appuser /etc/certificates/fly.rsa.pub && chown appuser /app/api-gateway

RUN chmod +x /app/api-gateway

# USE THE APPLICATION-USER
USER appuser

CMD ["/app/api-gateway"]
EXPOSE 61225

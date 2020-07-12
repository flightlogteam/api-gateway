FROM golang:alpine AS build_base

# WE NEED GIT
RUN apk update && apk add --no-cache git

RUN go get -u github.com/flightlogteam/api-gateway/src


#USER appuser
FROM alpine:3.9

RUN mkdir /app

RUN apk add openssl

# GENERATE CERTS
RUN openssl genrsa -out fly.rsa 2048
RUN openssl rsa -in fly.rsa -pubout > fly.rsa.pub

COPY --from=build_base /go/bin/api-gateway /app/

WORKDIR /app


# CHANGE RIGHTS ON CERTS
RUN adduser -S  -D appuser
RUN chown appuser /fly.rsa && chown appuser /fly.rsa.pub && chown appuser ./api-gateway

RUN chmod +x /app/api-gateway

# USE THE APPLICATION-USER
USER appuser

# SET ENVIRONMENT-VARIABLES
# override these while running image
ARG DATABASE_PASSWORD=""
ENV DATABASE_PASSWORD="${DATABASE_PASSWORD}"

ARG DATABASE_USERNAME="root"
ENV DATABASE_USERNAME="${DATABASE_USERNAME}"

ARG DATABASE_HOSTNAME="192.168.1.99"
ENV DATABASE_HOSTNAME="${DATABASE_HOSTNAME}"

ARG DATABASE_PORT="3306"
ENV DATABASE_PORT="${DATABASE_PORT}"

# IF EMPTY SERVICE API WILL START WITH LIMITED FEATURES
ARG USERSERVICE_URL=""
ENV USERSERVICE_URL="${USERSERVICE_URL}"

CMD ["/app/api-gateway"]
EXPOSE 61225

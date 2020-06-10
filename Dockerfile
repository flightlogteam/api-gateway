FROM golang:alpine
RUN mkdir /app
WORKDIR /app
COPY src .

RUN go version
RUN cat /etc/resolv.conf

## BUILD THE CODE
RUN go build -o gateway .

RUN apk add openssl


# GENERATE CERTIFICATES
RUN openssl genrsa -out fly.rsa 2048
RUN openssl rsa -in fly.rsa -pubout > fly.rsa.pub

RUN adduser -S -D -H -h /app appuser
RUN chown appuser fly.rsa && chown appuser fly.rsa.pub

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
ARG USESSERVICE_URL=""
ENV USESSERVICE_URL="${USESSERVICE_URL}"

CMD ["./gateway"]
EXPOSE 61225

FROM golang:1.23-alpine as builder

#DECLARE ENVIRONMENT VARIABLES HERE
ARG GO_ENV
ENV GO_ENV=${GO_ENV}
ARG REDIS_URL
ENV REDIS_URL=${REDIS_URL}
ARG REDIS_CERT
ENV REDIS_CERT=${REDIS_CERT}
ARG DB_HOST
ENV DB_HOST=${DB_HOST}
ARG DB_PORT
ENV DB_PORT=${DB_PORT}
ARG DB_USER
ENV DB_USER=${DB_USER}
ARG DB_PASSWORD
ENV DB_PASSWORD=${DB_PASSWORD}
ARG DB_NAME
ENV DB_NAME=${DB_NAME}

RUN echo $TZ

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . ./

RUN go build -v -o bin/playworker


FROM alpine
# Install any required dependencies.
RUN apk --no-cache add ca-certificates curl

WORKDIR /root/

COPY --from=builder /app/bin/playworker /usr/local/bin/

CMD ["playworker"]
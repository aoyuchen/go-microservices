FROM alpine:latest

RUN mkdir /app

COPY mailApp /app
COPY ./cmd/api/templates /templates

CMD [ "/app/mailApp" ]
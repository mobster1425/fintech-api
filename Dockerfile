
# Build stage
FROM golang:1.21-alpine3.18 AS builder
 RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY . .
RUN go build -o main main.go
#RUN CGO_ENABLED=0 go build -o main main.go




# Run stage
FROM alpine:3.18
RUN apk add --no-cache ca-certificates openssl
RUN rm -f /etc/ssl/certs/smtp.gmail.com.crt
RUN openssl s_client -starttls smtp -connect smtp.gmail.com:587 </dev/null 2>/dev/null | openssl x509 -outform PEM > /etc/ssl/certs/smtp.gmail.com.crt
RUN update-ca-certificates
WORKDIR /app
COPY --from=builder /app/main .
COPY app.env .
COPY start.sh .
COPY wait-for.sh .
COPY db/migrations ./db/migrations

 



EXPOSE 8080
EXPOSE 587 
#9090
CMD [ "/app/main" ]
ENTRYPOINT [ "/app/start.sh" ]







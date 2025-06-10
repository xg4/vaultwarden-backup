FROM alpine:latest

RUN apk add --no-cache \
    sqlite \
    bash \
    openssl

COPY backup.sh /

COPY entrypoint.sh /

RUN chmod +x /entrypoint.sh && \
    chmod +x /backup.sh

RUN echo "0 */6 * * * /backup.sh" > /etc/crontabs/root

ENTRYPOINT ["/entrypoint.sh"]
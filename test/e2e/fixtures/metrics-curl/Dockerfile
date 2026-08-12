FROM alpine:3.23@sha256:fd791d74b68913cbb027c6546007b3f0d3bc45125f797758156952bc2d6daf40
RUN apk add --no-cache curl \
    && adduser -D -u 1000 -g 1000 metrics
USER 1000

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
ARG TARGETPLATFORM
COPY ${TARGETPLATFORM}/mindweaver /usr/local/bin/

# Container defaults - all data under /data for easy volume mounting
ENV MW_DATA_DIR=/data

# Create data directory
RUN mkdir -p /data

EXPOSE 9421
ENTRYPOINT ["mindweaver"]
CMD ["--mode=combined"]

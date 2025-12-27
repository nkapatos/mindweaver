FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
ARG TARGETPLATFORM
COPY ${TARGETPLATFORM}/mindweaver /usr/local/bin/
EXPOSE 9421
ENTRYPOINT ["mindweaver"]
CMD ["--mode=combined"]

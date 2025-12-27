FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
COPY mindweaver /usr/local/bin/
EXPOSE 9421
ENTRYPOINT ["mindweaver"]
CMD ["--mode=combined"]

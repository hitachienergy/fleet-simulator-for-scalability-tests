FROM golang:1.20.5-buster

# Get certificates
RUN apt-get update && apt-get install -y ca-certificates openssl
ARG cert_location=/usr/local/share/ca-certificates
RUN openssl s_client -showcerts -connect github.com:443 </dev/null 2>/dev/null|openssl x509 -outform PEM > ${cert_location}/github.crt
RUN openssl s_client -showcerts -connect proxy.golang.org:443 </dev/null 2>/dev/null|openssl x509 -outform PEM >  ${cert_location}/proxy.golang.crt
RUN update-ca-certificates

WORKDIR /app

COPY . .
COPY docker/run.sh .

RUN go build -o simulator .

ENV PLUGIN_PATH /app/client.so
ENV PLUGIN_DIR /app/plugin
ENV NEED_COMPILE true
ENV CONFIG ""
ENV CLIENT_NUM 0
ENV OPATH /app/results
ENV IDX_OFFSET 0
ENV INFLUENCE "{}"
ENV STATUS_SERVER_PORT 8086

RUN mkdir $OPATH

CMD [ "/bin/bash", "run.sh"]
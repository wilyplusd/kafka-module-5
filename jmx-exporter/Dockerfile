FROM alpine:3.19

RUN apk add --no-cache openjdk11-jre-headless curl bash

WORKDIR /app

RUN curl -sL -o jmx_prometheus_httpserver.jar \
    "https://repo1.maven.org/maven2/io/prometheus/jmx/jmx_prometheus_httpserver/0.20.0/jmx_prometheus_httpserver-0.20.0.jar"

EXPOSE 9404

CMD ["java", "-jar", "jmx_prometheus_httpserver.jar", "9404", "/config/config.yml"]

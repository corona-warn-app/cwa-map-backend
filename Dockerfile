FROM golang:1.19-alpine as build
WORKDIR /backend/build

RUN apt-get update && apt-get install unzip

# Download and extract liquibase
RUN curl -o flyway-commandline-7.9.1-linux-x64.tar.gz https://repo1.maven.org/maven2/org/flywaydb/flyway-commandline/7.9.1/flyway-commandline-7.9.1-linux-x64.tar.gz \
       && tar -xf flyway-commandline-7.9.1-linux-x64.tar.gz \
       && rm flyway-commandline-7.9.1-linux-x64.tar.gz

RUN curl -o vault_1.7.2_linux_386.zip https://releases.hashicorp.com/vault/1.7.2/vault_1.7.2_linux_386.zip \
    && unzip vault_1.7.2_linux_386.zip \
    && rm vault_1.7.2_linux_386.zip

# Compile application
COPY src app
RUN cd app && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo ./cmd/backend

FROM debian:latest
WORKDIR /opt/backend

# copy app, flyway and vault from previous stage
COPY --from=build /backend/build/app/backend app/
COPY --from=build /backend/build/flyway-7.9.1 flyway
COPY --from=build /backend/build/vault vault

# copy resources
COPY resources/db db/
COPY resources/start.sh .

RUN apt-get update && apt-get install -y ca-certificates && \
    chmod +x flyway/flyway && chmod +x start.sh

CMD ["./start.sh"]
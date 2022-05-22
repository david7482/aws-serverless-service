# build stage
FROM golang:1.17.6 as builder_go

RUN apt-get update \
    && apt-get --no-install-recommends -y install \
    	ca-certificates \
	&& apt-get clean \
    && rm -rf /var/lib/apt/lists/*

COPY . /src
WORKDIR /src
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} make build

# final stage
FROM debian:11-slim

RUN adduser --disabled-login chatbot
USER chatbot

COPY --from=builder_go /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder_go /src/bin/chatbot /chatbot

CMD ["/chatbot"]
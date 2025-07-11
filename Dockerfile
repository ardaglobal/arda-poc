###############  Stage 1 – build  ###############
FROM golang:1.24-alpine AS builder
ARG IGNITE_VERSION=v28.10.0

RUN apk add --no-cache build-base git curl bash
RUN curl -L "https://get.ignite.com/cli@${IGNITE_VERSION}!" | bash

WORKDIR /src
COPY . .

# put ./build on the PATH so `which` can see the binary
ENV PATH="/src/build:${PATH}"

RUN ignite chain build && ignite chain init --home .arda-poc

# figure out where the node binary is *at runtime* and copy it out
RUN BIN=$(which arda-pocd 2>/dev/null || true)  \
 && if [ -z "$BIN" ]; then echo "arda-pocd not found on PATH"; exit 1; fi \
 && install -Dm755 "$BIN" /out/arda-pocd \
 && cp -r .arda-poc /out/.arda-poc-template

###############  Stage 2 – runtime ###############
FROM alpine:3.19
RUN apk update && apk add --no-cache bash curl

# binary + template only
COPY --from=builder /out/arda-pocd          /usr/local/bin/arda-pocd
COPY --from=builder /out/.arda-poc-template /template/.arda-poc

# copy in entrypoint script
COPY scripts/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

# declare persistent state location
VOLUME ["/data"]
ENV ARDA_HOME=/data/.arda-poc
WORKDIR /data

EXPOSE 26657 1317 9090 4500
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

FROM ubuntu

ARG ARCH=amd64
ARG OS=linux

RUN apt-get update && apt-get install -y curl && \
    curl -L https://github.com/weberc2/futhorc/releases/download/latest/futhorc-${ARCH}-${OS}-v0.1.6 -o /bin/futhorc && \
    chmod +x /bin/futhorc

WORKDIR /workspace

COPY . .

ARG FUTHORC_PROFILE=release
RUN /bin/futhorc build --profile ${FUTHORC_PROFILE} --output /blog

FROM caddy:2.4.5

COPY --from=0 /blog /usr/share/caddy

CMD caddy run --config /etc/caddy/Caddyfile

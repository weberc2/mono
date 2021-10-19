FROM rustlang/rust:nightly

ARG ARCH=amd64
ARG OS=linux

RUN cargo install futhorc

WORKDIR /workspace

COPY . .

ARG FUTHORC_PROFILE=release
RUN /bin/futhorc build --profile ${FUTHORC_PROFILE} --output /blog

FROM caddy:2.4.5

COPY --from=0 /blog /usr/share/caddy

CMD caddy run --config /etc/caddy/Caddyfile

FROM rustlang/rust:nightly

RUN cargo install futhorc

WORKDIR /workspace

COPY . .

ARG FUTHORC_PROFILE=release
RUN futhorc build --profile ${FUTHORC_PROFILE} --output /blog

FROM caddy:2.4.5

COPY --from=0 /blog /usr/share/caddy
COPY --from=0 /workspace/Caddyfile /etc/caddy/Caddyfile

CMD caddy run --config /etc/caddy/Caddyfile

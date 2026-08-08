ARG OPENTOFU_VERSION=1.12.5
FROM ghcr.io/opentofu/opentofu:${OPENTOFU_VERSION}-minimal AS tofu

FROM alpine:3.22
COPY --from=tofu /usr/local/bin/tofu /usr/local/bin/tofu
COPY scripts/run-engine.sh /compat/run-engine.sh
RUN chmod 0555 /usr/local/bin/tofu /compat/run-engine.sh

USER 65532:65532
HEALTHCHECK NONE
ENTRYPOINT ["/compat/run-engine.sh"]

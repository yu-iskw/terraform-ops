ARG TERRAFORM_VERSION=1.15.8
FROM hashicorp/terraform:${TERRAFORM_VERSION}

COPY scripts/run-engine.sh /compat/run-engine.sh
RUN chmod 0555 /compat/run-engine.sh

ENTRYPOINT ["/compat/run-engine.sh"]

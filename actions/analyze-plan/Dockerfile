FROM golang:1.24.4-alpine AS build

RUN apk add --no-cache git ca-certificates
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/terraform-ops ./cmd/terraform-ops

FROM alpine:3.22
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/terraform-ops /app/terraform-ops
COPY actions/analyze-plan/entrypoint.sh /app/entrypoint.sh
RUN chmod 0555 /app/terraform-ops /app/entrypoint.sh

HEALTHCHECK NONE
ENTRYPOINT ["/app/entrypoint.sh"]

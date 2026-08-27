ARG GO_VERSION=1.25
ARG ALPINE_VERSION=3.20

FROM golang:${GO_VERSION} AS build
ENV BIN_FILE=/opt/calendar/app

WORKDIR /go/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG APP=calendar
ARG LDFLAGS

RUN CGO_ENABLED=0 go build \
        -ldflags "${LDFLAGS}" \
        -o ${BIN_FILE} \
        ./cmd/${APP}

FROM alpine:${ALPINE_VERSION}
RUN apk add --no-cache ca-certificates gettext

ENV APP_BIN=/opt/calendar/app
ENV CONFIG_TEMPLATE=/etc/calendar/config.template.toml
ENV CONFIG_FILE=/etc/calendar/config.toml

COPY --from=build ${APP_BIN} ${APP_BIN}

ARG CONFIG=configs/calendar_config.template.toml
COPY ${CONFIG} ${CONFIG_TEMPLATE}

COPY build/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh && sed -i 's/\r$//' /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
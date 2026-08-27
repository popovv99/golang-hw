#!/bin/sh
set -e

: "${CONFIG_TEMPLATE:=/etc/calendar/config.template.toml}"
: "${CONFIG_FILE:=/etc/calendar/config.toml}"

# Список переменных, которые envsubst будет заменять.
# Остальные ${...} в шаблоне он тронуть не должен.
VARS='$LOGGER_LEVEL $STORAGE_TYPE $DB_HOST $DB_PORT $DB_USER $DB_PASSWORD $DB_NAME $DB_SCHEMA $SERVER_HOST $SERVER_PORT $KAFKA_BROKERS $KAFKA_TOPIC $KAFKA_GROUP_ID $KAFKA_RETRY $KAFKA_TIMEOUT $SCHEDULER_INTERVAL'

envsubst "$VARS" < "$CONFIG_TEMPLATE" > "$CONFIG_FILE"
exec "$APP_BIN" --config "$CONFIG_FILE" "$@"
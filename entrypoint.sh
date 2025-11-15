#!/bin/sh

# Entrypoint params

## DB check
DB_HOST="${OA_COCKPIT_DB_HOST:-postgres}"
DB_PORT="${OA_COCKPIT_DB_PORT:-5432}"
TIMEOUT_SECONDS=2

## runtime environment type (development/stagingproduction/ - defaults to development)
ENV_TYPE="${ENV_TYPE:-development}"

# avoid crashing the app due to warm-up delay of the DB instance
echo "checking for DB availability"
while true
do
  if nc -w $TIMEOUT_SECONDS "$DB_HOST" "$DB_PORT" </dev/null
  then
    echo "DB connection check is successful!"
    break
  else
    echo "waiting for DB connection to be available..."
    sleep 1
  fi
done

# run DB migrations
./deviceregistry  --env="$ENV_TYPE" migrate up --allow-missing

# run the application in normal mode
./deviceregistry  --env="$ENV_TYPE"

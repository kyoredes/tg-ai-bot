#!/bin/sh
set -e

/bin/subscription-migrate
exec /bin/subscription-service

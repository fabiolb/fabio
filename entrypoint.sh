#!/bin/sh
set -e

if [ "${1:0:1}" = '-' ]; then
    set -- /usr/bin/fabio "$@"
fi

exec "$@"


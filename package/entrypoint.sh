#!/bin/bash

if [ -n "$TARGET" ]; then
  cp /usr/bin/giddyup ${TARGET}
fi

exec "$@"

#! /usr/bin/env bash
CURDIR=$(cd $(dirname $0); pwd)
echo "$CURDIR/bin/user-service"
exec "$CURDIR/bin/user-service"
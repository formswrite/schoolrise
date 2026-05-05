#!/bin/sh
set -e

if [ -z "$POSTGRES_PASSWORD" ]; then
  echo "POSTGRES_PASSWORD must be set" >&2
  exit 1
fi

cat > /etc/pgbouncer/userlist.txt <<EOF
"schoolrise" "${POSTGRES_PASSWORD}"
EOF
chmod 600 /etc/pgbouncer/userlist.txt

exec pgbouncer /etc/pgbouncer/pgbouncer.ini

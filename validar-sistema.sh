#!/bin/sh
#
# Execute this after clients have finished and server is still online.

docker start server >/dev/null

clientData=$(mktemp)
{
  for agencyId in $(seq 5); do
      tr -d "\r" < "client/.data/agency-$agencyId.csv" |
      sed "s/^/$agencyId,/"
  done
} | sort > "$clientData"

serverData=$(mktemp)
docker exec server cat bets.csv | sort > "$serverData"

cmp "$clientData" "$serverData" > /dev/null && echo "OK" || echo "ERROR"

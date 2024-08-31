#!/bin/sh
#
# Execute this after clients have finished and server is still online.

clientData=$(mktemp)
{
  for agencyId in $(seq 5); do
      tr -d "\r" < "client/.data/agency-$agencyId.csv" |
      sed "s/^/$agencyId,/"
  done
} | sort > "$clientData"

serverData=$(mktemp)
docker exec server cat bets.csv | sort > "$serverData"

diff "$clientData" "$serverData"

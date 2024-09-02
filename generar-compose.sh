#!/bin/sh

if [ "$#" -ne 2 ]; then
  echo "Generates a compose.yaml file with a dynamic number of clients"
  echo
  echo "Usage: $0 <output> <clients>"
  exit 1
fi
OUTPUT=$1
CLIENTS=$2

# validate $CLIENTS is a number
if ! [ "$CLIENTS" -eq "$CLIENTS" ] 2>/dev/null; then
  echo "$CLIENTS is not a number"
  exit 1
fi

append() {
  echo "$1" >> "$OUTPUT"
}

# truncate file
: > "$OUTPUT"

append "name: tp0"
append "services:"
append "  server:"
append "    container_name: server"
append "    image: server:latest"
append "    entrypoint: /server"
append "    volumes:"
append "      - ./server/config.ini:/config.ini"
append "    networks:"
append "      - testing_net"
append ""

for i in $(seq 1 "$CLIENTS"); do
append "  client$i:"
append "    container_name: client$i"
append "    image: client:latest"
append "    entrypoint: /client"
append "    environment:"
append "      - CLI_ID=$i"
append "    volumes:"
append "      - ./client/config.yaml:/config.yaml"
append "    networks:"
append "      - testing_net"
append "    depends_on:"
append "      - server"
append ""
done

append "networks:"
append "  testing_net:"
append "    ipam:"
append "      driver: default"
append "      config:"
append "        - subnet: 172.25.125.0/24"

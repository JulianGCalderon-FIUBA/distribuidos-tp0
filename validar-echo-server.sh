#!/bin/sh

docker run --network tp0_testing_net -it alpine sh -c \
'
echo ping > input
nc server 12345 < input > output

if ! diff input output >/dev/null; then
  echo "action: test_echo_server | result: fail"
else
  echo "action: test_echo_server | result: success"
fi
'

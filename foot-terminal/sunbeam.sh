#!/bin/bash

SUNBEAM='foot -c /tmp/sunbeam.ini sunbeam'
RUNNING=$(pgrep -fa "sunbeam.ini")
if [[ $RUNNING ]]; then
  echo "sunbeam is running"
else
  echo "starting sunbeam"
  cp ./foot-terminal/sunbeam.ini /tmp/sunbeam.ini
  $SUNBEAM
fi

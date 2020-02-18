#!/bin/bash
echo "Starting entrypoint"
if [ $# -eq 0 ]
then
  /root/.rbenv/shims/passenger start
else
  bash -c $*
fi

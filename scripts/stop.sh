#!/usr/bin/env bash

PID=$(ps aux | grep "[t]ccbot-backend" | awk '{print $2}')
if [ -z "$PID" ]
then
    echo "tccbot-backend not runned"
else
    kill $PID
    echo "tccbot-backend stopped"
fi
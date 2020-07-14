#!/usr/bin/env bash

HOMEDIR=/home/tccbot
WORKDIR=$HOMEDIR/tccbot-backend

cd $WORKDIR && make build-usr
$WORKDIR/bin/tccbot-backend -config $HOMEDIR/config-yaml/config-local.yaml -level 5 -test -logdir $HOMEDIR/logs &

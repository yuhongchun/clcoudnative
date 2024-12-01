#!/usr/bin/env bash

basedir=${BDOS_SDKBUILD_HOME}
timestr=$(date +%Y%m%d%H%M)
commitID=`echo ${GIT_COMMIT} | cut -c1-8`


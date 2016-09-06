#!/bin/bash

. /etc/sysconfig/keepalived-vip
kube-keepalived-vip ${CONFIGMAP} ${SERVER} ${USESERVICE} ${USEADDRESSES} ${USESERVICEPORT} ${USELINKADDRESS} >> /data/log/keepalived-vip.log 2>&1 &

sleep 1
ret=$(ps -ef|grep -v grep|grep -c keepalived|grep -q 4)

if [[ ${ret} -eq 0 ]];then
    exit 0
else
    exit 1
fi

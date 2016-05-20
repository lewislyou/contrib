# kube-keepalived-vip
Kubernetes Virtual IP address/es using [keepalived](http://www.keepalived.org)

AKA "how to set up virtual IP addresses in kubernetes using [IPVS - The Linux Virtual Server Project](http://www.linuxvirtualserver.org/software/ipvs.html)".

## Disclaimer:
- This is a **work in progress**.

## Overview

In order to expose service use the same VIP.

## Requirements

[Daemonsets](https://github.com/kubernetes/kubernetes/blob/master/docs/design/daemon.md) enabled is the only requirement. Check this [guide](https://github.com/kubernetes/kubernetes/blob/master/docs/api.md#enabling-resources-in-the-extensions-group) with the required flags in kube-apiserver.


## Configuration

To expose one or more services use the flag `services-configmap`. The format of the data is: `external IP -> namespace/serviceName`. Optionally is possible to specify forwarding method using `:` after the service name. The valid options are `NAT` and `DR`. For instance `external IP -> namespace/serviceName:DR`.
By default if the method is not specified it will use NAT.

This IP must be routable inside the LAN and must be available. 
By default the IP address of the pods are used to route the traffic. This means that is one pod dies or a new one is created by a scale event the keepalived configuration file will be updated and reloaded.

## Example

```
$ echo "apiVersion: v1
kind: ConfigMap
metadata:
  name: vip-configmap
data:
  10.4.0.50-1: default/service1
  10.4.0.50-2: default/service2
  10.4.0.50:   default/service3" | kubectl create -f -
```
note: either ip-index(10.4.0.50-1) or ip(10.4.0.50) is ok. then, one vip 10.4.0.50 to multiple services

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

Now the creation of the daemonset
```
$ kubectl create -f vip-daemonset.yaml
daemonset "kube-keepalived-vip" created
$ kubectl get daemonset
NAME                  CONTAINER(S)          IMAGE(S)                         SELECTOR                        NODE-SELECTOR
kube-keepalived-vip   kube-keepalived-vip   gcr.io/google_containers/kube-keepalived-vip:0.7   name in (kube-keepalived-vip)   type=worker
```

**Note: the daemonset yaml file contains a node selector. This is not required, is just an example to show how is possible to limit the nodes where keepalived can run**

To verify if everything is working we should check if a `kube-keepalived-vip` pod is in each node of the cluster
```
$ kubectl get nodes
NAME       LABELS                                        STATUS    AGE
10.4.0.3   kubernetes.io/hostname=10.4.0.3,type=worker   Ready     1d
10.4.0.4   kubernetes.io/hostname=10.4.0.4,type=worker   Ready     1d
10.4.0.5   kubernetes.io/hostname=10.4.0.5,type=worker   Ready     1d
```

```
$ kubectl get pods
NAME                        READY     STATUS    RESTARTS   AGE
echoheaders-co4g4           1/1       Running   0          5m
kube-keepalived-vip-a90bt   1/1       Running   0          53s
kube-keepalived-vip-g3nku   1/1       Running   0          52s
kube-keepalived-vip-gd18l   1/1       Running   0          54s
```

```
$ kubectl logs kube-keepalived-vip-a90bt
I0410 14:24:45.860119       1 keepalived.go:161] cleaning ipvs configuration
I0410 14:24:45.873095       1 main.go:109] starting LVS configuration
I0410 14:24:45.894664       1 main.go:119] starting keepalived to announce VIPs
Starting Healthcheck child process, pid=17
Starting VRRP child process, pid=18
Initializing ipvs 2.6
Registering Kernel netlink reflector
Registering Kernel netlink reflector
Registering Kernel netlink command channel
Registering gratuitous ARP shared channel
Registering Kernel netlink command channel
Using LinkWatch kernel netlink reflector...
Using LinkWatch kernel netlink reflector...
I0410 14:24:56.017590       1 keepalived.go:151] reloading keepalived
Got SIGHUP, reloading checker configuration
Registering Kernel netlink reflector
Initializing ipvs 2.6
Registering Kernel netlink command channel
Registering gratuitous ARP shared channel
Registering Kernel netlink reflector
Opening file '/etc/keepalived/keepalived.conf'.
Registering Kernel netlink command channel
Opening file '/etc/keepalived/keepalived.conf'.
Using LinkWatch kernel netlink reflector...
VRRP_Instance(vips) Entering BACKUP STATE
Using LinkWatch kernel netlink reflector...
Activating healthchecker for service [10.2.68.5]:8080
VRRP_Instance(vips) Transition to MASTER STATE
VRRP_Instance(vips) Entering MASTER STATE
VRRP_Instance(vips) using locally configured advertisement interval (1000 milli-sec)
```

```
$ kubectl exec kube-keepalived-vip-a90bt cat /etc/keepalived/keepalived.conf

global_defs {
  vrrp_version 3
  vrrp_iptables KUBE-KEEPALIVED-VIP
}

vrrp_instance vips {
  state BACKUP
  interface eth1
  virtual_router_id 50
  priority 100
  nopreempt
  advert_int 1

  track_interface {
    eth1
  }



  virtual_ipaddress {
    172.17.4.90
  }
}


# Service: default/echoheaders
virtual_server 10.4.0.50 80 {
  delay_loop 5
  lvs_sched wlc
  lvs_method NAT
  persistence_timeout 1800
  protocol TCP


  real_server 10.2.68.5 8080 {
    weight 1
    TCP_CHECK {
      connect_port 8080
      connect_timeout 3
    }
  }

}

```


```
$ curl -v 10.4.0.50
* Rebuilt URL to: 10.4.0.50/
*   Trying 10.4.0.50...
* Connected to 10.4.0.50 (10.4.0.50) port 80 (#0)
> GET / HTTP/1.1
> Host: 10.4.0.50
> User-Agent: curl/7.43.0
> Accept: */*
>
* HTTP 1.0, assume close after body
< HTTP/1.0 200 OK
< Server: BaseHTTP/0.6 Python/3.5.0
< Date: Wed, 30 Dec 2015 19:52:39 GMT
<
CLIENT VALUES:
client_address=('10.4.0.148', 52178) (10.4.0.148)
command=GET
path=/
real path=/
query=
request_version=HTTP/1.1

SERVER VALUES:
server_version=BaseHTTP/0.6
sys_version=Python/3.5.0
protocol_version=HTTP/1.0

HEADERS RECEIVED:
Accept=*/*
Host=10.4.0.50
User-Agent=curl/7.43.0
* Closing connection 0

```

Scaling the replication controller should update and reload keepalived

```
note: either ip-index(10.4.0.50-1) or ip(10.4.0.50) is ok. then, one vip 10.4.0.50 to multiple services

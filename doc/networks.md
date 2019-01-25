## Implement Mydocker Networks Using Bridge

This document will explain what have been done when creating a new network (take the default network `mydocker0` for example, this is created by mydocker automatically) and connecting the network to a container.

### create a linux bridge named mydocker0

```bash
$ ip link add mydocker0 type bridge
$ ip addr show mydocker0
6: mydocker0: <BROADCAST,MULTICAST> mtu 1500 qdisc noop state DOWN group default qlen 1000
    link/ether 16:6e:09:ca:1e:de brd ff:ff:ff:ff:ff:ff
```

### set ip addr (gateway) on the bridge mydocker0

```bash
$ ip addr add 10.20.30.1 dev mydocker0
$ ip addr show mydocker0
6: mydocker0: <BROADCAST,MULTICAST> mtu 1500 qdisc noop state DOWN group default qlen 1000
    link/ether 16:6e:09:ca:1e:de brd ff:ff:ff:ff:ff:ff
    inet 10.20.30.1/32 scope global mydocker0
       valid_lft forever preferred_lft forever
```

### enable the new bridge mydocker0

```bash
$ ip link set mydocker0 up
$ ip addr show mydocker0
6: mydocker0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UNKNOWN group default qlen 1000
    link/ether 16:6e:09:ca:1e:de brd ff:ff:ff:ff:ff:ff
    inet 10.20.30.1/32 scope global mydocker0
       valid_lft forever preferred_lft forever
    inet6 fe80::146e:9ff:feca:1ede/64 scope link
       valid_lft forever preferred_lft forever
```

### set route for the subnet of bridge mydocker0

```bash
$ ip route add 10.20.30.0/24 dev mydocker0 proto kernel scope link src 10.20.30.1
$ route -n | grep mydocker0
10.20.30.0      0.0.0.0         255.255.255.0   U     0      0        0 mydocker0
```

### set iptables rules for the bridge mydocker0

- set MASQUERADE SNAT rule for packets from the bridge `mydocker0`.

```bash
$ iptables -t nat -A POSTROUTING -s 10.20.30.0/24 ! -o mydocker0 -j MASQUERADE
```

- isolate different networks on the same host by setting the mark `0x12151991` to all the packages from the bridge `mydocker0`.

```bash
$ mark="0x$(echo -n mydocker0 | sha256sum | head -c 8)"
$ echo ${mark}
0x12151991
$ iptables -t mangle -A PREROUTING -i mydocker0 -j MARK --set-mark ${mark}
```

- accept all the packets from the bridge `mydocker0` to the internet.

```bash
$ iptables -t mangle -A POSTROUTING -o ens3 -m mark --mark ${mark} -j ACCEPT
```

> notes: `ens3` is the physical network interface on this host. you have to set multiple similar rules if the host have multiple physical network interfaces. use the following command to get all physical network interfaces:

```bash
for netdev in /sys/class/net/*; do
    [[ $(readlink ${netdev}) =~ virtual ]] && continue
    basename ${netdev}
done
````

- drop all the packets from the bridge `mydocker0` to the other bridges on the same host.

```bash
$ iptables -t mangle -A POSTROUTING ! -o mydocker0 -m mark --mark ${mark} -j DROP
```

## Connect Network to Containers

### create two netns to mock two containers

```bash
$ ip netns add ns1
$ ip netns add ns2
```

### create two pairs of veth devices

```bash
$ ip link add veth1 type veth peer name ceth1
$ ip link add veth2 type veth peer name ceth2
```

### add veth1 and veth2 to the bridge mydocker0

```bash
$ ip link set veth1 master mydocker0
$ ip link set veth2 master mydocker0
$ brctl show mydocker0
bridge name   bridge id            STP enabled    interfaces
mydocker0     8000.229f0c568a85    no             veth1
                                                  veth2
```

### add ceth1 and ceth2 to the two different netns

```bash
$ ip link set ceth1 netns ns1
$ ip link set ceth2 netns ns2
```

### enable veth1 and veth2 on the bridge

```bash
$ ip link set veth1 up
$ ip link set veth2 up
```

### enable ceth1 and ceth2 in two netns

```bash
$ ip netns exec ns1 ip link set ceth1 up
$ ip netns exec ns2 ip link set ceth2 up
```

### set ip addr for ceth1 and ceth2 in two netns

```bash
$ ip netns exec ns1 ip addr add 10.20.30.10/24 dev ceth1
$ ip netns exec ns2 ip addr add 10.20.30.20/24 dev ceth2
```

### set default route in two netns

```bash
$ ip netns exec ns1 ip route add default dev ceth1 via 10.20.30.1
$ ip netns exec ns2 ip route add default dev ceth2 via 10.20.30.1
```

### the network topology of containers

```bash
+--------------------------------------------------------+--------------------------------+--------------------------------+
|                                                        |                                |                                |
|                          Host                          |           Container 1          |           Container 2          |
|                                                        |                                |                                |
|  +--------------------------------------------------+  |  +--------------------------+  |  +--------------------------+  |
|  |              Network Protocol Stack              |  |  |  Network Protocol Stack  |  |  |  Network Protocol Stack  |  |
|  +--------------------------------------------------+  |  +--------------------------+  |  +--------------------------+  |
|          ↑                  ↑                          |                ↑               |                ↑               |
|..........|.............(0x12151991)....................|................|...............|................|...............|
|          ↓                  ↓                          |                ↓               |                ↓               |
|  +----------------+  +-------------+                   |         +-------------+        |         +-------------+        |
|  | 192.168.31.153 |  |  10.20.30.1 |                   |         | 10.20.30.10 |        |         | 10.20.30.20 |        |
|  +----------------+  +-------------+      +-------+    |         +-------------+        |         +-------------+        |
|  |      ens3      |  |  mydocker0  | <--> | veth1 | <--|-------->|    ceth1    |        |         |    ceth2    |        |
|  +----------------+  +-------------+      +-------+    |         +-------------+        |         +-------------+        |
|          ↑                  ↑                          |                                |                ↑               |
|          |                  |                          |                                |                |               |
|          |                  ↓                          |                                |                |               |
|          |              +-------+                      |                                |                |               |
|          |              | veth2 |<---------------------|--------------------------------|----------------+               |
|          |              +-------+                      |                                |                                |
|          |                                             |                                |                                |
+----------|---------------------------------------------+--------------------------------+--------------------------------+
           ↓
   Physical Network Interface (192.168.31.153/24)
```

### check ip addr and route in two netns

```bash
$ ip netns exec ns1 ip addr show ceth1
7: ceth1@if8: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default qlen 1000
    link/ether 16:8a:a3:7c:a9:3f brd ff:ff:ff:ff:ff:ff link-netnsid 0
    inet 10.20.30.10/24 scope global ceth1
       valid_lft forever preferred_lft forever
    inet6 fe80::148a:a3ff:fe7c:a93f/64 scope link
       valid_lft forever preferred_lft forever
$ ip netns exec ns1 ip route
default via 10.20.30.1 dev ceth1
10.20.30.0/24 dev ceth1  proto kernel  scope link  src 10.20.30.10
```

```bash
$ ip netns exec ns2 ip addr show ceth2
9: ceth2@if10: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default qlen 1000
    link/ether 3e:c1:8f:c3:17:1b brd ff:ff:ff:ff:ff:ff link-netnsid 0
    inet 10.20.30.20/24 scope global ceth2
       valid_lft forever preferred_lft forever
    inet6 fe80::3cc1:8fff:fec3:171b/64 scope link
       valid_lft forever preferred_lft forever
$ ip netns exec ns2 ip route
default via 10.20.30.1 dev ceth2
10.20.30.0/24 dev ceth2  proto kernel  scope link  src 10.20.30.20
```

### connect each other between two netns

```bash
$ ip netns exec ns1 ping -c 4 10.20.30.20
PING 10.20.30.20 (10.20.30.20) 56(84) bytes of data.
64 bytes from 10.20.30.20: icmp_seq=1 ttl=64 time=0.277 ms
64 bytes from 10.20.30.20: icmp_seq=2 ttl=64 time=0.116 ms
64 bytes from 10.20.30.20: icmp_seq=3 ttl=64 time=0.114 ms
64 bytes from 10.20.30.20: icmp_seq=4 ttl=64 time=0.103 ms

--- 10.20.30.20 ping statistics ---
4 packets transmitted, 4 received, 0% packet loss, time 3053ms
rtt min/avg/max/mdev = 0.103/0.152/0.277/0.073 ms

$ ip netns exec ns2 ping -c 4 10.20.30.10
PING 10.20.30.10 (10.20.30.10) 56(84) bytes of data.
64 bytes from 10.20.30.10: icmp_seq=1 ttl=64 time=0.093 ms
64 bytes from 10.20.30.10: icmp_seq=2 ttl=64 time=0.112 ms
64 bytes from 10.20.30.10: icmp_seq=3 ttl=64 time=0.135 ms
64 bytes from 10.20.30.10: icmp_seq=4 ttl=64 time=0.144 ms

--- 10.20.30.10 ping statistics ---
4 packets transmitted, 4 received, 0% packet loss, time 3057ms
rtt min/avg/max/mdev = 0.093/0.121/0.144/0.019 ms
```

### connect internet in two netns

```bash
$ ip netns exec ns1 ping -c 4 8.8.8.8
PING 8.8.8.8 (8.8.8.8) 56(84) bytes of data.
64 bytes from 8.8.8.8: icmp_seq=1 ttl=40 time=58.0 ms
64 bytes from 8.8.8.8: icmp_seq=2 ttl=40 time=90.9 ms
64 bytes from 8.8.8.8: icmp_seq=3 ttl=40 time=78.6 ms
64 bytes from 8.8.8.8: icmp_seq=4 ttl=40 time=79.2 ms

--- 8.8.8.8 ping statistics ---
4 packets transmitted, 4 received, 0% packet loss, time 3004ms
$ ip netns exec ns2 ping -c 4 8.8.8.8
PING 8.8.8.8 (8.8.8.8) 56(84) bytes of data.
64 bytes from 8.8.8.8: icmp_seq=1 ttl=40 time=66.8 ms
64 bytes from 8.8.8.8: icmp_seq=2 ttl=40 time=82.9 ms
64 bytes from 8.8.8.8: icmp_seq=3 ttl=40 time=77.1 ms
64 bytes from 8.8.8.8: icmp_seq=4 ttl=40 time=68.3 ms

--- 8.8.8.8 ping statistics ---
4 packets transmitted, 4 received, 0% packet loss, time 3000ms
rtt min/avg/max/mdev = 66.888/73.824/82.921/6.575 ms
```

### can't connect veth bind on another bridge (e.g. docker0)

```bash
$ ip netns exec ns1 ping -c 4 172.17.0.2
PING 172.17.0.2 (172.17.0.2) 56(84) bytes of data.

--- 172.17.0.2 ping statistics ---
4 packets transmitted, 0 received, 100% packet loss, time 3057ms

$ ip netns exec ns2 ping -c 4 172.17.0.2
PING 172.17.0.2 (172.17.0.2) 56(84) bytes of data.

--- 172.17.0.2 ping statistics ---
4 packets transmitted, 0 received, 100% packet loss, time 3066ms

```

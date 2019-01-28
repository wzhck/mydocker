# The Mydocker Project

![Mydocker Project Logo](doc/picture/logo.png "The Mydocker Project")

Mydocker (get inspired from [mydocker](https://github.com/xianlubird/mydocker) project) is a docker-like container runtime written in Go.

This is just an experimental cli tool, which implements lowlevel techniques such as `cgroups`, `namespace`, `unionFS` (`aufs`, `overlay2`), `virtual networks` (`veth`, `bridge`), `iptables`, and etc. Notes that, implement a daemon service is not an emergency requirements of mydocker project right now.

## Environment Requirements

- [Ubuntu](https://www.ubuntu.com/download/server) 16.04/18.04 LTS
- [Docker-CE](https://docs.docker.com/v17.09/engine/installation/linux/docker-ce/ubuntu/) installed

## Download Mydocker From GitHub

```bash
$ mkdir -p $GOPATH/src/github.com/weikeit
$ cd $GOPATH/src/github.com/weikeit
$ git clone git@github.com:weikeit/mydocker.git
$ go build -o /usr/local/bin/mydocker mydocker
```

## Overview of Mydocker Command

```bash
$ mydocker -h
NAME:
   mydocker - mydocker is a simple container runtime implementation.
The purpose of this project is to learn how docker works and how to
write a docker-like container runtime by ourselves, enjoy it!

USAGE:
   mydocker [global options] command [command options] [arguments...]

VERSION:
   0.0.0

COMMANDS:
     run       Create a new mydocker container
     ps        List all containers on the host
     logs      Show all the logs of a container
     exec      Run a command in a running container
     stop      Stop one or more containers
     start     Start one or more containers
     restart   Restart one or more containers
     rm        Remove one or more containers
     rmn       Remove one or more networks
     rmi       Remove one or more images
     pull      Pull an image from a registry
     inspect   Print information of mydocker objects
     networks  List networks on the host
     images    List images on the host
     network   Manage container networks
     image     Manage container images
     help, h   Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --debug        print mydocker debug logs
   --help, -h     show help
   --version, -v  print the version
```

## Manage Mydocker Images

### pull a new image from a registry

```bash
$ mydocker pull mysql:5.7.25
$ mydocker image pull mysql:5.7.25
5.7.25: Pulling from library/mysql
......
Status: Downloaded newer image for mysql:5.7.25
```

### list images on this host

```bash
$ mydocker images
$ mydocker image list
IMAGE ID       REPO    TAG      COUNTS   CREATED               SIZE
141eda20897f   mysql   5.7.25   0        2019-01-25 09:43:22   354.9 MB
```

### remove one or more images

```bash
# can't remove images whose counts > 0
$ mydocker rmi mysql:5.7.25
$ mydocker image rm 141eda20897f
```

## Manage Mydocker Networks

### create a new network

```bash
$ mydocker network create -s 10.10.1.0/24 subnet1
```

### list networks on this host

```bash
$ mydocker networks
$ mydocker network list
NAME        IPNETS          GATEWAY      COUNTS   DRIVER   CREATED
mydocker0   10.20.30.0/24   10.20.30.1   0        bridge   2019-01-25 09:42:13
subnet1     10.10.1.0/24    10.10.1.1    0        bridge   2019-01-25 09:44:38
```

### remove one or more networks

```bash
# can't remove networks whose counts > 0
$ mydocker rmn subnet1
$ mydocker network rm subnet1
```

## Manage Mydocker Containers

### run a new container

```bash
$ mydocker run -h
NAME:
   mydocker run - Create a new mydocker container

USAGE:
   mydocker run [command options] [arguments...]

OPTIONS:
   --detach, -d                      Run the container in background
   --name value, -n value            Assign a name to the container
   --hostname value                  Set hostname in the container
   --dns value                       Set DNS servers in the container (default: "8.8.8.8", "8.8.4.4")
   --image value, -i value           The image to be used (name or id)
   --env value, -e value             Set environment variables, e.g. -e key=value
   --volume value, -v value          Bind a local directory/file, e.g. -v /src:/dst
   --network value, --net value      Connect the container to a network (none to disable)
   --publish value, -p value         Publish the container's port(s) to the host
   --storage-driver value, -s value  Storage driver to be used (default: "overlay2")
   --cpu-cfs-period value            Limit CPU CFS (Completely Fair Scheduler) period in us (default: 200000)
   --cpu-cfs-quota value             Limit CPU CFS (Completely Fair Scheduler) quota in us (default: 200000)
   --cpu-rt-period value             Limit CPU Real-Time Scheduler period in us (default: 1000000)
   --cpu-rt-runtime value            Limit CPU Real-Time Scheduler runtime in us (default: 950000)
   --cpu-shares value, -c value      CPU shares (relative weight) (default: 1024)
   --cpuset-cpus value               CPUs in which to allow execution (0-3, 0,1) (default: "0-7")
   --cpuset-mems value               MEMs in which to allow execution (0-3, 0,1) (default: "0-0")
   --memory-limit value              Memory limit in bytes; -1 indicates unlimited (default: -1)
   --memory-soft-limit value         Memory soft limit in bytes; -1 indicates unlimited (default: -1)
   --memory-swap-limit value         Swap limit equals to memory plus swap; -1 indicates unlimited (default: -1)
   --memory-swappiness value         Tune container memory swappiness (range [0, 100]) (default: 60)
   --oom-kill-disable                Disable oom killer, i.e., process will be hung if oom, NOT killed
   --kernel-memory-limit value       Kernel memory limit in bytes; -1 indicates unlimited (default: -1)
   --kernel-memory-tcp-limit value   Kernel memory tcp limit in bytes; -1 indicates unlimited (default: -1)
   --pids-max value                  Limit pids number in container; 0 indicates unlimited (default: 0)
```

e.g., run a mysql container using official mysql image:

```bash
$ mydocker run -d \
           -p 8036:3306 \
           -i mysql:5.7.25 \
           -e MYSQL_ROOT_PASSWORD=r00tme \
           -e MYSQL_DATABASE=testdb \
           -e MYSQL_USER=testuser \
           -e MYSQL_PASSWORD=r00test \
           -v /root/mysql:/var/lib/mysql \
           --name mysql-test \
           --hostname mysql-test \
           --cpu-cfs-period 200000 \
           --cpu-cfs-quota 500000 \
           --cpu-shares 2048 \
           --cpuset-cpus 1-2 \
           --cpuset-mems 0 \
           --memory-limit 512000000 \
           --memory-soft-limit 1024000000 \
           --pids-max 100
```

### list containers on this host

```bash
$ mydocker ps
CONTAINER ID   NAME         IMAGE          STATUS    DRIVER     PID     COMMAND                         IPS          PORTS        CREATED
4f2322145e66   mysql-test   mysql:5.7.25   running   overlay2   30942   [docker-entrypoint.sh mysqld]   10.20.30.2   8036->3306   2019-01-25 09:46:04
```

### show logs of a container

```bash
$ mydocker logs mysql-test
Initializing database
......
too long messages
......
```

tips: add the `-f` option to follow the logs' output:

```bash
$ mydocker logs -f mysql-test
2019-01-25T01:46:22.324979Z 0 [Warning] 'user' entry 'mysql.session@localhost' ignored in --skip-name-resolve mode.
2019-01-25T01:46:22.325012Z 0 [Warning] 'user' entry 'mysql.sys@localhost' ignored in --skip-name-resolve mode.
2019-01-25T01:46:22.325077Z 0 [Warning] 'db' entry 'performance_schema mysql.session@localhost' ignored in --skip-name-resolve mode.
2019-01-25T01:46:22.325094Z 0 [Warning] 'db' entry 'sys mysql.sys@localhost' ignored in --skip-name-resolve mode.
2019-01-25T01:46:22.325129Z 0 [Warning] 'proxies_priv' entry '@ root@localhost' ignored in --skip-name-resolve mode.
2019-01-25T01:46:22.331095Z 0 [Warning] 'tables_priv' entry 'user mysql.session@localhost' ignored in --skip-name-resolve mode.
2019-01-25T01:46:22.331138Z 0 [Warning] 'tables_priv' entry 'sys_config mysql.sys@localhost' ignored in --skip-name-resolve mode.
2019-01-25T01:46:22.350515Z 0 [Note] Event Scheduler: Loaded 0 events
2019-01-25T01:46:22.350895Z 0 [Note] mysqld: ready for connections.
Version: '5.7.25'  socket: '/var/run/mysqld/mysqld.sock'  port: 3306  MySQL Community Server (GPL)
[waiting for new log messages]......
```

### execute a command in the container

```bash
$ mydocker exec mysql-test -- mysql -uroot -pr00tme -e 'show databases;'
mysql: [Warning] Using a password on the command line interface can be insecure.
+--------------------+
| Database           |
+--------------------+
| information_schema |
| mysql              |
| performance_schema |
| sys                |
| testdb             |
+--------------------+
$ mydocker exec mysql-test bash
root@mysql-test:/# mysql -utestuser -pr00test -e 'show databases;'
mysql: [Warning] Using a password on the command line interface can be insecure.
+--------------------+
| Database           |
+--------------------+
| information_schema |
| testdb             |
+--------------------+
```

### access mysql via `hostIP:8036` on other host

```bash
# container mysql-test is running on 192.168.31.153
# run mysql command (192.168.31.146) to access it
$ mysql -h 192.168.31.153 -P 8036 -utestuser -pr00test -e 'show databases;'
mysql: [Warning] Using a password on the command line interface can be insecure.
+--------------------+
| Database           |
+--------------------+
| information_schema |
| testdb             |
+--------------------+
```

### (dis)connect a netowrk to a container

create a new virtual network interface connected to subnet1 to mysql-test container:

```bash
$ mydocker network connect subnet1 mysql-test
$ CONTAINER ID   NAME         IMAGE          STATUS    DRIVER     PID     COMMAND                         IPS                     PORTS        CREATED
  4f2322145e66   mysql-test   mysql:5.7.25   running   overlay2   30942   [docker-entrypoint.sh mysqld]   10.20.30.2, 10.10.1.2   8036->3306   2019-01-25 09:46:04
$ mydocker exec mysql-test bash
# the official mysql image doesn't install iproute2 package.
root@mysql-test:/# apt-get update && apt-get install iproute2 -y
......
root@mysql-test:/# ip addr
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
    inet6 ::1/128 scope host
       valid_lft forever preferred_lft forever
22: ceth-f0573c58@if23: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default qlen 1000
    link/ether 4e:70:8b:1c:2b:6d brd ff:ff:ff:ff:ff:ff link-netnsid 0
    inet 10.20.30.2/24 brd 10.20.30.255 scope global ceth-f0573c58
       valid_lft forever preferred_lft forever
    inet6 fe80::4c70:8bff:fe1c:2b6d/64 scope link
       valid_lft forever preferred_lft forever
24: ceth-67674376@if25: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default qlen 1000
    link/ether 8e:c5:e9:96:c1:34 brd ff:ff:ff:ff:ff:ff link-netnsid 0
    inet 10.10.1.2/24 brd 10.10.1.255 scope global ceth-67674376
       valid_lft forever preferred_lft forever
    inet6 fe80::8cc5:e9ff:fe96:c134/64 scope link
       valid_lft forever preferred_lft forever
```

remove the virtual network interface connected to subnet1 from mysql-test container:

```bash
$ mydocker network disconnect subnet1 mysql-test
$ mydocker exec mysql-test ip addr
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
    inet6 ::1/128 scope host
       valid_lft forever preferred_lft forever
26: ceth-f0573c58@if23: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default qlen 1000
    link/ether aa:d1:f5:4b:19:12 brd ff:ff:ff:ff:ff:ff link-netnsid 0
    inet 10.20.30.2/24 brd 10.20.30.255 scope global ceth-f0573c58
       valid_lft forever preferred_lft forever
    inet6 fe80::a8d1:f5ff:fe4b:1912/64 scope link
       valid_lft forever preferred_lft forever
```

### stop/start/restart/remove one or more containers

```bash
$ mydocker stop mysql-test
4f2322145e66
$ mydocker start mysql-test
4f2322145e66
$ mydocker restart mysql-test
4f2322145e66
4f2322145e66
$ mydocker rm mysql-test
4f2322145e66
```

## use `--debug` option of mydocker to show debug logs

```bash
$ mydocker --debug exec mysql-test -- mysql -uroot -pr00tme -e 'show databases;'
[2019-01-25 09:55:03] DEBUG initing networks only once before each mydocker command
[2019-01-25 09:55:03] DEBUG found a network mydocker0 of driver bridge
[2019-01-25 09:55:03] DEBUG initing network mydocker0 of driver bridge
[2019-01-25 09:55:03] DEBUG create the network mydocker0 of bridge driver with iprange 10.20.30.0/24
[2019-01-25 09:55:03] DEBUG the ip addr '10.20.30.1/24' on the interface 'mydocker0' already exists
[2019-01-25 09:55:03] DEBUG set the interface mydocker0 up
[2019-01-25 09:55:03] DEBUG found a network subnet1 of driver bridge
[2019-01-25 09:55:03] DEBUG initing network subnet1 of driver bridge
[2019-01-25 09:55:03] DEBUG create the network subnet1 of bridge driver with iprange 10.10.1.0/24
[2019-01-25 09:55:03] DEBUG the ip addr '10.10.1.1/24' on the interface 'subnet1' already exists
[2019-01-25 09:55:03] DEBUG set the interface subnet1 up
[2019-01-25 09:55:03] DEBUG found existed networks:
{
    "mydocker0": {
        "IPNet": "10.20.30.0/24",
        "Gateway": "10.20.30.1",
        "Name": "mydocker0",
        "Counts": 1,
        "Driver": "bridge",
        "CreateTime": "2019-01-25 09:42:13"
    },
    "subnet1": {
        "IPNet": "10.10.1.0/24",
        "Gateway": "10.10.1.1",
        "Name": "subnet1",
        "Counts": 0,
        "Driver": "bridge",
        "CreateTime": "2019-01-25 09:44:38"
    }
}
[2019-01-25 09:55:03] DEBUG will execute command <mysql -uroot -pr00tme -e 'show databases;'> in the container (pid: 30942)
got the env container_pid: 30942
got the env container_cmd: mysql -uroot -pr00tme -e 'show databases;'
got the env cgroup_root: /sys/fs/cgroup
got the env cgroup_path: /mydocker/4f2322145e66
add the process 32176 to /sys/fs/cgroup/cpu/mydocker/4f2322145e66/cgroup.procs
add the process 32176 to /sys/fs/cgroup/cpuset/mydocker/4f2322145e66/cgroup.procs
add the process 32176 to /sys/fs/cgroup/memory/mydocker/4f2322145e66/cgroup.procs
add the process 32176 to /sys/fs/cgroup/blkio/mydocker/4f2322145e66/cgroup.procs
add the process 32176 to /sys/fs/cgroup/devices/mydocker/4f2322145e66/cgroup.procs
add the process 32176 to /sys/fs/cgroup/pids/mydocker/4f2322145e66/cgroup.procs
add the process 32176 to /sys/fs/cgroup/net_cls/mydocker/4f2322145e66/cgroup.procs
add the process 32176 to /sys/fs/cgroup/net_prio/mydocker/4f2322145e66/cgroup.procs
add the process 32176 to /sys/fs/cgroup/freezer/mydocker/4f2322145e66/cgroup.procs
add the process 32176 to /sys/fs/cgroup/hugetlb/mydocker/4f2322145e66/cgroup.procs
set the process 32176 to namespace ipc
set the process 32176 to namespace uts
set the process 32176 to namespace net
set the process 32176 to namespace pid
set the process 32176 to namespace mnt
mysql: [Warning] Using a password on the command line interface can be insecure.
+--------------------+
| Database           |
+--------------------+
| information_schema |
| mysql              |
| performance_schema |
| sys                |
| testdb             |
+--------------------+
```

## use `inspect` subcommand to show lowlevel info of mydocker object

```bash
$ mydocker inspect mysql:5.7.25 mysql-test
Showing mysql:5.7.25 as a image:
{
    "Uuid": "141eda20897f",
    "Size": "354.9 MB",
    "Counts": 1,
    "RepoTag": "mysql:5.7.25",
    "WorkingDir": "",
    "CreateTime": "2019-01-25 09:43:22",
    "Entrypoint": [
        "docker-entrypoint.sh"
    ],
    "Command": [
        "mysqld"
    ],
    "Envs": [
        "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
        "GOSU_VERSION=1.7",
        "MYSQL_MAJOR=5.7",
        "MYSQL_VERSION=5.7.25-1debian9"
    ]
}
Showing mysql-test as a container:
{
    "Detach": true,
    "Uuid": "4f2322145e66",
    "Name": "mysql-test",
    "Hostname": "mysql-test",
    "Dns": [
        "8.8.8.8",
        "8.8.4.4"
    ],
    "Image": "mysql:5.7.25",
    "CreateTime": "2019-01-25 09:46:04",
    "Status": "running",
    "StorageDriver": "overlay2",
    "Rootfs": {
        "ContainerDir": "/var/lib/mydocker/containers/4f2322145e66",
        "ImageDir": "/var/lib/mydocker/images/141eda20897f",
        "WriteDir": "/var/lib/mydocker/containers/4f2322145e66/diff",
        "MergeDir": "/var/lib/mydocker/containers/4f2322145e66/merged"
    },
    "Commands": [
        "docker-entrypoint.sh",
        "mysqld"
    ],
    "Cgroups": {
        "Pid": 30942,
        "Path": "/mydocker/4f2322145e66",
        "Resources": {
            "CpuCfsPeriod": 200000,
            "CpuCfsQuota": 500000,
            "CpuRtPeriod": 1000000,
            "CpuRtRuntime": 950000,
            "CpuShares": 2048,
            "CpusetCpus": "1-2",
            "CpusetMems": "0",
            "MemoryLimit": 512000000,
            "MemorySoftLimit": 1024000000,
            "MemorySwapLimit": -1,
            "MemorySwappiness": 60,
            "OomKillDisable": false,
            "KernelMemoryLimit": -1,
            "KernelMemoryTCPLimit": -1,
            "BlkioWeight": 0,
            "BlkioLeafWeight": 0,
            "BlkioWeightDevice": null,
            "BlkioLeafWeightDevice": null,
            "BlkioThrottleReadBpsDevice": null,
            "BlkioThrottleWriteBpsDevice": null,
            "BlkioThrottleReadIOPSDevice": null,
            "BlkioThrottleWriteIOPSDevice": null,
            "Device": null,
            "PidsMax": 100,
            "NetClsClassid": 0,
            "NetPrioIfpriomap": null,
            "Freezer": "",
            "HugepagesLimit": null
        }
    },
    "Volumes": {
        "/root/mysql": "/var/lib/mydocker/containers/4f2322145e66/merged/var/lib/mysql"
    },
    "Envs": {
        "GOSU_VERSION": "1.7",
        "MYSQL_DATABASE": "testdb",
        "MYSQL_MAJOR": "5.7",
        "MYSQL_PASSWORD": "r00test",
        "MYSQL_ROOT_PASSWORD": "r00tme",
        "MYSQL_USER": "testuser",
        "MYSQL_VERSION": "5.7.25-1debian9",
        "PATH": "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
    },
    "Ports": {
        "8036": "3306"
    },
    "Endpoints": [
        {
            "IPAddr": "10.20.30.2",
            "Device": "veth-f0573c58@ceth-f0573c58",
            "Network": {
                "driver": "bridge",
                "name": "mydocker0"
            },
            "Uuid": "0f616391b7bb",
            "Ports": {
                "8036": "3306"
            }
        }
    ]
}
```

## License

See the [LICENSE](https://github.com/weikeit/mydocker/blob/master/LICENSE.md) file for license rights and limitations (MIT).

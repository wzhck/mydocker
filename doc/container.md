## Implement Mydocker Container

This document will explain what have been done when creating a new container.

### prepare rootfs for the container

1. use `docker` command to pull an official image

```bash
$ docker pull alpine:3.8
```

2. make rootfs from a `docker` container

```bash
$ docker run -d --name alpine alpine:3.8
$ docker export -o /tmp/alpine.tar alpine
$ mkdir -p /tmp/container/{readonly,diff,work,merged}
$ tar -xvf /tmp/alpine.tar -C /tmp/container/readonly
```

3. use `overlay2` to mount a rootfs of the container

```bash
$ options="lowerdir=/tmp/container/readonly,upperdir=/tmp/container/diff,workdir=/tmp/container/work"
$ mount -t overlay -o ${options} overlay /tmp/container/merged
$ ll /tmp/container/merged
total 64K
drwxr-xr-x  2 root root 4.0K Jan 30 10:55 bin/
drwxr-xr-x  4 root root 4.0K Feb  1 09:49 dev/
drwxr-xr-x 15 root root 4.0K Feb  1 09:49 etc/
drwxr-xr-x  2 root root 4.0K Jan 30 10:55 home/
drwxr-xr-x  5 root root 4.0K Jan 30 10:55 lib/
drwxr-xr-x  5 root root 4.0K Jan 30 10:55 media/
drwxr-xr-x  2 root root 4.0K Jan 30 10:55 mnt/
dr-xr-xr-x  2 root root 4.0K Jan 30 10:55 proc/
drwx------  2 root root 4.0K Jan 30 10:55 root/
drwxr-xr-x  2 root root 4.0K Jan 30 10:55 run/
drwxr-xr-x  2 root root 4.0K Jan 30 10:55 sbin/
drwxr-xr-x  2 root root 4.0K Jan 30 10:55 srv/
drwxr-xr-x  2 root root 4.0K Jan 30 10:55 sys/
drwxrwxrwt  2 root root 4.0K Jan 30 10:55 tmp/
drwxr-xr-x  7 root root 4.0K Jan 30 10:55 usr/
drwxr-xr-x 11 root root 4.0K Jan 30 10:55 var/
```

### use `unshare` to run a new `bash` command in new namespaces

1. check the origin namespaces of mnt, uts, ipc, net, pid first:

```bash
$ ls -l /proc/$$/ns
total 0
lrwxrwxrwx 1 root root 0 Feb  1 10:37 cgroup -> cgroup:[4026531835]
lrwxrwxrwx 1 root root 0 Feb  1 10:32 ipc -> ipc:[4026531839]
lrwxrwxrwx 1 root root 0 Feb  1 10:37 mnt -> mnt:[4026531840]
lrwxrwxrwx 1 root root 0 Feb  1 10:37 net -> net:[4026531993]
lrwxrwxrwx 1 root root 0 Feb  1 10:37 pid -> pid:[4026531836]
lrwxrwxrwx 1 root root 0 Feb  1 10:37 pid_for_children -> pid:[4026531836]
lrwxrwxrwx 1 root root 0 Feb  1 10:37 user -> user:[4026531837]
lrwxrwxrwx 1 root root 0 Feb  1 10:32 uts -> uts:[4026531838]
```

2. create a `bash` shell in new namespaces of mnt, uts, ipc, net, pid using `unshare` command:

```bash
$ cd /tmp/container/merged
$ unshare -m -u -i -n -p -f --mount-proc bash
$ ls -l /proc/$$/ns
total 0
lrwxrwxrwx 1 root root 0 Feb  1 13:53 cgroup -> cgroup:[4026531835]
lrwxrwxrwx 1 root root 0 Feb  1 13:53 ipc -> ipc:[4026532253]
lrwxrwxrwx 1 root root 0 Feb  1 13:53 mnt -> mnt:[4026532251]
lrwxrwxrwx 1 root root 0 Feb  1 13:53 net -> net:[4026532256]
lrwxrwxrwx 1 root root 0 Feb  1 13:53 pid -> pid:[4026532254]
lrwxrwxrwx 1 root root 0 Feb  1 13:53 pid_for_children -> pid:[4026532254]
lrwxrwxrwx 1 root root 0 Feb  1 13:53 user -> user:[4026531837]
lrwxrwxrwx 1 root root 0 Feb  1 13:53 uts -> uts:[4026532252]
$ ps -ef
UID        PID  PPID  C STIME TTY          TIME CMD
root         1     0  0 12:29 pts/0    00:00:00 bash
root        12     1  0 12:29 pts/0    00:00:00 ps -ef
$ ip addr
1: lo: <LOOPBACK> mtu 65536 qdisc noop state DOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
```

as we can see, the namespace files of mnt, uts, ipc, net, pid are all different from the origin.

### use `pivot_root` syscall to isolate rootfs of the container

```bash
$ mount --make-rprivate /
$ mount /tmp/container/merged -o bind /tmp/container/merged
$ mkdir -p /tmp/container/merged/.oldroot
$ pivot_root /tmp/container/merged /tmp/container/merged/.oldroot
$ chroot / /bin/sh
$ umount -l /.oldroot
$ rm -rf /.oldroot
```

### mount some necessary virtual filesystems

```bash
$ mount -t proc proc /proc
$ mount -t sysfs -o ro sysfs /sys
$ mount -t tmpfs -o mode=0755,size=100M tmpfs /dev
$ mkdir -p /dev/pts
$ mount -t devpts -o mode=0620,newinstance,ptmxmode=0666,gid=5 devpts /dev/pts
$ mkdir -p /dev/shm
$ mount -t tmpfs -o mode=1777,size=100M shm /dev/shm
$ mkdir -p /dev/mqueue
$ mount -t mqueue -o mode=1777,size=100M mqueue /dev/mqueue
$ df -hT
Filesystem           Type            Size      Used Available Use% Mounted on
overlay              overlay        90.0G     54.0G     31.4G  63% /
tmpfs                tmpfs         100.0M         0    100.0M   0% /dev
shm                  tmpfs         100.0M         0    100.0M   0% /dev/shm
```

### create some necessary symlinks

> can't mock this operation here!

```bash
$ ln -sf /dev/fd /proc/self/fd
$ ln -sf /dev/stdin /proc/self/fd/0
$ ln -sf /dev/stdout /proc/self/fd/1
$ ln -sf /dev/stderr /proc/self/fd/2
```

### create some necessary character special devices

> the following character special devices are necessary for many applications.

```bash
$ mknod /dev/null c 1 3
$ mknod /dev/zero c 1 5
$ mknod /dev/full c 1 7
$ mknod /dev/random c 1 8
$ mknod /dev/urandom c 1 9
$ mknod /dev/tty c 5 0
$ mknod /dev/console c 5 1
$ find /dev -type c -exec ls -l {} +
crw-r--r--    1 root     root        5,   1 Feb  1 04:36 /dev/console
crw-r--r--    1 root     root        1,   7 Feb  1 04:36 /dev/full
crw-r--r--    1 root     root        1,   3 Feb  1 04:36 /dev/null
crw-rw-rw-    1 root     root        5,   2 Feb  1 04:35 /dev/pts/ptmx
crw-r--r--    1 root     root        1,   8 Feb  1 04:36 /dev/random
crw-r--r--    1 root     root        5,   0 Feb  1 04:36 /dev/tty
crw-r--r--    1 root     root        1,   9 Feb  1 04:36 /dev/urandom
crw-r--r--    1 root     root        1,   5 Feb  1 04:36 /dev/zero
```

### set hostname and dns server in the container

```bash
$ hostname container-demo
$ echo 'nameserver 8.8.8.8' > /etc/resolv.conf
```

### create network for the container

let's connect container to the default network `mydocker0` (ref: [networks](https://github.com/weikeit/mydocker/blob/master/doc/networks.md))

1. create a pair of veth peers on the host

```bash
$ ip link add veth-demo type veth peer name ceth-demo
$ ip link set veth-demo master mydocker0
$ mkdir -p /var/run/netns
$ ppid=$(ps -ef |grep -oP "\d+(?=\s+$(pgrep unshare))")
$ pid=$(ps -ef |grep -oP "\d+(?=\s+${ppid})")
$ ln -sfT /proc/${pid}/ns/net /var/run/netns/container-netns
$ ip netns list
container-netns
$ ip link set ceth-demo netns container-netns
$ ip link set veth-demo up
```

2. config networks in the container

```bash
$ ip link set lo up
$ ip link set ceth-demo up
$ ip addr add 10.20.30.40/24 dev ceth-demo
$ ip route add default dev ceth-demo via 10.20.30.1
$ ip addr
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
    inet6 ::1/128 scope host
       valid_lft forever preferred_lft forever
7: ceth-demo@if8: <BROADCAST,MULTICAST,UP,LOWER_UP,M-DOWN> mtu 1500 qdisc noqueue state UP qlen 1000
    link/ether 06:3a:0e:cf:0e:c2 brd ff:ff:ff:ff:ff:ff
    inet 10.20.30.40/24 scope global ceth-demo
       valid_lft forever preferred_lft forever
    inet6 fe80::43a:eff:fecf:ec2/64 scope link
       valid_lft forever preferred_lft forever
$ ip route
default via 10.20.30.1 dev ceth-demo
10.20.30.0/24 dev ceth-demo scope link  src 10.20.30.40
```

3. test network connections in the container

```bash
# connect to the bridge mydocker0
$ ping -c 4 10.20.30.1
PING 10.20.30.1 (10.20.30.1): 56 data bytes
64 bytes from 10.20.30.1: seq=0 ttl=64 time=0.140 ms
64 bytes from 10.20.30.1: seq=1 ttl=64 time=0.144 ms
64 bytes from 10.20.30.1: seq=2 ttl=64 time=0.177 ms
64 bytes from 10.20.30.1: seq=3 ttl=64 time=0.145 ms

--- 10.20.30.1 ping statistics ---
4 packets transmitted, 4 packets received, 0% packet loss
round-trip min/avg/max = 0.140/0.151/0.177 ms
# connect to the internet
$ ping -c 4 baidu.com
PING baidu.com (220.181.57.216): 56 data bytes
64 bytes from 220.181.57.216: seq=0 ttl=50 time=30.625 ms
64 bytes from 220.181.57.216: seq=1 ttl=50 time=34.665 ms
64 bytes from 220.181.57.216: seq=2 ttl=50 time=27.205 ms
64 bytes from 220.181.57.216: seq=3 ttl=50 time=36.944 ms

--- baidu.com ping statistics ---
4 packets transmitted, 4 packets received, 0% packet loss
round-trip min/avg/max = 27.205/32.359/36.944 ms
```

### use `cgroup` to limit resources

add the pid of init process in the container to a new cgroup of all subsystems.

```bash
$ mkdir -p /sys/fs/cgroup/{cpu,pids,memory}/container-demo
$ echo ${pid} > /sys/fs/cgroup/cpu/container-demo/cgroup.procs
$ echo ${pid} > /sys/fs/cgroup/pids/container-demo/cgroup.procs
$ echo ${pid} > /sys/fs/cgroup/memory/container-demo/cgroup.procs
```

1. test cpu limits in the container

```bash
$ echo 200000 > /sys/fs/cgroup/cpu/container-demo/cpu.cfs_period_us
$ echo 100000 > /sys/fs/cgroup/cpu/container-demo/cpu.cfs_quota_us
```

> Notes: the container can only use at most 50% usages of 1 cpu.

2. test memory limits in the container

```bash
$ echo 102400000 > /sys/fs/cgroup/memory/container-demo/memory.limit_in_bytes
```

> Notes: the container can only use at most 100MB memory.

3. test pids limits in the container

```bash
$ echo 50 > /sys/fs/cgroup/pids/container-demo/pids.max
```

> Notes: the container can only fork at most 50 processes.

Now, let's test `pids` limit using classic fork bomb:

```bash
# install bash in the container
$ apk update
$ apk add bash
$ bash
$ :(){ :|:& }; :
......
bash: fork: retry: Resource temporarily unavailable
bash: fork: retry: Resource temporarily unavailable
bash: fork: retry: Resource temporarily unavailable
......
$ (on the host) cat /sys/fs/cgroup/pids/container-demo/pids.current
50
```

### call `execve()` syscall to run user's command in the container

```bash
$ exec <user's commands>
```

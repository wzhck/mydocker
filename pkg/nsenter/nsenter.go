package nsenter

/*
#cgo CFLAGS: -Wall
#define _GNU_SOURCE
#include <fcntl.h>
#include <sched.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

__attribute__((constructor)) void enter_namespace(void) {
	char *debug = getenv("debug_nsenter");
	char *container_pid = getenv("container_pid");
	char *container_cmd = getenv("container_cmd");
	char *cgroup_root = getenv("cgroup_root");
	char *cgroup_path = getenv("cgroup_path");

	if (!container_pid || !container_cmd || !cgroup_root || !cgroup_path) {
		return;
	}

	if (strcmp(debug, "true") == 0) {
		printf("got the env container_pid: %s\n", container_pid);
		printf("got the env container_cmd: %s\n", container_cmd);
		printf("got the env cgroup_root: %s\n", cgroup_root);
		printf("got the env cgroup_path: %s\n", cgroup_path);
	}

	char child_pid[12];
	sprintf(child_pid, "%d", getpid());

	char *subsystems[] = {"cpu", "cpuset", "memory", "blkio", "devices",
			"pids", "net_cls", "net_prio", "freezer", "hugetlb"};

	int i;
	char procsfile[1024];
	// note: need to add process to cgroup.procs before calling setns().
	for (i = 0; i < sizeof(subsystems) / sizeof(*subsystems); i++) {
		// note: cgroup_path contains the leading slash, e.g., /mydocker/11e7b2361e2c
		sprintf(procsfile, "%s/%s%s/cgroup.procs", cgroup_root, subsystems[i], cgroup_path);
		if (strcmp(debug, "true") == 0) {
			printf("add the process %s to %s\n", child_pid, procsfile);
		}

		int fd = open(procsfile, O_WRONLY);
		if (write(fd, child_pid, strlen(child_pid)) <= 0) {
			printf("failed to add the process %s to %s\n", child_pid, procsfile);
		}

		close(fd);
	}

	char nsfile[1024];
	char *namespaces[] = {"ipc", "uts", "net", "pid", "mnt"};

	int j;
	for (j = 0; j < sizeof(namespaces) / sizeof(*namespaces); j++) {
		sprintf(nsfile, "/proc/%s/ns/%s", container_pid, namespaces[j]);
		if (strcmp(debug, "true") == 0) {
			printf("set the process %s to namespace %s\n", child_pid, namespaces[j]);
		}

		int fd = open(nsfile, O_RDONLY);
		if (setns(fd, 0) == -1) {
			printf("failed to set the process %s to namespace %s\n", child_pid, namespaces[j]);
		}

		close(fd);
	}

	exit(system(container_cmd));
}
*/
import "C"

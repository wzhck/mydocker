package nsenter

/*
#include <errno.h>
#include <sched.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>

__attribute__((constructor)) void enter_namespace(void) {
	char *container_pid;
	container_pid = getenv("ContainerPid");
	if (!container_pid) {
		//fprintf(stdout, "missing container_pid!\n");
		return;
	}
	//fprintf(stdout, "got container_pid: %s\n", container_pid);

	char *container_cmd;
	container_cmd = getenv("ContainerCmd");
	if (!container_cmd) {
		//fprintf(stdout, "missing container_cmd!\n");
		return;
	}
	//fprintf(stdout, "got container_cmd: <%s>\n", container_cmd);

	int i;
	char nspath[1024];
	char *namespaces[] = {"ipc", "uts", "net", "pid", "mnt"};

	for (i=0; i<5; i++) {
		sprintf(nspath, "/proc/%s/ns/%s", container_pid, namespaces[i]);
		int fd = open(nspath, O_RDONLY);

		if (setns(fd, 0) == -1) {
			fprintf(stderr, "failed to setns on %s namespace: %s\n",
					namespaces[i], strerror(errno));
		}
		close(fd);
	}
	exit(system(container_cmd));
}
*/
import "C"

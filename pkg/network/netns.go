package network

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netns"
	"os"
	"runtime"
)

// call this function using the following way:
// defer EnterContainerNetns(pid)()
func EnterContainerNetns(pid int) func() {
	netnsFileName := fmt.Sprintf("/proc/%d/ns/net", pid)
	netnsFile, err := os.OpenFile(netnsFileName, os.O_RDONLY, 0)
	if err != nil {
		return func() {}
	}

	netnsFd := netnsFile.Fd()

	// need to lock current process's OS thread first
	// or goroutine maybe schedule current process to
	// another OS thread, and now, we can ensure that
	// current process is always in container netns.
	runtime.LockOSThread()

	// keep current netns first, so that we can return back
	// to the same netns after exiting from container netns
	originNetns, err := netns.Get()
	if err != nil {
		log.Errorf("failed to get current netns: %v", err)
		return func() {
			runtime.UnlockOSThread()
			netnsFile.Close()
		}
	}

	////////////////////////////////////////////////////////////
	// move current process into the netns of container (pid) //
	////////////////////////////////////////////////////////////

	if err := netns.Set(netns.NsHandle(netnsFd)); err != nil {
		log.Errorf("failed to set netns: %v", err)
		return func() {
			originNetns.Close()
			runtime.UnlockOSThread()
			netnsFile.Close()
		}
	}

	return func() {
		netns.Set(originNetns)
		originNetns.Close()
		runtime.UnlockOSThread()
		netnsFile.Close()
	}
}

package network

import (
	"fmt"
	"github.com/weikeit/mydocker/util"
	"os/exec"
)

func setPortMap(outPort, inIP, inPort string) error {
	var cmd *exec.Cmd
	outIPs, err := util.GetHostIPs()
	if err != nil {
		return err
	}

	// for the host with ips: 192.168.138.179 10.20.30.1
	// and a portMap rule: 0.0.0.0/8000 -> 10.20.30.2:80
	// we need to set the following iptables rules:

	// step1: set normal DNAT forwarding rules
	// -t nat -A PREROUTING ! -s 127.0.0.1 ! -d 127.0.0.1 -p tcp -m tcp --dport 8000 \
	//        -j DNAT --to-destination 10.20.30.2:80
	// Note: we need to exclude the packet whose src or dst is 127.0.0.1,
	// for which we will set special iptables rules in step2.
	cmd = GetDnatIPTablesCmd("-C", outPort, inIP, inPort)
	if err := cmd.Run(); err != nil {
		cmd = GetDnatIPTablesCmd("-A", outPort, inIP, inPort)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("faild to set iptables rule `%s`: %v", cmd.Args, err)
		}
	}

	// ref: https://serverfault.com/a/646534/447507
	// step2: set special iptables rules to deal with 127.0.0.1
	// this enable to access 10.20.30.2:80 via 127.0.0.1:8000 or localhost:8000

	// -t nat -A OUTPUT -d 127.0.0.1 -p tcp -m tcp --dport 8000 \
	//        -j DNAT --to-destination 10.20.30.2:80
	// this rule just changes the packet's dst from 127.0.0.1 to 10.20.30.2
	// and packet can get routed outside the source device, but destination
	// will not know how to correctly return it!
	cmd = GetHostIPTablesCmd("-C", "127.0.0.1", outPort, inIP, inPort)
	if err := cmd.Run(); err != nil {
		cmd = GetHostIPTablesCmd("-A", "127.0.0.1", outPort, inIP, inPort)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("faild to set iptables rule `%s`: %v", cmd.Args, err)
		}
	}

	// -t nat -A POSTROUTING -d 10.20.30.2 -p tcp -m tcp --dport 80 \
	//        -j SNAT --to-source 192.168.138.179
	// here, we also need to add additional SNAT rule to let destination
	// know how to correctly return it, we choose the first host ipaddr.
	cmd = GetSnatIPTablesCmd("-C", outIPs[0], inIP, inPort)
	if err := cmd.Run(); err != nil {
		cmd = GetSnatIPTablesCmd("-A", outIPs[0], inIP, inPort)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("faild to set iptables rule `%s`: %v", cmd.Args, err)
		}
	}

	// step3: set special iptables rules to deal with other localhost ipaddrs
	// this enable to access 10.20.30.2:80 via 192.168.138.179:8000 or 10.20.30.1:8000
	for _, outIP := range outIPs {
		// -t nat -A OUTPUT -d 192.168.138.179 -p tcp -m tcp --dport 8000 \
		//        -j DNAT --to-destination 10.20.30.2:80
		cmd = GetHostIPTablesCmd("-C", outIP, outPort, inIP, inPort)
		if err := cmd.Run(); err != nil {
			cmd = GetHostIPTablesCmd("-A", outIP, outPort, inIP, inPort)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("faild to set iptables rule `%s`: %v", cmd.Args, err)
			}
		}
	}

	return nil
}

func delPortMap(outPort, inIP, inPort string) error {
	var cmd *exec.Cmd
	outIPs, err := util.GetHostIPs()
	if err != nil {
		return err
	}

	// -t nat -D PREROUTING ! -s 127.0.0.1 ! -d 127.0.0.1 -p tcp -m tcp --dport 8000 \
	//        -j DNAT --to-destination 10.20.30.2:80
	cmd = GetDnatIPTablesCmd("-C", outPort, inIP, inPort)
	if err := cmd.Run(); err == nil {
		cmd = GetDnatIPTablesCmd("-D", outPort, inIP, inPort)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("faild to del iptables rule `%s`: %v", cmd.Args, err)
		}
	}

	// -t nat -D OUTPUT -d 127.0.0.1 -p tcp -m tcp --dport 8000 \
	//        -j DNAT --to-destination 10.20.30.2:80
	cmd = GetHostIPTablesCmd("-C", "127.0.0.1", outPort, inIP, inPort)
	if err := cmd.Run(); err == nil {
		cmd = GetHostIPTablesCmd("-D", "127.0.0.1", outPort, inIP, inPort)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("faild to del iptables rule `%s`: %v", cmd.Args, err)
		}
	}

	// -t nat -D POSTROUTING -d 10.20.30.2 -p tcp -m tcp --dport 80 \
	//        -j SNAT --to-source 192.168.138.179
	cmd = GetSnatIPTablesCmd("-C", outIPs[0], inIP, inPort)
	if err := cmd.Run(); err == nil {
		cmd = GetSnatIPTablesCmd("-D", outIPs[0], inIP, inPort)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("faild to del iptables rule `%s`: %v", cmd.Args, err)
		}
	}

	for _, outIP := range outIPs {
		// -t nat -D OUTPUT -d 192.168.138.179 -p tcp -m tcp --dport 8000 \
		//        -j DNAT --to-destination 10.20.30.2:80
		cmd = GetHostIPTablesCmd("-C", outIP, outPort, inIP, inPort)
		if err := cmd.Run(); err == nil {
			cmd = GetHostIPTablesCmd("-D", outIP, outPort, inIP, inPort)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("faild to del iptables rule `%s`: %v", cmd.Args, err)
			}
		}
	}

	return nil
}

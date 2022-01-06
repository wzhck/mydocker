package network

import (
	"fmt"
	"net"
	"os/exec"
	"strings"

	"weike.sh/mydocker/util"
)

func setBridgeIptablesRules(bridgeName string, subnet *net.IPNet) error {
	var cmd *exec.Cmd

	// for the host with physcial ip: 192.168.138.179
	// the bridge mydocker0 with subnet 10.20.30.0/24
	// the mark of bridge mydocker0 is: 0x12151991

	// step1: set MASQUERADE SNAT rule for packets from this bridge
	// -t nat -A POSTROUTING -s 10.20.30.0/24 ! -o mydocker0 -j MASQUERADE
	cmd = genMasqIPTablesCmd("-C", subnet.String(), bridgeName)
	if err := cmd.Run(); err != nil {
		cmd = genMasqIPTablesCmd("-A", subnet.String(), bridgeName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("faild to set iptables rule %s: %v", cmd.Args, err)
		}
	}

	// generate the mark (uint32) using the first 8 hexadecimal chars of the
	// bridgeName's sha256 checksum, and treat it as an uint32 value (hex).
	mark := "0x" + util.Sha256Sum(bridgeName)[:8]

	// step2: mark all packets from the bridge mydocker0 with value 0x12151991
	// -t mangle -A PREROUTING -i mydocker0 -j MARK --set-mark 0x12151991
	cmd = genMarkIPTablesCmd("-C", bridgeName, mark)
	if err := cmd.Run(); err != nil {
		cmd = genMarkIPTablesCmd("-A", bridgeName, mark)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("faild to set iptables rule %s: %v", cmd.Args, err)
		}
	}

	physNics, err := GetPhysicalNics()
	if err != nil {
		return fmt.Errorf("failed to get physical nics: %v", err)
	}

	// step3: accept all the packets from the bridge mydocker0 to the Internet.
	// -t mangle -A POSTROUTING -o eth0 -m mark --mark 0x12151991 -j ACCEPT
	// -t mangle -A POSTROUTING -o eth1 -m mark --mark 0x12151991 -j ACCEPT
	for _, physNic := range physNics {
		cmd = genHpysIPTablesCmd("-C", physNic, mark)
		if err := cmd.Run(); err != nil {
			cmd = genHpysIPTablesCmd("-A", physNic, mark)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("faild to set iptables rule %s: %v", cmd.Args, err)
			}
		}
	}

	// step4: drop all all the packets from the bridge mydocker0 to other bridges.
	// -t mangle -A POSTROUTING ! -o mydocker0 -m mark --mark 0x12151991 -j DROP
	cmd = genDropIPTablesCmd("-C", bridgeName, mark)
	if err := cmd.Run(); err != nil {
		cmd = genDropIPTablesCmd("-A", bridgeName, mark)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("faild to set iptables rule %s: %v", cmd.Args, err)
		}
	}

	return nil
}

func delBridgeIptablesRules(bridgeName string, subnet *net.IPNet) error {
	var cmd *exec.Cmd

	// -t nat -D POSTROUTING -s 10.20.30.0/24 ! -o mydocker0 -j MASQUERADE
	cmd = genMasqIPTablesCmd("-C", subnet.String(), bridgeName)
	if err := cmd.Run(); err == nil {
		cmd = genMasqIPTablesCmd("-D", subnet.String(), bridgeName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("faild to del iptables rule %s: %v", cmd.Args, err)
		}
	}

	// generate the mark (uint32) using the first 8 hexadecimal chars of the
	// bridgeName's sha256 checksum, and treat it as an uint32 value (hex).
	mark := "0x" + util.Sha256Sum(bridgeName)[:8]

	// -t mangle -D PREROUTING -i mydocker0 -j MARK --set-mark 0x12151991
	cmd = genMarkIPTablesCmd("-C", bridgeName, mark)
	if err := cmd.Run(); err == nil {
		cmd = genMarkIPTablesCmd("-D", bridgeName, mark)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("faild to del iptables rule %s: %v", cmd.Args, err)
		}
	}

	physNics, err := GetPhysicalNics()
	if err != nil {
		return fmt.Errorf("failed to get physical nics: %v", err)
	}

	// -t mangle -D POSTROUTING -o eth0 -m mark --mark 0x12151991 -j ACCEPT
	// -t mangle -D POSTROUTING -o eth1 -m mark --mark 0x12151991 -j ACCEPT
	for _, physNic := range physNics {
		cmd = genHpysIPTablesCmd("-C", physNic, mark)
		if err := cmd.Run(); err == nil {
			cmd = genHpysIPTablesCmd("-D", physNic, mark)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("faild to del iptables rule %s: %v", cmd.Args, err)
			}
		}
	}

	// -t mangle -D POSTROUTING ! -o mydocker0 -m mark --mark 0x12151991 -j DROP
	cmd = genDropIPTablesCmd("-C", bridgeName, mark)
	if err := cmd.Run(); err == nil {
		cmd = genDropIPTablesCmd("-D", bridgeName, mark)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("faild to del iptables rule %s: %v", cmd.Args, err)
		}
	}

	return nil
}

func setPortMap(outPort, inIP, inPort string) error {
	var cmd *exec.Cmd
	outIPs, err := GetPhysicalIPs()
	if err != nil {
		return err
	}

	// for the host with physical ips: 192.168.138.179
	// and a portMap rule: 0.0.0.0/8000 -> 10.20.30.2:80
	// we need to set the following iptables rules:

	// step1: set normal DNAT forwarding rules
	// -t nat -A PREROUTING ! -s 127.0.0.1 ! -d 127.0.0.1 -p tcp -m tcp --dport 8000 \
	//        -j DNAT --to-destination 10.20.30.2:80
	// Note: we need to exclude the packet whose src or dst is 127.0.0.1,
	// for which we will set special iptables rules in step2.
	cmd = genDnatIPTablesCmd("-C", outPort, inIP, inPort)
	if err := cmd.Run(); err != nil {
		cmd = genDnatIPTablesCmd("-A", outPort, inIP, inPort)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("faild to set iptables rule %s: %v", cmd.Args, err)
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
	cmd = genHostIPTablesCmd("-C", "127.0.0.1", outPort, inIP, inPort)
	if err := cmd.Run(); err != nil {
		cmd = genHostIPTablesCmd("-A", "127.0.0.1", outPort, inIP, inPort)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("faild to set iptables rule %s: %v", cmd.Args, err)
		}
	}

	// -t nat -A POSTROUTING -s 127.0.0.1 -d 10.20.30.2 -p tcp -m tcp --dport 80 \
	//        -j SNAT --to-source 192.168.138.179
	// here, we also need to add additional SNAT rule to let destination
	// know how to correctly return it, we choose the first host ipaddr.
	cmd = genSnatIPTablesCmd("-C", outIPs[0], inIP, inPort)
	if err := cmd.Run(); err != nil {
		cmd = genSnatIPTablesCmd("-A", outIPs[0], inIP, inPort)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("faild to set iptables rule %s: %v", cmd.Args, err)
		}
	}

	// step3: set special iptables rules to deal with physical ipaddrs
	// this enable to access 10.20.30.2:80 via 192.168.138.179:8000
	for _, outIP := range outIPs {
		// -t nat -A OUTPUT -d 192.168.138.179 -p tcp -m tcp --dport 8000 \
		//        -j DNAT --to-destination 10.20.30.2:80
		cmd = genHostIPTablesCmd("-C", outIP, outPort, inIP, inPort)
		if err := cmd.Run(); err != nil {
			cmd = genHostIPTablesCmd("-A", outIP, outPort, inIP, inPort)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("faild to set iptables rule %s: %v", cmd.Args, err)
			}
		}
	}

	return nil
}

func delPortMap(outPort, inIP, inPort string) error {
	var cmd *exec.Cmd
	outIPs, err := GetPhysicalIPs()
	if err != nil {
		return err
	}

	// -t nat -D PREROUTING ! -s 127.0.0.1 ! -d 127.0.0.1 -p tcp -m tcp --dport 8000 \
	//        -j DNAT --to-destination 10.20.30.2:80
	cmd = genDnatIPTablesCmd("-C", outPort, inIP, inPort)
	if err := cmd.Run(); err == nil {
		cmd = genDnatIPTablesCmd("-D", outPort, inIP, inPort)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("faild to del iptables rule %s: %v", cmd.Args, err)
		}
	}

	// -t nat -D OUTPUT -d 127.0.0.1 -p tcp -m tcp --dport 8000 \
	//        -j DNAT --to-destination 10.20.30.2:80
	cmd = genHostIPTablesCmd("-C", "127.0.0.1", outPort, inIP, inPort)
	if err := cmd.Run(); err == nil {
		cmd = genHostIPTablesCmd("-D", "127.0.0.1", outPort, inIP, inPort)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("faild to del iptables rule %s: %v", cmd.Args, err)
		}
	}

	// -t nat -D POSTROUTING -s 127.0.0.1 -d 10.20.30.2 -p tcp -m tcp --dport 80 \
	//        -j SNAT --to-source 192.168.138.179
	cmd = genSnatIPTablesCmd("-C", outIPs[0], inIP, inPort)
	if err := cmd.Run(); err == nil {
		cmd = genSnatIPTablesCmd("-D", outIPs[0], inIP, inPort)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("faild to del iptables rule %s: %v", cmd.Args, err)
		}
	}

	for _, outIP := range outIPs {
		// -t nat -D OUTPUT -d 192.168.138.179 -p tcp -m tcp --dport 8000 \
		//        -j DNAT --to-destination 10.20.30.2:80
		cmd = genHostIPTablesCmd("-C", outIP, outPort, inIP, inPort)
		if err := cmd.Run(); err == nil {
			cmd = genHostIPTablesCmd("-D", outIP, outPort, inIP, inPort)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("faild to del iptables rule %s: %v", cmd.Args, err)
			}
		}
	}

	return nil
}

/////////////////////////////////////////////////////////////////////
// some util functions to generate specific iptables rules command //
/////////////////////////////////////////////////////////////////////

func genMasqIPTablesCmd(action, subnet, bridge string) *exec.Cmd {
	argsReplacer := strings.NewReplacer(
		"{action}", action,
		"{subnet}", subnet,
		"{bridge}", bridge)
	args := argsReplacer.Replace(bridgeIPTRules["masq"])
	return exec.Command("iptables", strings.Split(args, " ")...)
}

func genMarkIPTablesCmd(action, bridge, mark string) *exec.Cmd {
	argsReplacer := strings.NewReplacer(
		"{action}", action,
		"{bridge}", bridge,
		"{mark}", mark)
	args := argsReplacer.Replace(bridgeIPTRules["mark"])
	return exec.Command("iptables", strings.Split(args, " ")...)
}

func genHpysIPTablesCmd(action, physnic, mark string) *exec.Cmd {
	argsReplacer := strings.NewReplacer(
		"{action}", action,
		"{physnic}", physnic,
		"{mark}", mark)
	args := argsReplacer.Replace(bridgeIPTRules["phys"])
	return exec.Command("iptables", strings.Split(args, " ")...)
}

func genDropIPTablesCmd(action, bridge, mark string) *exec.Cmd {
	argsReplacer := strings.NewReplacer(
		"{action}", action,
		"{bridge}", bridge,
		"{mark}", mark)
	args := argsReplacer.Replace(bridgeIPTRules["drop"])
	return exec.Command("iptables", strings.Split(args, " ")...)
}

func genDnatIPTablesCmd(action, outPort, inIP, inPort string) *exec.Cmd {
	argsReplacer := strings.NewReplacer(
		"{action}", action,
		"{outPort}", outPort,
		"{inIP}", inIP,
		"{inPort}", inPort)
	args := argsReplacer.Replace(portMapsIPTRules["dnat"])
	return exec.Command("iptables", strings.Split(args, " ")...)
}

func genHostIPTablesCmd(action, outIP, outPort, inIP, inPort string) *exec.Cmd {
	argsReplacer := strings.NewReplacer(
		"{action}", action,
		"{outIP}", outIP,
		"{outPort}", outPort,
		"{inIP}", inIP,
		"{inPort}", inPort)
	args := argsReplacer.Replace(portMapsIPTRules["host"])
	return exec.Command("iptables", strings.Split(args, " ")...)
}

func genSnatIPTablesCmd(action, outIP, inIP, inPort string) *exec.Cmd {
	argsReplacer := strings.NewReplacer(
		"{action}", action,
		"{outIP}", outIP,
		"{inIP}", inIP,
		"{inPort}", inPort)
	args := argsReplacer.Replace(portMapsIPTRules["snat"])
	return exec.Command("iptables", strings.Split(args, " ")...)
}

package clab

import (
	"fmt"
	"os/exec"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func (c *cLab) InitVirtualWiring() {
	log.Debug("delete dummyA eth link")
	var cmd *exec.Cmd
	var err error
	// list intefaces
	cmd = exec.Command("sudo", "ip", "link", "show")
	// TODO
	cmd = exec.Command("sudo", "ip", "link", "del", "dummyA")
	err = runCmd(cmd)
	if err != nil {
		log.Debugf("%s failed with: %v", cmd.String(), err)
	}
	cmd = exec.Command("sudo", "ip", "link", "del", "dummyB")
	err = runCmd(cmd)
	if err != nil {
		log.Debugf("%s failed with: %v", cmd.String(), err)
	}
}

// CreateVirtualWiring provides the virtual topology between the containers
func (c *cLab) CreateVirtualWiring(id int, link *Link) (err error) {
	log.Infof("Create virtual wire : %s, %s, %s, %s", link.A.Node.LongName, link.B.Node.LongName, link.A.EndpointName, link.B.EndpointName)

	var cmd *exec.Cmd
	if link.A.Node.Kind != "bridge" && link.B.Node.Kind != "bridge" {
		err = c.createAToBveth(link)
		if err != nil {
			return err
		}
	}
	//
	if link.A.Node.Kind != "bridge" && link.B.Node.Kind != "bridge" { // none of the 2 nodes is a bridge
		log.Debug("create dummy veth pair")
		cmd = exec.Command("sudo", "ip", "link", "add", "dummyA", "type", "veth", "peer", "name", "dummyB")
		err = runCmd(cmd)
		if err != nil {
			log.Fatalf("%s failed with: %v", cmd.String(), err)
		}
	} else if link.A.Node.Kind != "bridge" { // node of link A is a bridge
		log.Debug("create dummy veth pair")
		cmd = exec.Command("sudo", "ip", "link", "add", "dummyA", "type", "veth", "peer", "name", link.B.EndpointName)
		err = runCmd(cmd)
		if err != nil {
			log.Fatalf("%s failed with: %v", cmd.String(), err)
		}
	} else { // node of link B is a bridge
		log.Debug("create dummy veth pair")
		cmd = exec.Command("sudo", "ip", "link", "add", "dummyB", "type", "veth", "peer", "name", link.A.EndpointName)
		err = runCmd(cmd)
		if err != nil {
			log.Fatalf("%s failed with: %v", cmd.String(), err)
		}
	}

	if link.A.Node.Kind != "bridge" { // if node A of link is not a bridge
		log.Debug("map dummy interface on container A to NS")
		cmd = exec.Command("sudo", "ip", "link", "set", "dummyA", "netns", link.A.Node.LongName)
		err = runCmd(cmd)
		if err != nil {
			log.Fatalf("%s failed with: %v", cmd.String(), err)
		}
	}

	if link.B.Node.Kind != "bridge" { // if node B of link is not a bridge
		log.Debug("map dummy interface on container B to NS")
		cmd = exec.Command("sudo", "ip", "link", "set", "dummyB", "netns", link.B.Node.LongName)
		err = runCmd(cmd)
		if err != nil {
			log.Fatalf("%s failed with: %v", cmd.String(), err)
		}
	}

	if link.A.Node.Kind != "bridge" { // if node A of link is not a bridge
		log.Debug("rename interface container NS A")
		cmd = exec.Command("sudo", "ip", "netns", "exec", link.A.Node.LongName, "ip", "link", "set", "dummyA", "name", link.A.EndpointName)
		err = runCmd(cmd)
		if err != nil {
			log.Fatalf("%s failed with: %v", cmd.String(), err)
		}
	} else {
		log.Debug("map veth pair to bridge")
		cmd = exec.Command("sudo", "ip", "link", "set", link.A.EndpointName, "master", link.A.Node.ShortName)
		err = runCmd(cmd)
		if err != nil {
			log.Fatalf("%s failed with: %v", cmd.String(), err)
		}
	}

	if link.B.Node.Kind != "bridge" { // if node B of link is not a bridge
		log.Debug("rename interface container NS B")
		cmd = exec.Command("sudo", "ip", "netns", "exec", link.B.Node.LongName, "ip", "link", "set", "dummyB", "name", link.B.EndpointName)
		err = runCmd(cmd)
		if err != nil {
			log.Fatalf("%s failed with: %v", cmd.String(), err)
		}
	} else {
		log.Debug("map veth pair to bridge")
		cmd = exec.Command("sudo", "ip", "link", "set", link.B.EndpointName, "master", link.B.Node.ShortName)
		err = runCmd(cmd)
		if err != nil {
			log.Fatalf("%s failed with: %v", cmd.String(), err)
		}
	}

	if link.A.Node.Kind != "bridge" { // if node A of link is not a bridge
		log.Debug("set interface up in container NS A")
		cmd = exec.Command("sudo", "ip", "netns", "exec", link.A.Node.LongName, "ip", "link", "set", link.A.EndpointName, "up")
		err = runCmd(cmd)
		if err != nil {
			log.Fatalf("%s failed with: %v", cmd.String(), err)
		}
	} else {
		log.Debug("set interface up in bridge")
		cmd = exec.Command("sudo", "ip", "link", "set", link.A.EndpointName, "up")
		err = runCmd(cmd)
		if err != nil {
			log.Fatalf("%s failed with: %v", cmd.String(), err)
		}
	}

	if link.B.Node.Kind != "bridge" { // if node B of link is not a bridge
		log.Debug("set interface up in container NS B")
		cmd = exec.Command("sudo", "ip", "netns", "exec", link.B.Node.LongName, "ip", "link", "set", link.B.EndpointName, "up")
		err = runCmd(cmd)
		if err != nil {
			log.Fatalf("%s failed with: %v", cmd.String(), err)
		}
	} else {
		log.Debug("set interface up in bridge")
		cmd = exec.Command("sudo", "ip", "link", "set", link.B.EndpointName, "up")
		err = runCmd(cmd)
		if err != nil {
			log.Fatalf("%s failed with: %v", cmd.String(), err)
		}
	}

	if link.A.Node.Kind != "bridge" { // if node A of link is not a bridge
		log.Debug("set RX, TX offload off on container A")
		cmd = exec.Command("docker", "exec", link.A.Node.LongName, "ethtool", "--offload", link.A.EndpointName, "rx", "off", "tx", "off")
		err = runCmd(cmd)
		if err != nil {
			log.Debugf("%s failed with: %v", cmd.String(), err)
		}
	} else {
		log.Debug("set RX, TX offload off on veth of the bridge interface")
		cmd = exec.Command("sudo", "ethtool", "--offload", link.A.EndpointName, "rx", "off", "tx", "off")
		err = runCmd(cmd)
		if err != nil {
			log.Debugf("%s failed with: %v", cmd.String(), err)
		}
	}

	if link.B.Node.Kind != "bridge" { // if node B of link is not a bridge
		log.Debug("set RX, TX offload off on container B")
		cmd = exec.Command("docker", "exec", link.B.Node.LongName, "ethtool", "--offload", link.B.EndpointName, "rx", "off", "tx", "off")
		err = runCmd(cmd)
		if err != nil {
			log.Debugf("%s failed with: %v", cmd.String(), err)
		}
	} else {
		log.Debug("set RX, TX offload off on veth of the bridge interface")
		cmd = exec.Command("sudo", "ethtool", "--offload", link.B.EndpointName, "rx", "off", "tx", "off")
		err = runCmd(cmd)
		if err != nil {
			log.Debugf("%s failed with: %v", cmd.String(), err)
		}
	}

	//sudo ip link add tmp_a type veth peer name tmp_b
	//sudo ip link set tmp_a netns $srl_a
	//ip link set tmp_b netns $srl_b
	//ip netns exec $srl_a ip link set tmp_a name $srl_a_int
	//ip netns exec $srl_b ip link set tmp_b name $srl_b_int
	//ip netns exec $srl_a ip link set $srl_a_int up
	//ip netns exec $srl_b ip link set $srl_b_int up
	//docker exec -ti $srl_a ethtool --offload $srl_a_int rx off tx off
	//docker exec -ti $srl_b ethtool --offload $srl_b_int rx off tx off

	//sudo ip link add <bridge-name> type bridge
	//sudo ip link set <bridge-name> up

	//sudo ip link add tmp_a type veth peer name <vethint>
	//sudo ip link set tmp_a netns <container>
	//sudo ip netns exec <container> ip link set tmp_a name e1-10
	//sudo ip netns exec <container> ip link set e1-10 up
	//sudo ip link set <vethint> master <bridge-name>
	//sudo ip link set <vethint> up
	//docker exec -ti <container> --offload $srl_a_int rx off tx off
	//sudo ethtool --offload <vethint> rx off tx off

	return nil

}

func (c *cLab) createAToBveth(l *Link) error {
	log.Debug("create dummy veth pair")
	interfaceA := fmt.Sprintf("dA-%s", genIfName())
	interfaceB := fmt.Sprintf("dB-%s", genIfName())

	var cmd *exec.Cmd
	var err error
	cmd = exec.Command("sudo", "ip", "link", "add", interfaceA, "type", "veth", "peer", "name", interfaceB)
	err = runCmd(cmd)
	if err != nil {
		return err
	}
	log.Debug("map dummy interface on container A to NS")
	cmd = exec.Command("sudo", "ip", "link", "set", interfaceA, "netns", l.A.Node.LongName)
	err = runCmd(cmd)
	if err != nil {
		return err
	}
	log.Debug("map dummy interface on container B to NS")
	cmd = exec.Command("sudo", "ip", "link", "set", interfaceB, "netns", l.B.Node.LongName)
	err = runCmd(cmd)
	if err != nil {
		return err
	}
	log.Debug("rename interface container NS A")
	cmd = exec.Command("sudo", "ip", "netns", "exec", l.A.Node.LongName, "ip", "link", "set", interfaceA, "name", l.A.EndpointName)
	err = runCmd(cmd)
	if err != nil {
		return err
	}
	log.Debug("rename interface container NS B")
	cmd = exec.Command("sudo", "ip", "netns", "exec", l.B.Node.LongName, "ip", "link", "set", interfaceB, "name", l.B.EndpointName)
	err = runCmd(cmd)
	if err != nil {
		return err
	}
	log.Debug("set interface up in container NS A")
	cmd = exec.Command("sudo", "ip", "netns", "exec", l.A.Node.LongName, "ip", "link", "set", l.A.EndpointName, "up")
	err = runCmd(cmd)
	if err != nil {
		return err
	}
	log.Debug("set interface up in container NS B")
	cmd = exec.Command("sudo", "ip", "netns", "exec", l.B.Node.LongName, "ip", "link", "set", l.B.EndpointName, "up")
	err = runCmd(cmd)
	if err != nil {
		return err
	}
	log.Debug("set RX, TX offload off on container A")
	cmd = exec.Command("docker", "exec", l.A.Node.LongName, "ethtool", "--offload", l.A.EndpointName, "rx", "off", "tx", "off")
	err = runCmd(cmd)
	if err != nil {
		return err
	}
	log.Debug("set RX, TX offload off on container B")
	cmd = exec.Command("docker", "exec", l.B.Node.LongName, "ethtool", "--offload", l.B.EndpointName, "rx", "off", "tx", "off")
	err = runCmd(cmd)
	if err != nil {
		return err
	}
	return nil
}
func (c *cLab) createvethToBridge(l *Link) error {
	var cmd *exec.Cmd
	var err error
	dummyIface := fmt.Sprintf("dA-%s", genIfName())
	// assume A is a bridge
	bridgeIfname := l.A.EndpointName
	containerNS := l.B.Node.LongName
	containerIfName := l.B.EndpointName

	if l.A.Node.Kind != "bridge" { // change var values if A is not a bridge
		bridgeIfname = l.B.EndpointName
		containerIfName = l.A.EndpointName
		containerNS = l.B.Node.LongName
	}
	log.Debug("create dummy veth pair")
	cmd = exec.Command("sudo", "ip", "link", "add", dummyIface, "type", "veth", "peer", "name", bridgeIfname)
	err = runCmd(cmd)
	if err != nil {
		return err
	}
	log.Debug("map dummy interface on container A to NS")
	cmd = exec.Command("sudo", "ip", "link", "set", dummyIface, "netns", containerNS)
	err = runCmd(cmd)
	if err != nil {
		return err
	}
	log.Debug("rename interface container NS A")
	cmd = exec.Command("sudo", "ip", "netns", "exec", containerNS, "ip", "link", "set", dummyIface, "name", containerIfName)
	err = runCmd(cmd)
	if err != nil {
		return err
	}
	log.Debug("set interface up in container NS A")
	cmd = exec.Command("sudo", "ip", "netns", "exec", containerNS, "ip", "link", "set", containerIfName, "up")
	err = runCmd(cmd)
	if err != nil {
		return err
	}
	log.Debug("set RX, TX offload off on container A")
	cmd = exec.Command("docker", "exec", containerNS, "ethtool", "--offload", containerIfName, "rx", "off", "tx", "off")
	err = runCmd(cmd)
	if err != nil {
		return err
	}
	return nil
}

// DeleteVirtualWiring deletes the virtual wiring
func (c *cLab) DeleteVirtualWiring(id int, link *Link) (err error) {
	log.Info("Delete virtual wire :", link.A.Node.ShortName, link.B.Node.ShortName, link.A.EndpointName, link.B.EndpointName)

	var cmd *exec.Cmd

	if link.A.Node.Kind != "bridge" {
		log.Debug("Delete netns: ", link.A.Node.LongName)
		cmd = exec.Command("sudo", "ip", "netns", "del", link.A.Node.LongName)
		err = runCmd(cmd)
		if err != nil {
			log.Debugf("%s failed with: %v", cmd.String(), err)
		}
	}

	if link.B.Node.Kind != "bridge" {
		log.Debug("Delete netns: ", link.B.Node.LongName)
		cmd = exec.Command("sudo", "ip", "netns", "del", link.B.Node.LongName)
		err = runCmd(cmd)
		if err != nil {
			log.Debugf("%s failed with: %v", cmd.String(), err)
		}
	}

	return nil
}

func runCmd(cmd *exec.Cmd) error {
	b, err := cmd.CombinedOutput()
	if err != nil {
		log.Debugf("'%s' failed with: %v", cmd.String(), err)
		log.Debugf("'%s' failed output: %v", cmd.String(), string(b))
		return err
	}
	log.Debugf("'%s' output: %v", cmd.String(), string(b))
	return nil
}

func genIfName() string {
	s, _ := uuid.New().MarshalText() // .MarshalText() always return a nil error
	return string(s[:8])
}

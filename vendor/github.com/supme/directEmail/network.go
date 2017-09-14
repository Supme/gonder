package directEmail

import (
	"fmt"
	"errors"
	"net"
)

// SetInterfaceDefault set default interface for sending
func (self *Email) SetInterfaceDefault(ip string) {
	self.Ip = ""
}

// SetInterfaceByIp set IP from which the sending will be made
func (self *Email) SetInterfaceByIp(ip string) {
	self.Ip = ip
}

// SetInterfaceByName set interface from which the sending will be made
func (self *Email) SetInterfaceByName(name string) error {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return err
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return err
	}
	if len(addrs) > 1 {
		fmt.Printf("%+v", addrs)
		return errors.New("Interface have more then one address")
	}
	self.Ip = addrs[0].String()
	return nil
}

// // SetInterfaceBySocks set SOCKS server through which the sending will be made
func (self *Email) SetInterfaceSocks(server string, port int) {
	self.Ip = fmt.Sprintf("socks://%s:%d", server, port)
}

// SetMapGlobalIpForLocal set glibal IP for local IP address
func (self *Email) SetMapGlobalIpForLocal(globalIp, localIp string) {
	self.MapIp[localIp] = globalIp
}

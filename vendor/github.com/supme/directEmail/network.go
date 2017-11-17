package directEmail

import (
	"errors"
	"fmt"
	"net"
)

// SetInterfaceDefault set default interface for sending
func (slf *Email) SetInterfaceDefault(ip string) {
	slf.Ip = ""
}

// SetInterfaceByIp set IP from which the sending will be made
func (slf *Email) SetInterfaceByIp(ip string) {
	slf.Ip = ip
}

// SetInterfaceByName set interface from which the sending will be made
func (slf *Email) SetInterfaceByName(name string) error {
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
	slf.Ip = addrs[0].String()
	return nil
}

// // SetInterfaceBySocks set SOCKS server through which the sending will be made
func (slf *Email) SetInterfaceSocks(server string, port int) {
	slf.Ip = fmt.Sprintf("socks://%s:%d", server, port)
}

// SetMapGlobalIpForLocal set glibal IP for local IP address
func (slf *Email) SetMapGlobalIpForLocal(globalIp, localIp string) {
	slf.MapIp[localIp] = globalIp
}

package smtpSender

import (
	"fmt"
	"golang.org/x/net/proxy"
	"net"
	"net/smtp"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	dialTimeout = 60 * time.Second
	dialTries   = 3
	connTimeout = 5 * time.Minute // SMTP RFC 5 min
	connTries   = 5
)

// Connect to smtp server from configured interface
type Connect struct {
	iface    string
	hostname string
	portSMTP int
	mapIP    map[string]string
}

// SetMapIP if use NAT set global IP address
func (c *Connect) SetMapIP(localIP, globalIP string) {
	if c.mapIP == nil {
		c.mapIP = map[string]string{}
	}
	c.mapIP[localIP] = globalIP
}

// SetSMTPport set SMTP server port. Default 25
// use for translate local Iface to global if NAT
// if use Socks server translate Iface SOCKS server to real Iface
func (c *Connect) SetSMTPport(port int) {
	c.portSMTP = port
}

// SetIface use this Iface for send. Default use default interface.
// Example:
//   IP "1.2.3.4"
//   Socks5 proxy "socks5://user:password@1.2.3.4" or "socks5://1.2.3.4"
func (c *Connect) SetIface(iface string) {
	c.iface = iface
}

// SetHostName set server hostname for HELO. If left blanc then use resolv name.
func (c *Connect) SetHostName(name string) {
	c.hostname = name
}

func (c *Connect) newClient(domain string, lookupMX bool) (client *smtp.Client, err error) {
	var (
		dialer func(network, address string) (net.Conn, error)
		mxs    []*net.MX
	)

	if c.portSMTP == 0 {
		c.portSMTP = 25
	}

	if dialer, err = dialFunction(c.iface); err != nil {
		return nil, err
	}

	if lookupMX {
		for tries := 0; tries < dialTries; tries++ {
			mxs, err = net.LookupMX(domain)
			if err == nil {
				break
			}
			time.Sleep(time.Second)
		}
	} else {
		mxs = append(mxs, &net.MX{Host: domain, Pref: 10})
	}

	if len(mxs) == 0 {
		return nil, fmt.Errorf("Max MX lookup tries reached")
	}

	for i := range mxs {
		var conn net.Conn
		server := strings.TrimSpace(mxs[i].Host)
		for tries := 1; tries <= connTries; tries++ {
			conn, err = dialer("tcp", net.JoinHostPort(server, strconv.Itoa(c.portSMTP)))
			if err == nil {
				e := conn.SetDeadline(time.Now().Add(connTimeout))
				if e != nil {
					return nil, e
				}
				e = conn.SetReadDeadline(time.Now().Add(connTimeout))
				if e != nil {
					return nil, e
				}
				e = conn.SetWriteDeadline(time.Now().Add(connTimeout))
				if e != nil {
					return nil, e
				}
				break
			}
			if tries != connTries {
				time.Sleep(5 * time.Second)
			}
		}
		if err != nil {
			return
		}

		var ip string
		if c.iface == "" {
			ip, _, err = net.SplitHostPort(conn.LocalAddr().String())
			if err != nil {
				return nil, err
			}
		} else {
			var u *url.URL
			u, err = url.Parse(c.iface)
			if strings.ToLower(u.Scheme) == "socks" || strings.ToLower(u.Scheme) == "socks5" {
				ip, _, err = net.SplitHostPort(u.Host)
				if err != nil {
					return nil, err
				}
			}
		}
		if myGlobalIP, ok := c.mapIP[ip]; ok {
			ip = myGlobalIP
		}
		if c.hostname == "" {
			name, err := lookup(ip)
			if err != nil {
				return nil, err
			}
			c.hostname = name
		}

		client, err = smtp.NewClient(conn, server)
		if err != nil {
			continue
		}
		err = client.Hello(strings.TrimRight(c.hostname, "."))
		if err == nil {
			break
		}
	}

	if err != nil {
		return
	}

	return
}

var resolvedHosts struct {
	host map[string]string
	sync.Mutex
}

func lookup(ip string) (string, error) {
	resolvedHosts.Lock()
	defer resolvedHosts.Unlock()
	if resolvedHosts.host == nil {
		resolvedHosts.host = map[string]string{}
	}
	if name, ok := resolvedHosts.host[ip]; ok {
		return name, nil
	}
	names, err := net.LookupAddr(ip)
	if err != nil {
		return "", err
	}
	resolvedHosts.host[ip] = names[0]
	return names[0], nil
}

func dialFunction(iface string) (dialFunc func(network, address string) (net.Conn, error), err error) {
	if iface == "" {
		iface := net.Dialer{
			Timeout: dialTimeout,
		}
		dialFunc = iface.Dial
	} else {
		var u *url.URL
		u, err = url.Parse(iface)
		if strings.ToLower(u.Scheme) == "socks" || strings.ToLower(u.Scheme) == "socks5" {
			if err != nil {
				return nil, fmt.Errorf("Error parse socks: %s", err.Error())
			}
			var iface proxy.Dialer
			if u.User != nil {
				auth := proxy.Auth{}
				auth.User = u.User.Username()
				auth.Password, _ = u.User.Password()
				iface, err = proxy.SOCKS5("tcp", u.Host, &auth, proxy.FromEnvironment())
				if err != nil {
					return
				}
			} else {
				iface, err = proxy.SOCKS5("tcp", u.Host, nil, proxy.FromEnvironment())
				if err != nil {
					return
				}
			}
			dialFunc = iface.Dial
		} else {
			iface := net.Dialer{
				LocalAddr: &net.TCPAddr{IP: net.ParseIP(iface)},
				Timeout:   dialTimeout,
			}
			dialFunc = iface.Dial
		}
	}
	return
}

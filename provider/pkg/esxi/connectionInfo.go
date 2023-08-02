package esxi

import (
	"fmt"
)

type ConnectionInfo struct {
	Host        string
	SSHPort     string
	SslPort     string
	UserName    string
	Password    string
	OvfLocation string
}

func (c *ConnectionInfo) getSSHConnection() string {
	return fmt.Sprintf("%s:%s", c.Host, c.SSHPort)
}

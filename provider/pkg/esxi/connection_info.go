package esxi

import (
	"fmt"
)

type ConnectionInfo struct {
	Host        string
	SshPort     string
	SslPort     string
	UserName    string
	Password    string
	OvfLocation string
}

func (c *ConnectionInfo) getSshConnection() string {
	return fmt.Sprintf("%s:%s", c.Host, c.SshPort)
}

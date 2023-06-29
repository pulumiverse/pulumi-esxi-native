package esxi

import (
	"fmt"
	"github.com/golang/glog"
	"os"
	"strings"
	"time"

	"github.com/tmc/scp"
	"golang.org/x/crypto/ssh"
)

type Host struct {
	ClientConfig *ssh.ClientConfig
	Connection   *ConnectionInfo
}

func NewHost(host, sshPort, sslPort, user, pass, ovfLoc string) Host {
	connection := ConnectionInfo{
		Host:        host,
		SshPort:     sshPort,
		SslPort:     sslPort,
		UserName:    user,
		Password:    pass,
		OvfLocation: ovfLoc,
	}
	clientConfig := &ssh.ClientConfig{
		User: connection.UserName,
		Auth: []ssh.AuthMethod{
			ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) ([]string, error) {
				// Reply password to all questions
				answers := make([]string, len(questions))
				for i := range answers {
					answers[i] = connection.Password
				}

				return answers, nil
			}),
		},
	}
	clientConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	return Host{
		Connection:   &connection,
		ClientConfig: clientConfig,
	}
}

func (h *Host) ValidateCreds() error {
	var remoteCmd string
	var err error

	remoteCmd = fmt.Sprintf("vmware --version")
	_, err = h.Execute(remoteCmd, "Connectivity test, get vmware version")
	if err != nil {
		return fmt.Errorf("Failed to connect to esxi host: %s\n", err)
	}

	mkdir, err := h.Execute("mkdir -p ~", "Create home directory if missing")
	glog.V(9).Infof("ValidateCreds: Create home! %s %s", mkdir, err)

	if err != nil {
		return err
	}
	if err != nil {
		return err
	}

	return nil
}

// Connect to esxi host using ssh
func (h *Host) Connect(attempt int) (*ssh.Client, *ssh.Session, error) {
	//attempt := 10
	for attempt > 0 {
		client, err := ssh.Dial("tcp", h.Connection.getSshConnection(), h.ClientConfig)
		if err != nil {
			glog.V(9).Infof("Connect: Retry attempt %d", attempt)
			attempt -= 1
			time.Sleep(1 * time.Second)
		} else {

			session, err := client.NewSession()
			if err != nil {
				closeErr := client.Close()
				if closeErr != nil {
					return nil, nil, fmt.Errorf("session connection error. (closing client error: %s)", closeErr)
				}
				return nil, nil, fmt.Errorf("session connection error")
			}

			return client, session, nil

		}
	}
	return nil, nil, fmt.Errorf("client connection error")
}

func (h *Host) Execute(command string, shortCmdDesc string) (string, error) {
	glog.V(9).Infof("Execute: %s", shortCmdDesc)

	var attempt int

	if command == "vmware --version" {
		attempt = 3
	} else {
		attempt = 10
	}
	client, session, err := h.Connect(attempt)
	if err != nil {
		glog.V(9).Infof("Execute: Failed connecting to host! %s", err.Error())
		return "Failed to connect to esxi host", err
	}

	stdoutRaw, err := session.CombinedOutput(command)
	stdout := strings.TrimSpace(string(stdoutRaw))

	if stdout == "<unset>" {
		return "Failed to connect to esxi host or Management Agent has been restarted", err
	}

	logMessage := fmt.Sprintf("Execute: cmd => %s", command)
	if len(stdout) > 0 {
		logMessage = fmt.Sprintf("%s\n\tstdout => %s\n", logMessage, stdout)
	}
	if err != nil {
		logMessage = fmt.Sprintf("%s\tstderr => %s\n", logMessage, err)
	}
	glog.V(9).Infof(logMessage)

	closeErr := client.Close()
	if closeErr != nil {
		return "", closeErr
	}
	return stdout, closeErr
}

func (h *Host) CopyFile(content string, path string, shortCmdDesc string) (string, error) {
	glog.V(9).Infof("CopyFile: %s", shortCmdDesc)

	f, _ := os.CreateTemp("", "")
	_, err := fmt.Fprintln(f, content)
	if err != nil {
		return "", err
	}
	fCloseErr := f.Close()
	if fCloseErr != nil {
		return "", fCloseErr
	}
	defer os.Remove(f.Name())

	client, session, err := h.Connect(10)
	if err != nil {
		glog.V(9).Infof("CopyFile: Failed connecting to host! %s", err.Error())
		return "Failed connection to host!", err
	}

	err = scp.CopyPath(f.Name(), path, session)
	if err != nil {
		glog.V(9).Infof("CopyFile: Failed copying the file! %s", err.Error())
		return "Failed to copy file to esxi host!", err
	}

	closeErr := client.Close()
	if closeErr != nil {
		return "", closeErr
	}

	return content, err
}

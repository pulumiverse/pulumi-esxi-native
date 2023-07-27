package esxi

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/common/util/logging"
	"github.com/tmc/scp"
	"golang.org/x/crypto/ssh"
)

type Host struct {
	ClientConfig *ssh.ClientConfig
	Connection   *ConnectionInfo
}

func NewHost(host, sshPort, sslPort, user, pass, ovfLoc string) (*Host, error) {
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

	instance := &Host{
		Connection:   &connection,
		ClientConfig: clientConfig,
	}

	err := instance.validateCreds()
	if err != nil {
		return nil, err
	}

	return instance, nil
}

func (esxi *Host) validateCreds() error {
	var remoteCmd string
	var err error

	remoteCmd = fmt.Sprintf("vmware --version")
	_, err = esxi.Execute(remoteCmd, "Connectivity test, get vmware version")
	if err != nil {
		return fmt.Errorf("failed to connect to esxi host: %s", err)
	}

	mkdir, err := esxi.Execute("mkdir -p ~", "Create home directory if missing")
	logging.V(9).Infof("ValidateCreds: Create home! %s %s", mkdir, err)

	if err != nil {
		return err
	}

	return nil
}

// Connect to esxi host using ssh
func (esxi *Host) connect(attempt int) (*ssh.Client, *ssh.Session, error) {
	for attempt > 0 {
		client, err := ssh.Dial("tcp", esxi.Connection.getSshConnection(), esxi.ClientConfig)
		if err != nil {
			logging.V(9).Infof("Connect: Retry attempt %d", attempt)
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

func (esxi *Host) Execute(command string, shortCmdDesc string) (string, error) {
	logging.V(9).Infof("Execute: %s", shortCmdDesc)

	var attempt int

	if command == "vmware --version" {
		attempt = 3
	} else {
		attempt = 10
	}
	client, session, err := esxi.connect(attempt)
	if err != nil {
		logging.V(9).Infof("Execute: Failed connecting to host! %s", err)
		return "failed to connect to esxi host", err
	}

	stdoutRaw, err := session.CombinedOutput(command)
	stdout := strings.TrimSpace(string(stdoutRaw))

	if stdout == "<unset>" {
		return "failed to connect to esxi host or Management Agent has been restarted", err
	}

	logMessage := fmt.Sprintf("Execute: cmd => %s", command)
	if len(stdout) > 0 {
		logMessage = fmt.Sprintf("%s\n\tstdout => %s\n", logMessage, stdout)
	}
	if err != nil {
		logMessage = fmt.Sprintf("%s\tstderr => %s\n", logMessage, err)
	}
	logging.V(9).Infof(logMessage)
	if closeErr := client.Close(); closeErr != nil {
		logging.V(9).Infof("Failed closing the client connection to host! %s", closeErr)
	}

	return stdout, err
}

func (esxi *Host) WriteFile(content string, path string, shortCmdDesc string) (string, error) {
	logging.V(9).Infof("WriteFile: %s", shortCmdDesc)

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

	client, session, err := esxi.connect(10)
	if err != nil {
		logging.V(9).Infof("Execute: Failed connecting to host! %s", err)
		return "failed to connect to esxi host", err
	}
	err = scp.CopyPath(f.Name(), path, session)
	if err != nil {
		logging.V(9).Infof("WriteFile: Failed copying the file! %s", err)
		return "failed to copy file to esxi host", err
	}
	if closeErr := client.Close(); closeErr != nil {
		logging.V(9).Infof("Failed closing the client connection to host! %s", closeErr)
	}

	return content, err
}

func (esxi *Host) CopyFile(localPath string, hostPath string, shortCmdDesc string) (string, error) {
	logging.V(9).Infof("CopyFile: %s", shortCmdDesc)

	client, session, err := esxi.connect(10)
	if err != nil {
		logging.V(9).Infof("Execute: Failed connecting to host! %s", err)
		return "failed to connect to esxi host", err
	}
	err = scp.CopyPath(localPath, hostPath, session)
	if err != nil {
		return "Failed to copy file to esxi host!", err
	}
	if closeErr := client.Close(); closeErr != nil {
		logging.V(9).Infof("Failed closing the client connection to host! %s", closeErr)
	}

	return "", nil
}

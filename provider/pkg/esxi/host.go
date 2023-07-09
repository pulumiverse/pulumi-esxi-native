package esxi

import (
	"fmt"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/logging"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/tmc/scp"
	"golang.org/x/crypto/ssh"
)

type Host struct {
	ClientConfig *ssh.ClientConfig
	Client       *ssh.Client
	Session      *ssh.Session
	Connection   *ConnectionInfo
}

var lock = &sync.Mutex{}
var instance *Host

func NewHost(host, sshPort, sslPort, user, pass, ovfLoc string) (*Host, error) {
	if instance == nil {
		lock.Lock()
		defer lock.Unlock()
		if instance == nil {
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

			instance = &Host{
				Connection:   &connection,
				ClientConfig: clientConfig,
			}

			err := connect(5)
			if err != nil {
				logging.V(9).Infof("Failed connecting to host! %s", err)
				return nil, fmt.Errorf("failed to connect to esxi host; err: %s", err)
			}
			defer disconnect()

			err = instance.validateCreds()
			if err != nil {
				return nil, err
			}
		} else {
			logging.V(9).Infof("Host instance already created.")
		}
	} else {
		logging.V(9).Infof("Host instance already created.")
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
func connect(attempt int) error {
	//attempt := 10
	for attempt > 0 {
		client, err := ssh.Dial("tcp", instance.Connection.getSshConnection(), instance.ClientConfig)
		if err != nil {
			logging.V(9).Infof("Connect: Retry attempt %d", attempt)
			attempt -= 1
			time.Sleep(1 * time.Second)
		} else {

			session, err := client.NewSession()
			if err != nil {
				closeErr := client.Close()
				if closeErr != nil {
					return fmt.Errorf("session connection error. (closing client error: %s)", closeErr)
				}
				return fmt.Errorf("session connection error")
			}

			instance.Client = client
			instance.Session = session

			return nil
		}
	}
	return fmt.Errorf("client connection error")
}

func disconnect() {
	if instance.Client != nil {
		logging.V(9).Infof("Disconnecting SSH Client from the Host...")
		err := instance.Client.Close()
		if err != nil {
			logging.V(9).Infof("Failed closing the client connection to host! %s", err)
		}
	}
}

func (esxi *Host) Execute(command string, shortCmdDesc string) (string, error) {
	logging.V(9).Infof("Execute: %s", shortCmdDesc)

	retried := false
	for {
		// Code to be executed indefinitely
		stdoutRaw, err := esxi.Session.CombinedOutput(command)
		stdout := strings.TrimSpace(string(stdoutRaw))

		if stdout == "<unset>" && retried == false {
			retried = true
			attempt := 10
			if command == "vmware --version" {
				attempt = 3
			}
			err = connect(attempt)
			if err != nil {
				logging.V(9).Infof("Execute: Failed connecting to host! %s", err)
				return "failed to connect to esxi host", err
			}
		} else if stdout == "<unset>" && retried {
			return "failed to connect to esxi host or Management Agent has been restarted", err
		} else {
			logMessage := fmt.Sprintf("Execute: cmd => %s", command)
			if len(stdout) > 0 {
				logMessage = fmt.Sprintf("%s\n\tstdout => %s\n", logMessage, stdout)
			}
			if err != nil {
				logMessage = fmt.Sprintf("%s\tstderr => %s\n", logMessage, err)
			}
			logging.V(9).Infof(logMessage)
			return stdout, err
		}
	}
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

	err = scp.CopyPath(f.Name(), path, esxi.Session)
	if err != nil {
		logging.V(9).Infof("WriteFile: Failed copying the file! %s", err)
		return "Failed to copy file to esxi host!", err
	}

	return content, err
}

func (esxi *Host) CopyFile(localPath string, hostPath string, shortCmdDesc string) (string, error) {
	logging.V(9).Infof("CopyFile: %s", shortCmdDesc)

	err := scp.CopyPath(localPath, hostPath, esxi.Session)
	if err != nil {
		return "Failed to copy file to esxi host!", err
	}

	return "", nil
}

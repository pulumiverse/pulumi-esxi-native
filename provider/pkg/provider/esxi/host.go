package esxi

import (
	"fmt"
	"io/ioutil"
	"log"
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
	log.Printf("[validateEsxiCreds]\n")

	var remoteCmd string
	var err error

	remoteCmd = fmt.Sprintf("vmware --version")
	_, err = h.execute(remoteCmd, "Connectivity test, get vmware version")
	if err != nil {
		return fmt.Errorf("Failed to connect to esxi host: %s\n", err)
	}

	h.execute("mkdir -p ~", "Create home directory if missing")

	return nil
}

// connect to esxi host using ssh
func (h *Host) connect(attempt int) (*ssh.Client, *ssh.Session, error) {
	//attempt := 10
	for attempt > 0 {
		client, err := ssh.Dial("tcp", h.Connection.getSshConnection(), h.ClientConfig)
		if err != nil {
			log.Printf("[runRemoteSshCommand] Retry connection: %d\n", attempt)
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

func (h *Host) execute(command string, shortCmdDesc string) (string, error) {
	log.Println("[execute] :" + shortCmdDesc)

	var attempt int

	if command == "vmware --version" {
		attempt = 3
	} else {
		attempt = 10
	}
	client, session, err := h.connect(attempt)
	if err != nil {
		log.Println("[execute] Failed err: " + err.Error())
		return "Failed to ssh to esxi host", err
	}

	stdoutRaw, err := session.CombinedOutput(command)
	stdout := strings.TrimSpace(string(stdoutRaw))

	if stdout == "<unset>" {
		return "Failed to ssh to esxi host or Management Agent has been restarted", err
	}

	log.Printf("[execute] cmd:/%s/\n stdout:/%s/\nstderr:/%s/\n", command, stdout, err)

	closeErr := client.Close()
	if closeErr != nil {
		return "", closeErr
	}
	return stdout, closeErr
}

//  Function to scp file to esxi host.
func (h *Host) WriteContentToFile(content string, path string, shortCmdDesc string) (string, error) {
	log.Println("[writeContentToRemoteFile] :" + shortCmdDesc)

	f, _ := ioutil.TempFile("", "")
	_, err := fmt.Fprintln(f, content)
	if err != nil {
		return "", err
	}
	fCloseErr := f.Close()
	if fCloseErr != nil {
		return "", fCloseErr
	}
	defer os.Remove(f.Name())

	client, session, err := h.connect(10)
	if err != nil {
		log.Println("[writeContentToRemoteFile] Failed err: " + err.Error())
		return "Failed to ssh to esxi host", err
	}

	err = scp.CopyPath(f.Name(), path, session)
	if err != nil {
		log.Println("[writeContentToRemoteFile] Failed err: " + err.Error())
		return "Failed to scp file to esxi host", err
	}

	closeErr := client.Close()
	if closeErr != nil {
		return "", closeErr
	}

	return content, err
}

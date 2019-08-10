package mysql

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	"github.com/mugli/go-kill-mysql-query/configuration"

	"github.com/jmoiron/sqlx"
	"github.com/rhysd/abspath"
)

type SSHDialer struct {
	client *ssh.Client
}

func (self *SSHDialer) Dial(addr string) (net.Conn, error) {
	return self.client.Dial("tcp", addr)
}

var (
	sshClientConn *ssh.Client
	dbConn        *sqlx.DB
)

func Disconnect() {
	if dbConn != nil {
		err := dbConn.Close()

		if err != nil {
			fmt.Printf("Failure: %s", err.Error())
		}
	}

	if sshClientConn != nil {
		err := sshClientConn.Close()

		if err != nil {
			fmt.Printf("Failure: %s", err.Error())
		}
	}
}

func sshAgent() (ssh.AuthMethod, error) {
	sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))

	if err != nil {
		return nil, err
	}

	authMethod := ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
	return authMethod, nil
}

func Connect(config configuration.Config) (*sqlx.DB, error) {
	useSSHTunnel := config.SSH.UseTunnel
	sshHost := config.SSH.Host     // SSH Server Hostname/IP
	sshPort := config.SSH.Port     // SSH Port
	sshUser := config.SSH.Username // SSH Username
	sshPass := config.SSH.Password // Empty string for no password
	sshKey := config.SSH.Key
	sshKeyPassphrase := config.SSH.KeyPassphrase

	dbUser := config.MySQL.Username // DB username
	dbPass := config.MySQL.Password // DB Password
	dbHost := config.MySQL.Host     // DB Hostname/IP
	dbPort := config.MySQL.Port     // DB Port

	var network string

	fmt.Println("🔌	Connecting to database...")
	if useSSHTunnel {
		sshConfig := &ssh.ClientConfig{
			User:            sshUser,
			Auth:            []ssh.AuthMethod{},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}

		if sshKey != "" {
			keyPath, err := abspath.ExpandFrom(sshKey)
			if err != nil {
				return nil, err
			}

			key, err := ioutil.ReadFile(keyPath.String())

			if err != nil {
				return nil, err
			}

			var signer ssh.Signer
			if sshKeyPassphrase != "" {
				signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(sshKeyPassphrase))
			} else {
				signer, err = ssh.ParsePrivateKey(key)
			}

			if err != nil {
				return nil, err
			}

			sshConfig.Auth = append(sshConfig.Auth, ssh.PublicKeys(signer))
		}

		if sshPass != "" {
			sshConfig.Auth = append(sshConfig.Auth, ssh.Password(sshPass))
		}

		if sshKey == "" && sshPass == "" {
			// Try using ssh-agent
			agent, err := sshAgent()
			if err != nil {
				return nil, err
			}

			sshConfig.Auth = append(sshConfig.Auth, agent)
		}

		// Connect to the SSH Server
		sshClientConn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", sshHost, sshPort), sshConfig)

		if err != nil {
			return nil, err
		}

		// Register the SSHDialer with the ssh connection as a parameter
		mysql.RegisterDial("mysql+tcp", (&SSHDialer{sshClientConn}).Dial)

		network = "mysql+tcp"
	}

	connString := fmt.Sprintf("%s:%s@%s(%s:%d)/", dbUser, dbPass, network, dbHost, dbPort)
	dbConn, err := sqlx.Connect("mysql", connString)

	if err != nil {
		return nil, err
	}

	fmt.Println("⚡️	Successfully connected to the database")
	return dbConn, nil
}

package main

import (
	"errors"
	"log"
	"net"
	"os"
	"sync"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type SSHConnection struct {
	SFTPClient *sftp.Client
	SSHClient  *ssh.Client
}

type ServerConnections struct {
	servers map[string]*SSHConnection
	m       sync.Mutex
}

func NewServerConnections() *ServerConnections {
	return &ServerConnections{
		servers: make(map[string]*SSHConnection),
	}
}

func (s *ServerConnections) AddServer(server string, connection *SSHConnection) {
	s.m.Lock()
	defer s.m.Unlock()
	s.servers[server] = connection
}

func (s *ServerConnections) RemoveServer(server string) {
	s.m.Lock()
	defer s.m.Unlock()
	delete(s.servers, server)
}

func (s *ServerConnections) GetServer(server string) (*SSHConnection, error) {

	conn, ok := s.servers[server]
	if !ok {
		log.Printf("not found in map. connecting...")
		matches := serverRegex.FindStringSubmatch(server)
		if len(matches) != 3 {
			return nil, errors.New("invalid server format. user@domain.com, user@12.23.45.56")
		}
		user, host := matches[1], matches[2]
		newConn, err := s.Connect(user, host)
		if err != nil {
			return nil, err
		}
		conn = newConn

		s.AddServer(server, conn)
	}
	return conn, nil
}

func (s *ServerConnections) Connect(user, host string) (*SSHConnection, error) {

	agentConn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		return nil, err
	}
	defer agentConn.Close()

	sshAgent := agent.NewClient(agentConn)
	singers, err := sshAgent.Signers()
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			// Use the PublicKeys method for remote authentication.
			ssh.PublicKeys(singers...),
		},
	}

	conn, err := ssh.Dial("tcp", host, config)
	if err != nil {
		return nil, err
	}
	// defer conn.Close()

	client, err := sftp.NewClient(conn)
	if err != nil {
		return nil, err
	}

	return &SSHConnection{
		SSHClient:  conn,
		SFTPClient: client,
	}, nil

}

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/crypto/ssh"

	"github.com/neovim/go-client/nvim/plugin"
)

var servers *ServerConnections
var serverRegex = regexp.MustCompile(`^([^@]*)@(.*)$`)

func hello(args []string) (string, error) {
	return "Hello " + strings.Join(args, " "), nil
}

func main() {

	servers = NewServerConnections()

	plugin.Main(func(p *plugin.Plugin) error {
		p.HandleFunction(&plugin.FunctionOptions{Name: "Upload"}, upload)
		return nil
	})
}

func parseConfig(path string) (*Config, error) {

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	config := Config{}

	err = json.NewDecoder(file).Decode(&config)
	return &config, err
}

func upload(args []string) (string, error) {

	if len(args) != 3 {
		return "", errors.New("missing args")
	}

	configPath := args[0]
	local := args[1]
	remote_relative := args[2]

	config, err := parseConfig(configPath)
	if err != nil {
		return "", errors.New("error parsing config file")
	}

	server := fmt.Sprintf("%s@%s:%d", config.User, config.Host, config.Port)
	remote := filepath.Join(config.RemotePath, remote_relative)

	localFileInfo, err := os.Stat(local)
	if err != nil {
		return "", err //errors.New("local file not found")
	}

	if localFileInfo.IsDir() {
		return "", errors.New("local file is directory")
	}

	client, err := servers.GetServer(server)
	if err != nil {
		return "", errors.New("unable to connect to server")
	}

	retryCount := 0
	// make better
retry:
	remoteFile, err := client.SFTPClient.Create(remote)
	switch err {
	case nil:
		break
	case os.ErrNotExist:
		if retryCount == 0 {
			retryCount += 1
			err = MkDirRecursive(client.SSHClient, remote)
			if err != nil {
				return "", errors.New("mkdir fail on remote")
			}
			goto retry
		} else {
			return "", errors.New("mkdir fail on remote")
		}
	case io.EOF:
		servers.RemoveServer(server)
		return "", errors.New("error creating file on remote. deleting server from cache")
	default:
		return "", errors.New("error creating file on remote")

	}
	defer remoteFile.Close()

	localFile, err := os.Open(local)
	if err != nil {
		return "", errors.New("unable to open local file")
	}
	defer localFile.Close()

	n, err := remoteFile.ReadFrom(localFile)
	if err != nil {
		return "", errors.New("unable to open remote file")
	}

	err = remoteFile.Chmod(localFileInfo.Mode())
	if err != nil {
		return "", errors.New("unable to set filemode on remote file")
	}

	return fmt.Sprintf("wrote %d bytes to %s:%s\n", n, server, remote), nil

}

func MkDirRecursive(client *ssh.Client, path string) error {

	parent := filepath.Dir(path)
	session, err := client.NewSession()
	if err != nil {
		return err
	}

	err = session.Run("mkdir -p " + parent)
	if err != nil {
		return err
	}

	return nil

}

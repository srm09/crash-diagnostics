// Copyright (c) 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package testing

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/vladimirvivien/echo"
)

type SSHServer struct {
	name        string
	port        string
	keyMountDir string
	username    string
	e           *echo.Echo
}

func NewSSHServer(name, username, port string) (*SSHServer, error) {
	mountDir, err := genSSHKeys()
	if err != nil {
		return nil, err
	}
	return &SSHServer{name: name, port: port, keyMountDir: mountDir, username: username, e: echo.New()}, nil
}

func genSSHKeys() (string, error) {
	_, b, _, _ := runtime.Caller(0)
	d := path.Join(path.Dir(b))
	tmpDir := path.Join(filepath.Dir(d), ".testing")
	if err := os.MkdirAll(tmpDir, 0777); err != nil {
		logrus.Error(err)
		return "", err
	}
	err := GenerateKeyPair(tmpDir)
	return tmpDir, err
}

// StartSSHServer starts starts sshd process using image linuxserver/openssh-server.DockerRunSSH
// The server opnes up port 2222 with the following command
/*

docker create \
  --name=test-sshd \
  -e PUBLIC_KEY_FILE=$HOME/.ssh/id_rsa.pub \
  -e USER_NAME=$USER \
  -e SUDO_ACCESS=true \
  -p 2222:2222 \
  -v ./crash-diagnostics/testing/keys:/config
  linuxserver/openssh-server

*/
func (s *SSHServer) Start() error {
	if len(s.e.Prog.Avail("docker")) == 0 {
		return fmt.Errorf("docker command not found")
	}

	if strings.Contains(s.e.Run("docker ps"), s.name) {
		logrus.Info("Skipping SSHServer.Start, container already running:", s.name)
		return nil
	}

	s.e.SetVar("CONTAINER_NAME", s.name)
	s.e.SetVar("SSH_PORT", fmt.Sprintf("%s:2222", s.port))
	s.e.SetVar("SSH_DOCKER_IMAGE", "vladimirvivien/openssh-server")
	s.e.SetVar("USERNAME", s.username)
	s.e.SetVar("KEY_VOLUME_MOUNT", s.keyMountDir)

	cmd := s.e.Eval("docker run --rm --detach --name=$CONTAINER_NAME -p $SSH_PORT -e PUBLIC_KEY_FILE=/config/id_rsa.pub -e USER_NAME=$USERNAME -e SUDO_ACCESS=true -v $KEY_VOLUME_MOUNT:/config $SSH_DOCKER_IMAGE")
	logrus.Debugf("Starting SSH server: %s", cmd)
	proc := s.e.RunProc(cmd)
	result := proc.Result()
	if proc.Err() != nil {
		msg := fmt.Sprintf("%s: %s", proc.Err(), result)
		return fmt.Errorf(msg)
	}
	logrus.Infof("SSH server container started: name=%s, port=%s (docker id - %s)", s.name, s.port, result)

	return nil
}

func (s *SSHServer) Stop() error {
	if len(s.e.Prog.Avail("docker")) == 0 {
		return fmt.Errorf("docker command not found")
	}

	s.e.SetVar("CONTAINER_NAME", s.name)

	if !strings.Contains(s.e.Run("docker ps"), s.name) {
		logrus.Info("Skipping SSHServerStop, container not running:", s.name)
		return nil
	}

	proc := s.e.RunProc("docker stop $CONTAINER_NAME")
	result := proc.Result()
	if proc.Err() != nil {
		msg := fmt.Sprintf("failed to stop container: %s: %s", proc.Err(), result)
		return fmt.Errorf(msg)
	}
	logrus.Info("SSH server stopped: ", result)
	defer os.RemoveAll(s.keyMountDir)

	// attempt to remove container if still lingering
	if strings.Contains(s.e.Run("docker ps"), s.name) {
		logrus.Info("Forcing container removal:", s.name)
		proc := s.e.RunProc("docker rm --force $CONTAINER_NAME")
		result := proc.Result()
		if proc.Err() != nil {
			msg := fmt.Sprintf("failed to remove container: %s: %s", proc.Err(), result)
			return fmt.Errorf(msg)
		}
		logrus.Info("SSH server container removed: ", result)
	}

	return nil
}

func (s *SSHServer) PrivateKey() string {
	return filepath.Join(s.keyMountDir, "id_rsa")
}

package nwssh

import (
	"time"
	"errors"
)

type H3cSSH struct {
	*SSHBase
}

func (s *H3cSSH) SessionPreparation() bool {

	if !s.preparateWriting() {
		return false
	}

	if !s.disablePaging("screen-length disable") {
		return false
	}

	s.clearBuffer()

	return true
}

func (s *H3cSSH) SaveRuningConfig() bool {
	_, err := s.ExecCommandExpectPrompt("save force", time.Second*20)
	if err != nil {
		return false
	}
	return true
}

func (s *H3cSSH) InterfaceConfig() (string, error) {
	resp, err := s.ExecCommand("display cu interface")
	if err != nil {
		return "", err
	}
	return resp, err
}


func (s *H3cSSH) RunTranscation(trans string) (string, error){
	if trans == "ifconfig"{
		return s.InterfaceConfig()
	}

	return "", errors.New("Unsupport transcation!")
}
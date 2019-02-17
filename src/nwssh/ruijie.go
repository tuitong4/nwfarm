package nwssh

import (
	"errors"
	"log"
	"time"
)

type RuijieSSH struct {
	*SSHBase
}

func (s *RuijieSSH) SessionPreparation() bool {

	if !s.preparateWriting() {
		return false
	}

	if !s.disablePaging("terminal length 0") {
		return false
	}

	s.clearBuffer()

	return true
}

func (s *RuijieSSH) SaveRuningConfig() bool {

	_, err := s.ExecCommandExpectPrompt("copy running-config startup-config", time.Second*20)

	if err != nil {
		log.Fatalf("%v", err)
		return false
	}
	return true
}

func (s *RuijieSSH) InterfaceConfig() (string, error) {
	resp, err := s.ExecCommand("show running")
	if err != nil {
		return "", err
	}
	return resp, err
}

func (s *RuijieSSH) RunTranscation(trans string) (string, error) {
	if trans == "ifconfig" {
		return s.InterfaceConfig()
	}

	return "", errors.New("Unsupport transcation!")
}

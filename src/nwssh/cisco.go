package nwssh

import (
	"errors"
	"time"
)

type NexusSSH struct {
	*SSHBase
}

func (s *NexusSSH) SessionPreparation() bool {

	if !s.preparateWriting() {
		return false
	}

	if !s.disablePaging("terminal length 0") {
		return false
	}

	s.clearBuffer()

	return true
}

func (s *NexusSSH) SaveRuningConfig() bool {

	_, err := s.ExecCommandExpectPrompt("copy running-config startup-config", time.Second*20)

	if err != nil {
		return false
	}
	return true
}

func (s *NexusSSH) InterfaceConfig() (string, error) {
	resp, err := s.ExecCommand("show run interface")
	if err != nil {
		return "", err
	}
	return resp, err
}

func (s *NexusSSH) RunTranscation(trans string) (string, error) {
	if trans == "ifconfig" {
		return s.InterfaceConfig()
	}

	return "", errors.New("Unsupport transcation!")
}

type CiscoSSH struct {
	*SSHBase
}

func (s *CiscoSSH) SessionPreparation() bool {

	if !s.preparateWriting() {
		return false
	}

	if !s.disablePaging("terminal length 0") {
		return false
	}

	s.clearBuffer()

	return true
}

func (s *CiscoSSH) SaveRuningConfig() bool {
	_, err := s.ExecCommandExpect("copy running-config startup-config", "]?", time.Second*5)

	if err != nil {
		return false
	}
	_, err = s.ExecCommandExpectPrompt("", time.Second*20)

	if err != nil {
		return false
	}
	return true
}

func (s *CiscoSSH) InterfaceConfig() (string, error) {
	resp, err := s.ExecCommand("show running")
	if err != nil {
		return "", err
	}
	return resp, err
}

func (s *CiscoSSH) RunTranscation(trans string) (string, error) {
	if trans == "ifconfig" {
		return s.InterfaceConfig()
	}

	return "", errors.New("Unsupport transcation!")
}

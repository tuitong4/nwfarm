package nwssh

import (
	"log"
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
		log.Fatalf("%v", err)
		return false
	}
	return true
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
		log.Fatalf("%v", err)
		return false
	}
	_, err = s.ExecCommandExpectPrompt("", time.Second*20)

	if err != nil {
		log.Fatalf("%v", err)
		return false
	}
	return true
}

package nwssh

import (
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

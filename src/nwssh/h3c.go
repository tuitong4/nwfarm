package nwssh

import (
	"log"
	"time"
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
		log.Fatalf("%v", err)
		return false
	}
	return true
}

package nwssh

import (
	"log"
	"time"
)

type H3CSSH struct {
	*SSHBase
}

func (s *H3CSSH) SessionPreparation() bool {

	if !s.preparateWriting() {
		log.Fatalf("Unable to init the executable enviriment on remote terminal.")
		return false
	}

	if !s.disablePaging("screen-length disable") {
		return false
	}

	s.clearBuffer()

	return true
}

func (s *H3CSSH) SaveRuningConfig() bool {
	_, err := s.ExecCommandExpectPrompt("save force", time.Second*20)
	if err != nil {
		log.Fatalf("%v", err)
		return false
	}
	return true
}

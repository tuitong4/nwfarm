package nwssh

import (
	"log"
	"time"
)

type HuaweiSSH struct {
	*SSHBase
}

func (s *HuaweiSSH) SessionPreparation() bool {

	if !s.preparateWriting() {
		return false
	}

	if !s.disablePaging("screen-length 0 temporary") {
		return false
	}

	s.clearBuffer()

	return true
}

func (s *HuaweiSSH) SaveRuningConfig() bool {
	_, err := s.ExecCommandExpect("save", "[Y/N]:", time.Second*5)

	if err != nil {
		log.Fatalf("%v", err)
		return false
	}
	_, err = s.ExecCommandExpectPrompt("y", time.Second*5)

	if err != nil {
		log.Fatalf("%v", err)
		return false
	}
	return true
}

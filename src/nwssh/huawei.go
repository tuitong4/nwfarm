package nwssh

import (
	"time"
	"errors"
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
		return false
	}
	_, err = s.ExecCommandExpectPrompt("y", time.Second*5)

	if err != nil {
		return false
	}
	return true
}

func (s *HuaweiSSH) InterfaceConfig() (string, error) {
	resp, err := s.ExecCommand("display cu interface")
	if err != nil {
		return "", err
	}
	return resp, err
}


func (s *HuaweiSSH) RunTranscation(trans string) (string, error){
	if trans == "ifconfig"{
		return s.InterfaceConfig()
	}

	return "", errors.New("Unsupport transcation!")
}
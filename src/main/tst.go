package main

import (
	"fmt"
	"log"
	"nwssh"
	"syscall"
	"time"
)

func main() {
	sshoptions := nwssh.SSHOptions{
		PrivateKeyFile: "",
		IgnorHostKey:   true,
		BannerCallback: func(banner string) error {
			return nil
		},
		TermType:     "vt100",
		TermHeight:   560,
		TermWidht:    480,
		ReadWaitTime: time.Millisecond * 500, //Read data from a ssh channel timeout
	}

	device, err := nwssh.SSH("10.196.4.224", "22", "duanchengping", "dev&ops", 10, sshoptions)
	if err != nil {
		log.Fatalf("%v", err)
	}

	if err := device.Connect(); err != nil {
		log.Fatalf("Failed init ssh: %v", err)
	}

	h3c_sw := nwssh.RuijieSSH{device}
	defer h3c_sw.Close()

	if !h3c_sw.SessionPreparation() {
		log.Fatalf("Executable envirment unavalible!")
		syscall.Exit(1)
	}
	//fmt.Println(time.Now())
	resp := h3c_sw.SaveRuningConfig()
	//fmt.Println(time.Now())
	//resp = nwssh.SanitizeRespone(resp, true, true)
	//fmt.Println(time.Now())
	//if err != nil {
	//	log.Fatalf("Error occured when exec cmd: %v", err)
	//} else {
	//	fmt.Println(resp)
	//}
	fmt.Println(resp)
}

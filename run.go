package main

import (
	"fmt"
	"os"
	"time"
	//"io/ioutil"
	"golang.org/x/crypto/ssh"
	_ "log"
	"regexp"
	_ "strconv"
	"strings"
)

var VENDOR_FLAG map[string]string = map[string]string{
	"H3C":    "h3c",
	"HUAWEI": "huawei",
	"NEXUS":  "nexus",
	"CISCO":  "cisco",
	"RUIJIE": "ruijie"}

func LogAndRun(host string) {
	port := "22"
	vendor := ""
	username := ""
	pass := ""
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(pass),
		},
		Timeout:         30 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		BannerCallback: func(banner string) error {
			lowerbanner := strings.ToLower(banner)
			for k, v := range VENDOR_FLAG {
				matched, _ := regexp.MatchString(v, lowerbanner)
				if matched {
					vendor = k
					return nil
				}
			}
			return nil
		},

		Config: ssh.Config{Ciphers: append(sshConfig.Config.Ciphers, "aes128-cbc")},
	}
	var cmd string
	if vendor == "H3C" || vendor == "HUAWEI" {
		cmd = "display version"
	} else {
		cmd = "show version"
	}

	client, err := ssh.Dial("tcp", host+":"+port, config)
	if err != nil {
		fmt.Println(err)
	}
	defer client.Close()

	sess, err := client.NewSession()
	if err != nil {
		fmt.Println(err)
	}
	defer sess.Close()

	sess.Stdout = os.Stdout
	sess.Stderr = os.Stderr
	err = sess.Run(cmd)
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	LogAndRun("172.16.130.178")

}

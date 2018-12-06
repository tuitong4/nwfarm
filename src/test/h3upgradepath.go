package main

import (
	"bufio"
	"io/ioutil"
	"log"
	"nwssh"
	"os"
	"strings"
	"sync"
	"time"
)

func run(host, port string, sshoptions nwssh.SSHOptions) (string, bool) {

	var devssh *nwssh.SSHBase
	var err error

	username := ""
	password := ""

	devssh, err = nwssh.SSH(host, port, username, password, time.Duration(10)*time.Second, sshoptions)
	defer devssh.Close()
	if err != nil {
		log.Printf("[%s]%v\n", host, err)
		return "", false
	}
	if err = devssh.Connect(); err != nil {
		log.Printf("[%s]%v\n", host, err)
		return "", false
	}

	var device nwssh.SSHBASE

	device = &nwssh.H3cSSH{devssh}

	if !device.SessionPreparation() {
		log.Printf("[%s]Failed init execute envirment. Try to exectue command directly.", host)
	}

	upgrade_cmd := `install activate patch flash:/S6800-CMW710-SYSTEM-R2609H17.bin all`
	output := ""
	o, err := device.ExecCommandExpect(upgrade_cmd, "[Y/N]", time.Second*5)
	if err != nil {
		log.Printf("[%s]%v\n", host, err)
		return output + o, false
	}
	output += o
	device.ExecCommandExpect("Y", "assword:", time.Second*5)
	if err != nil {
		log.Printf("[%s]%v\n", host, err)
		return output + o, false
	}
	output += o
	o, err = device.ExecCommandTiming("Y", time.Second*120)
	if err != nil {
		log.Printf("[%s]%v\n", host, err)
		return output + o, false
	}
	o, err = device.ExecCommandTiming("install commit", time.Second*120)
	if err != nil {
		log.Printf("[%s]%v\n", host, err)
		return output + o, false
	}
	output += o
	return output, false
}

func readlines(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(file)
	var lines []string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		lines = append(lines, strings.TrimSpace(line))
	}

	return lines, nil
}

func writefile(file, conntent string) error {
	return ioutil.WriteFile(file, []byte(conntent), 0666)
}

func main() {

	sshoptions := nwssh.SSHOptions{
		IgnorHostKey: true,
		BannerCallback: func(msg string) error {
			return nil
		},
		TermType:     "vt100",
		TermHeight:   560,
		TermWidht:    480,
		ReadWaitTime: time.Duration(500) * time.Millisecond, //Read data from a ssh channel timeout
	}

	var err error

	hostfile := "./hosts"
	hosts, err := readlines(hostfile)
	if err != nil {
		log.Fatal("%v", err)
	}

	maxThread := 500
	threadchan := make(chan struct{}, maxThread)

	wait := sync.WaitGroup{}

	for _, host := range hosts {
		wait.Add(1)
		go func(host string) {
			threadchan <- struct{}{}
			output, _ := run(host, "22", sshoptions)
			writefile("./out/"+host, output)
			<-threadchan
			wait.Done()
		}(host)
	}
	wait.Wait()

}

package test

import (
	"bufio"
	"io/ioutil"
	"log"
	"nwssh"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

func run(host, port string, sshoptions nwssh.SSHOptions, cmds []string) bool {

	var devssh *nwssh.SSHBase
	var err error

	username := ""
	password := ""

	devssh, err = nwssh.SSH(host, port, username, password, time.Duration(10)*time.Second, sshoptions)
	defer devssh.Close()
	if err != nil {
		log.Printf("[%s]%v\n", host, err)
		return false
	}
	if err = devssh.Connect(); err != nil {
		log.Printf("[%s]%v\n", host, err)
		return false
	}

	var device nwssh.SSHBASE

	device = &nwssh.H3cSSH{devssh}

	key_regrex := regexp.MustCompile("MMU ERR")

	if !device.SessionPreparation() {
		log.Printf("[%s]Failed init execute envirment. Try to exectue command directly.", host)
	}
	for _, cmd := range cmds {
		o, err := device.ExecCommand(cmd)
		if err != nil {
			log.Printf("[%s]Failed exec cmd '%s'. Error: %v", host, cmd, err)
			log.Printf("[%s]Exit execution!\n", host)
			break
		}
		if key_regrex.MatchString(o) {
			return true
		}
		time.Sleep(time.Second * time.Duration(1))
	}

	return false

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

	var cmds []string
	var err error

	hostfile := "/root/s6800hosts"
	hosts, err := readlines(hostfile)
	if err != nil {
		log.Fatal("%v", err)
	}

	for {
		maxThread := 500
		threadchan := make(chan struct{}, maxThread)
		result := make([]string, 2, 1)
		wait := sync.WaitGroup{}

		for _, host := range hosts {
			wait.Add(1)
			go func(host string) {
				threadchan <- struct{}{}
				if run(host, "22", sshoptions, cmds) {
					result = append(result, host)
				}
				<-threadchan
				wait.Done()
			}(host)
		}
		wait.Wait()

		if result != nil {
			var output string
			for _, v := range result {
				output += v + "\n"
			}
			writefile("/tmp/h3ceccoutput", output)
		}
		time.Sleep(time.Duration(24) * time.Hour)
	}
}

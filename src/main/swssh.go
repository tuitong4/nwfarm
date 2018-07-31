package main

import (
	"bufio"
	"debug/elf"
	"flag"
	"fmt"
	"log"
	"nwssh"
	"os"
	"strings"
	"sync"
	"time"
)

type Args struct {
	hostfile     string
	host         string
	cmd          string
	cmdprefix    string
	swvendor     string
	username     string
	password     string
	port         string
	saveconfig   bool
	strictmode   bool
	timeout      int
	readwaittime int
	cmdtimeout   int
	cmdinterval  int
	logdir       string
	conffiledir  string
	transcation  string
	privatekey   string
}

var args = Args{}

func initflag() {
	flag.StringVar(&args.hostfile, "f", "", `Target hosts list file, host format is :
	'10.10.10.10'
	'12.12.12.12'.`)

	flag.StringVar(&args.host, "host", "", "Target hostname, format is '10.10.10.10'.")

	flag.StringVar(&args.cmdprefix, "cmd_prefix", "", `A prefix of command list file. For example:
	test.cmd.cisco
	test.cmd.nexus
	test.cmd.h3c
	test.cmd.huawei
	test.cmd.ruijie
	'test' is the prefix (cmd_prefix).`)

	flag.StringVar(&args.cmd, "cmd", "", "Command(s) that's sended to remote terminal, if number of command is greater than one, "+
		"use the ';' as dilimiter.")

	flag.StringVar(&args.swvendor, "V", "", "The vendor of target host, if not spicified, script will detect vendor autoly.")

	flag.StringVar(&args.username, "u", "", "Username to login.")

	flag.StringVar(&args.password, "p", "", "Password of user.")

	flag.StringVar(&args.port, "port", "22", "Target TCP port to connect.")

	flag.BoolVar(&args.saveconfig, "save", false, "Auto save running-config after finished execute command.")

	flag.BoolVar(&args.strictmode, "strict", false, `Use strict mode to execute command, that means it expects the host prompt to
confirm command execute succsess until timeout reached.This is not recommand when you're requried enter 'Y/N' to confirm info.`)

	flag.IntVar(&args.timeout, "timeout", 10, "SSH connection timeout in seconds.")

	flag.IntVar(&args.readwaittime, "readwaittime", 500, "Time of wait ssh channel return the remote respone, if time reached,"+
		" return the respone.")

	flag.StringVar(&args.logdir, "logpath", "", "Log command output to /<path>/<ip_addr> instead of stdout.")

	flag.StringVar(&args.conffiledir, "confpath", "", `Configuration file path, configuration file should be like '/path/10.10.10.10',
filename will used as target hostname.`)

	flag.StringVar(&args.transcation, "tran", "", `Run a defined transcation such as get ifconifg, bgp neighbors .etc.`)

	flag.StringVar(&args.privatekey, "pkey", "", `Private key used to login on devie.`)

	flag.StringVar(&args.cmdtimeout, "cmdtimeout", 10, `Waiting time when exec command expect prompt, if timeout reached, 
return the respone and ignore the prompt.`)

	flag.StringVar(&args.cmdinterval, "cmdinterval", 2, `The interval between execution of two commands.`)

	flag.Parse()

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

var BannerVendorKeys map[string]string = map[string]string{
	"H3C":    "h3c",
	"HUAWEI": "huawei",
	"NEXUS":  "nexus",
	"CISCO":  "cisco",
	"RUIJIE": "ruijie",
}

func guessVendorByBanner(banner string) string {
	s := strings.ToLower(banner)
	for k, v := range BannerVendorKeys {
		if strings.Contains(s, v) {
			return k
		}
	}
	return ""
}

var WelecomInfoVendorKeys map[string]string = map[string]string{
	"H3C":    "h3c",
	"HUAWEI": "info",
	"NEXUS":  "nexus",
	"CISCO":  "user",
	"RUIJIE": "ruijie", //RUIJIE is not support welecominfo default, so we still use "ruijie" as the key even it takes no effect.
}

func guessVendorByWelecomInfo(welecominfo string) string {
	s := strings.ToLower(welecominfo)
	for k, v := range WelecomInfoVendorKeys {
		if strings.Contains(s, v) {
			return k
		}
	}
	return ""
}

var VersionInfoVendorKeys map[string]string = map[string]string{
	"H3C":    "H3C",
	"HUAWEI": "HUAWEI",
	"NEXUS":  "Nexus",
	"CISCO":  "Cisco IOS",
	"RUIJIE": "Ruijie",
}

func guessVendorByVesionInfo(versioninfo string) string {
	for k, v := range WelecomInfoVendorKeys {
		if strings.Contains(versioninfo, v) {
			return k
		}
	}
	return ""
}

func guessVendor(s *nwssh.SSHBase, banner string) string {
	if vendor := guessVendorByWelecomInfo(s.WelecomInfo); vendor == "" {
		if vendor = guessVendorByBanner(banner); vendor == "" {

			resp, _ := s.ExecCommand(`show version | in Ruij`)
			if strings.Contains(resp, "Ruijie") {
				return "RUIJIE"
			}

			resp, _ = s.ExecCommand(`show version | in Software`)
			if strings.Contains(resp, "Nexus") {
				return "NEXUS"
			}
			if strings.Contains(resp, "Cisco") {
				return "CISCO"
			}

			resp, _ = s.ExecCommand(`display version | in Copyright`)
			rsp := strings.ToLower(resp)
			if strings.Contains(rsp, "h3c") {
				return "H3C"
			}
			if strings.Contains(rsp, "huawei") {
				return "HUAWEI"
			}

		}
		return vendor
	}
}

func run(host, port string, sshoptions nwssh.SSHOptions, cmds []string, args *Args) {
	var banner string
	var devssh nwssh.SSHBase
	vendor := *args.swvendor
	if vendor == "" {
		sshoptions.BannerCallback = func(message string) error {
			banner = message
			return nil
		}
	}

	devssh = nwssh.SSH{host, port, *args.username, *args.password, *args.timeout * time.Second, sshoptions}

	if err := devssh.Connect(); err != nil {
		log.Printf("[%s]%v\n", host, err)
		return
	}

	if vendor == "" {
		vendor = guessVendor(&devssh, banner)
		if vendor == "" {
			log.Printf("[%s]Failed prasie device's vendor automatically.\n", host)
			return
		}
	}

	var device nwssh.SSHBASE
	if vendor == "H3C" {
		device = nwssh.H3cSSH{&devssh}
	} else if vendor == "HUAWEI" {
		device = nwssh.HuaweiSSH{&devssh}
	} else if vendor == "NEXUS" {
		device = nwssh.NexusSSH{&devssh}
	} else if vendor == "CISCO" {
		device = nwssh.CiscoSSH{&devssh}
	} else if vendor == "RUIJIE" {
		device = nwssh.RuijieSSH{&devssh}
	}

	var output string
	var mutex sync.Mutex
	if *args.strictmode {
		for _, cmd := range cmds {
			o, err := device.ExecCommandExpectPrompt(cmd, time.Second*args.cmdtimeout)

			if err != nil {
				log.Printf("[%s]Failed exec cmd '%s'. Error: %v", host, cmd, err)
				output += o
				if *args.logdir != "" {
					writefile(*args.logdir+host, output)
				} else {
					mutex.Lock()
					os.Stdin.Write([]byte(output))
					mutex.Unlock()
				}
				return
			}
		}
		if *args.logdir != "" {
			writefile(*args.logdir+host, output)
		} else {
			mutex.Lock()
			os.Stdin.Write([]byte(output))
			mutex.Unlock()
		}
		return
	}

	if !*args.strictmode {
		for _, cmd := range cmds {
			o, err := device.ExecCommand(cmd)
			if err != nil {
				log.Printf("[%s]Failed exec cmd '%s'. Error: %v", host, cmd, err)
				output += o
				if *args.logdir != "" {
					writefile(*args.logdir+host, output)
				} else {
					mutex.Lock()
					os.Stdin.Write([]byte(output))
					mutex.Unlock()
				}
				return
			}

			time.Sleep(time.Second * *args.cmdinterval)
		}

		if *args.logdir != "" {
			writefile(*args.logdir+host, output)
		} else {
			mutex.Lock()
			os.Stdin.Write([]byte(output))
			mutex.Unlock()
		}
		return
	}

}

func main() {
	fmt.Println(readhostfile("hosts"))

}

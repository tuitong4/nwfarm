package main

import (
	"flag"
	"os"
	"bufio"
	"strings"
	"fmt"
	"nwssh"
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
	logdir       string
	conffiledir  string
	transcation	 string
	privatekey	 string
}

var args = Args{}

func initflag() {
	flag.StringVar(&args.hostfile,"f", "", `Target hosts list file, host format is :
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

	flag.StringVar(&args.cmd, "cmd", "", "Command(s) that's sended to remote terminal, if number of command is greater than one, " +
		"use the ';' as dilimiter.")

	flag.StringVar(&args.swvendor, "V", "", "The vendor of target host, if not spicified, script will detect vendor autoly.")

	flag.StringVar(&args.username, "u", "", "Username to login.")

	flag.StringVar(&args.password, "p", "", "Password of user.")

	flag.StringVar(&args.port, "port", "22", "Target TCP port to connect.")

	flag.BoolVar(&args.saveconfig,"save", false, "Auto save running-config after finished execute command.")

	flag.BoolVar(&args.strictmode, "strict", false, `Use strict mode to execute command, that means it expects the host prompt to
confirm command execute succsess until timeout reached.This is not recommand when you're requried enter 'Y/N' to confirm info.`)

	flag.IntVar(&args.timeout, "timeout", 10, "SSH connection timeout in seconds.")

	flag.IntVar(&args.readwaittime, "readwaittime", 500, "Time of wait ssh channel return the remote respone, if time reached," +
		" return the respone.")

	flag.StringVar(&args.logdir, "logpath", "", "Log command output to /<path>/<ip_addr> instead of stdout.")

	flag.StringVar(&args.conffiledir, "confpath", "", `Configuration file path, configuration file should be like '/path/10.10.10.10',
filename will used as target hostname.`)

	flag.StringVar(&args.transcation, "tran", "", `Run a defined transcation such as get ifconifg, bgp neighbors .etc.`)

	flag.StringVar(&args.privatekey, "pkey", "", `Private key used to login on devie.`)

	flag.Parse()

}

func readlines(filename string) ([]string, error){
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(file)
	var lines []string
	for {
		line, err := reader.ReadString('\n')
		if err != nil{
			break
		}
		lines = append(lines, strings.TrimSpace(line))
	}

	return lines, nil
}

var BannerVendorKeys map[string] string = map[string] string{
	"H3C" : "h3c",
	"HUAWEI" : "huawei",
	"NEXUS"	: "nexus",
	"CISCO" : "cisco",
	"RUIJIE": "ruijie",
}

func guessVendorByBanner(banner string) string{
	s := strings.ToLower(banner)
	for k, v := range BannerVendorKeys{
		if strings.Contains(s, v){
			return k
		}
	}

	return ""
}

func run(host, port string, sshoptions nwssh.SSHOptions, *args Args){
	if *args.swvendor != "" {
		var swvendor string
		sshoptions.BannerCallback = func(message string) error {
			swvendor = message
			return nil
		}
	}

}


func main() {
	fmt.Println(readhostfile("hosts"))
}

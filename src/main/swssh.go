package main

import (
	"flag"
	"os/exec"
)

type args struct {
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
}

func initflag() {
	hostfile := flag.String("f", "", `Target hosts list file, host format is :
	'10.10.10.10'
	'12.12.12.12'.`)

	host := flag.String("host", "", "Target hostname, format is '10.10.10.10'.")

	cmdprefix := flag.String("cmd_prefix", "", `A prefix of command list file. For example:
	test.cmd.cisco
	test.cmd.nexus
	test.cmd.h3c
	test.cmd.huawei
	test.cmd.ruijie
	'test' is the prefix (cmd_prefix).`)

	cmd := flag.String("cmd", "", "Command(s) that's sended to remote terminal, if number of command is greater than one, use the ';' as dilimiter.")

	swvendor := flag.String("V", "", "The vendor of target host, if not spicified, script will auto detect vendor.")

	username := flag.String("u", "", "Username to login.")

	password := flag.String("p", "", "Password to login.")

	port := flag.String("port", "22", "Target TCP port to connect.")

	saveconfig := flag.Bool("save", false, "Auto save running-config after finished execute command.")

	strictmode := flag.Bool("strict", false, `Use strict mode to execute command, that means it expects the host prompt to confirm command execute succsess until timeout reached. This is not recommand when you're requried enter 'Y/N' to confirm info.`)

	timeout := flag.Int("timeout", 10, "SSH connection timeout in seconds.")

	readwaittime := flag.Int("readwaittime", 500, "Time of wait ssh channel return the remote respone, if time reached, return the respone.")

	logdir := flag.String("log_path", "", "Log command output to /<path>/<ip_addr> instead of stdout.")

	conffiledir := flag.String("conf_path", "", `Configuration file path, configuration file should be like '/path/10.10.10.10', filename will used as target hostname.`)

	flag.Parse()

}

func main() {
	initflag()

}

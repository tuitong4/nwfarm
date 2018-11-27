package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
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
	prettyoutput bool
	help         bool
	nopage       bool
}

var args = Args{}

func initflag() {
	flag.StringVar(&args.hostfile, "f", "", `Target hosts list file, one ip on a separate line, for example:
'10.10.10.10'
'12.12.12.12'.`)
	flag.StringVar(&args.host, "host", "", `Target host, ip address is expect. Multiple hosts are supported 
when used ';' as dilimiter.`)
	flag.StringVar(&args.cmdprefix, "cmd_prefix", "", `A prefix of command list file. 
For example:
	test.cmd.cisco
	test.cmd.nexus
	test.cmd.h3c
	test.cmd.huawei
	test.cmd.ruijie 
'test' is the prefix.`)
	flag.StringVar(&args.cmd, "cmd", "", `The command(s) to be executed remotely. Multiple commands are 
supported when used ';' as dilimiter.`)
	flag.StringVar(&args.swvendor, "V", "", `Vendor of target host, if not spicified, it will be checked 
automatically.`)
	flag.StringVar(&args.username, "u", "", "Username for login.")
	flag.StringVar(&args.password, "p", "", "Password for login.")
	flag.StringVar(&args.port, "port", "22", "Target SSH port to connect.")
	flag.BoolVar(&args.saveconfig, "save", false, "Automatically save running-config after finished execution.")
	flag.BoolVar(&args.strictmode, "strict", false, `Execute command using strict mode, host's prompt is expected to 
confirm command was successfully executed until timeout reached.
This is not recommand when it's requried enter 'Y/N' to confirm 
execution.`)
	flag.IntVar(&args.timeout, "timeout", 10, "SSH connection timeout in seconds.")
	flag.IntVar(&args.readwaittime, "readwaittime", 500, `The time to wait ssh channel return the respone, if time reached,
end the wait. In Millisecond.`)
	flag.StringVar(&args.logdir, "logpath", "", "Log command output to /<path>/<ip_addr> instead of stdout.")
	flag.StringVar(&args.conffiledir, "confpath", "", `Configuration file path, filename will be used as target hostname.`)
	flag.StringVar(&args.transcation, "tran", "", `Run a defined transcation such as get 'ifconifg', 'bgpneighbors'.`)
	flag.StringVar(&args.privatekey, "pkey", "", `Private key used for login, if spicified, it'll ignore password.`)
	flag.IntVar(&args.cmdtimeout, "cmdtimeout", 10, `The wait time for executing commands remotely, if timeout reached, 
means execution is failed.`)
	flag.IntVar(&args.cmdinterval, "cmdinterval", 2, `The interval of sending command to remotely host.`)
	flag.BoolVar(&args.prettyoutput, "pretty", false, `Strip the command line and device prompt of the respone output.`)
	flag.BoolVar(&args.help, "help", false, `Usage of CLI.`)
	flag.BoolVar(&args.nopage, "nopage", true, `Disable enter "SPACE" to show more output line.`)

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
	if s.WelecomInfo == "" {
		time.Sleep(time.Second * 1)
	}
	vendor := guessVendorByWelecomInfo(s.WelecomInfo)
	if vendor == "" {
		vendor = guessVendorByBanner(banner)
		if vendor == "" {
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
	}
	return vendor
}

func writefile(file, conntent string) error {
	return ioutil.WriteFile(file, []byte(conntent), 0666)
}

func run(host, port string, sshoptions nwssh.SSHOptions, cmds []string, args *Args, basiscmd map[string][]string) {
	var banner string
	var devssh *nwssh.SSHBase
	var err error
	vendor := strings.ToUpper(args.swvendor)
	if vendor == "" {
		sshoptions.BannerCallback = func(message string) error {
			banner = message
			return nil
		}
	}

	devssh, err = nwssh.SSH(host, port, args.username, args.password, time.Duration(args.timeout)*time.Second, sshoptions)
	defer devssh.Close()
	if err != nil {
		log.Printf("[%s]%v\n", host, err)
		return
	}
	if err = devssh.Connect(); err != nil {
		log.Printf("[%s]%v\n", host, err)
		return
	}

	if vendor == "" {
		vendor = guessVendor(devssh, banner)
		if vendor == "" {
			log.Printf("[%s]Failed parse device's vendor automatically.\n", host)
			return
		}
	}

	var device nwssh.SSHBASE
	if vendor == "H3C" {
		device = &nwssh.H3cSSH{devssh}
	} else if vendor == "HUAWEI" {
		device = &nwssh.HuaweiSSH{devssh}
	} else if vendor == "NEXUS" {
		device = &nwssh.NexusSSH{devssh}
	} else if vendor == "CISCO" {
		device = &nwssh.CiscoSSH{devssh}
	} else if vendor == "RUIJIE" {
		device = &nwssh.RuijieSSH{devssh}
	}

	if len(cmds) == 0 {
		cmds = basiscmd[vendor]
	}

	var output string
	var mutex sync.Mutex
	if args.strictmode && len(cmds) > 0 {
		if args.nopage && !device.SessionPreparation() {
			log.Printf("[%s]Failed init execute envirment. Try to exectue command directly.", host)
		}
		for _, cmd := range cmds {
			o, err := device.ExecCommandExpectPrompt(cmd, time.Second*time.Duration(args.cmdtimeout))
			if args.prettyoutput == true {
				o = nwssh.SanitizeRespone(o, true, true)
			}
			output += o
			if err != nil {
				log.Printf("[%s]Failed exec cmd '%s'. Error: %v", host, cmd, err)
				log.Printf("[%s]Exit execution!\n", host)
				break
			}
		}
	}

	if !args.strictmode && len(cmds) > 0 {
		if args.nopage && !device.SessionPreparation() {
			log.Printf("[%s]Failed init execute envirment. Try to exectue command directly.", host)
		}
		for _, cmd := range cmds {
			o, err := device.ExecCommand(cmd)
			if args.prettyoutput {
				o = nwssh.SanitizeRespone(o, true, true)
			}
			output += o
			if err != nil {
				log.Printf("[%s]Failed exec cmd '%s'. Error: %v", host, cmd, err)
				log.Printf("[%s]Exit execution!\n", host)
				break
			}
			time.Sleep(time.Second * time.Duration(args.cmdinterval))
		}
	}

	if args.transcation != "" {
		if args.nopage && !device.SessionPreparation() {
			log.Printf("[%s]Failed init executable envirment. Try to exectue commands directly.", host)
		}
		output, err = device.RunTranscation(args.transcation)
		if err != nil {
			log.Printf("[%s]Failed exec transcation '%s'. Error: %v", host, args.transcation, err)
		}
	}

	if args.saveconfig {
		if !device.SaveRuningConfig() {
			log.Printf("[%s]Failed save configuration.", host)
		}
	}

	if args.logdir != "" {
		writefile(args.logdir+host, output)
	} else {
		mutex.Lock()
		os.Stdout.Write([]byte(output + "\n"))
		mutex.Unlock()
	}

	log.Printf("[%s]Finished execution!\n", host)
}

func main() {
	initflag()
	if args.help {
		fmt.Println("Usage of CLI:")
		flag.PrintDefaults()
		os.Exit(0)
	}
	if args.host == "" && args.hostfile == "" && args.conffiledir == "" {
		fmt.Println("Traget host is expected but got none. See help docs.")
		os.Exit(0)
	}

	if args.username == "" {
		fmt.Println("Username is expected but got none. See help docs.")
		os.Exit(0)
	}

	if args.password == "" {
		fmt.Printf("Please input the password:")
		fmt.Scanf("%s", &args.password)
	}

	sshoptions := nwssh.SSHOptions{
		PrivateKeyFile: args.privatekey,
		IgnorHostKey:   true,
		BannerCallback: func(msg string) error {
			return nil
		},
		TermType:     "vt100",
		TermHeight:   560,
		TermWidht:    480,
		ReadWaitTime: time.Duration(args.readwaittime) * time.Millisecond, //Read data from a ssh channel timeout
	}

	var cmds []string
	var err error
	basiscmd := make(map[string][]string)

	maxThread := 500
	threadchan := make(chan struct{}, maxThread)
	wait := sync.WaitGroup{}
	if args.conffiledir != "" {
		fileinfo, err := ioutil.ReadDir(args.conffiledir)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
		for _, f := range fileinfo {
			host := f.Name()

			if !f.IsDir() && len(strings.Split(host, ".")) == 4 {
				cmds, err = readlines(args.conffiledir + host)
				if err != nil {
					fmt.Println("Fialed to read command from configuration file.")
				}
				wait.Add(1)
				go func(host string, cmds []string) {
					threadchan <- struct{}{}
					run(host, args.port, sshoptions, cmds, &args, basiscmd)
					<-threadchan
					wait.Done()
				}(host, cmds)
			}
		}
	}

	var hosts []string
	if args.host != "" {
		hosts = strings.Split(args.host, ";")

	} else if args.hostfile != "" {
		hosts, err = readlines(args.hostfile)
		if err != nil {
			log.Fatal("%v", err)
		}
	}

	if args.cmd != "" {
		cmds = strings.Split(args.cmd, ";")
	} else if args.cmdprefix != "" {
		basiscmd = map[string][]string{
			"H3C":    nil,
			"HUAWEI": nil,
			"NEXUS":  nil,
			"CISCO":  nil,
			"RUIJIE": nil,
		}
		for k, _ := range basiscmd {
			cmds_t, err := readlines(args.cmdprefix + ".cmd." + strings.ToLower(k))
			if err == nil {
				basiscmd[k] = cmds_t
			}
		}
	}

	for _, host := range hosts {
		wait.Add(1)
		go func(host string) {
			threadchan <- struct{}{}
			run(host, args.port, sshoptions, cmds, &args, basiscmd)
			<-threadchan
			wait.Done()
		}(host)
	}
	wait.Wait()
}

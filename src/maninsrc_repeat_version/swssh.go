package main

import (
	"bufio"
	"encoding/csv"
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
	hostfile       string
	host           string
	cmd            string
	cmdprefix      string
	swvendor       string
	username       string
	password       string
	port           string
	saveconfig     bool
	strictmode     bool
	timeout        int
	readwaittime   int
	cmdtimeout     int
	cmdinterval    int
	repeatinterval int
	repeatduration int
	logdir         string
	conffiledir    string
	cmdfile        string
	csvfile        string
	transcation    string
	privatekey     string
	prettyoutput   bool
	help           bool
	nopage         bool
	repeat         bool
}

var args = Args{}

func initflag() {
	flag.StringVar(&args.hostfile, "f", "", `Read list of targets from a file, for example:
'10.10.10.10'
'12.12.12.12'.`)
	flag.StringVar(&args.host, "host", "", `Single target address or multiple targets separated by ';'.`)
	flag.StringVar(&args.cmdprefix, "cmd_prefix", "", `A prefix of command list file. 
For example:
	test.cmd.cisco
	test.cmd.nexus
	test.cmd.h3c
	test.cmd.huawei
	test.cmd.ruijie 
'test' is the prefix.`)
	flag.StringVar(&args.cmd, "cmd", "", `Command to be executed remotely. Multiple commands are 
separated by ';'.`)
	flag.StringVar(&args.swvendor, "V", "", `Vendor of target host, if not spicified, it will be detected 
automatically.`)
	flag.StringVar(&args.username, "u", "", "Username for login.")
	flag.StringVar(&args.password, "p", "", "Password for login.")
	flag.StringVar(&args.port, "port", "22", "Port to connect to on the remote host.")
	flag.BoolVar(&args.saveconfig, "save", false, "Automatically save running-config after execution completed.")
	flag.BoolVar(&args.strictmode, "strict", false, `Use strict mode, when enabled, host prompt is expected to 
confirm if command was successfully executed until timeout reached.
This is not recommand when it's requried enter 'Y/N' to confirm 
execution.`)
	flag.IntVar(&args.timeout, "timeout", 10, "SSH connection timeout(in seconds).")
	flag.IntVar(&args.readwaittime, "readwaittime", 500, `The time to wait ssh channel return the respone, if readwaittime 
reached, stop waiting, return received data. In Millisecond.`)
	flag.StringVar(&args.logdir, "logpath", "", "Log command output to /<path>/<ip_addr> instead of stdout.")
	flag.StringVar(&args.conffiledir, "confpath", "", `Configuration file path, the filename will be used as target hostname.`)
	flag.StringVar(&args.cmdfile, "cmdfile", "", `Read commands for a file, one command per line.`)
	flag.StringVar(&args.transcation, "tran", "", `Run a defined transcation such as get 'ifconifg', 'bgpneighbors'.`)
	flag.StringVar(&args.privatekey, "pkey", "", `Private key used for login, if spicified, password will be ignored.`)
	flag.IntVar(&args.cmdtimeout, "cmdtimeout", 10, `The time of waiting for the command to finish executing, if timeout 
reached, means execution is failed.`)
	flag.IntVar(&args.cmdinterval, "cmdinterval", 2, `The interval of sending command to remote host.`)
	flag.BoolVar(&args.prettyoutput, "pretty", false, `Strip command line and device prompt from the respone string.`)
	flag.BoolVar(&args.help, "help", false, `Usage of CLI.`)
	flag.BoolVar(&args.nopage, "nopage", true, `Disable enter "SPACE" to show more output lines.`)
	flag.StringVar(&args.csvfile, "csvfile", "", `Read targets from a csv file. Each record consists of host, vendor, 
username, password separated by comma. It's recommanded to use 'cmd_prefix'
to specify executable commands. Note that csv file has no title line.`)
	flag.BoolVar(&args.repeat, "repeat", false, `Execute the commands repeatedly at the given 'repeatinterval',
it will end when the duration reached at 'repeatduration'.`)
	flag.IntVar(&args.repeatinterval, "repeatinterval", 60, `Interval between commands executions(in seconds), 
it's different to the 'cmdinterval'.`)
	flag.IntVar(&args.repeatduration, "repeatduration", 0, `Duration of the repeatedly executions(in seconds), 
0 means permanently. (default 0)`)
	flag.Parse()
}

func readlines(filename string) ([]string, error) {
	file, err := os.Open(filename)
	defer file.Close()
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

func readcsv(filename string) ([][]string, error) {
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return nil, err
	}
	reader := csv.NewReader(file)
	var records [][]string
	for {
		record, err := reader.Read()
		if err != nil {
			break
		}
		var _records []string
		for _, _item := range record {
			_records = append(_records, strings.TrimSpace(_item))
		}
		records = append(records, _records)
	}

	return records, nil
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

func pathIsExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func createPath(path string) error {
	path = strings.TrimRight(path, "/")

	paths := strings.Split(path, "/")
	if len(paths) == 0 {
		return fmt.Errorf("Invalied file path '%s'.", path)
	}

	pathToCreate := ""
	for _, dir := range paths[0:len(paths)] {
		pathToCreate += dir + "/"
		if !pathIsExist(pathToCreate) {
			if err := os.Mkdir(pathToCreate, 0777); err != nil {
				return err
			}
			if err := os.Chmod(pathToCreate, 0777); err != nil {
				return err
			}
		}
	}
	return nil
}

func writefile(file, conntent string) error {
	return os.WriteFile(file, []byte(conntent), 0666)
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
			log.Printf("[%s]Failed to parse device's vendor automatically.\n", host)
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
			log.Printf("[%s]Failed to init execute envirment. Try to execute command directly.", host)
		}
		for _, cmd := range cmds {
			o, err := device.ExecCommandExpectPrompt(cmd, time.Second*time.Duration(args.cmdtimeout))
			if args.prettyoutput == true {
				o = nwssh.SanitizeRespone(o, true, true)
			}
			output += o
			if err != nil {
				log.Printf("[%s]Failed to exec cmd '%s'. Error: %v.\n", host, cmd, err)
				log.Printf("[%s]Exit execution!\n", host)
				break
			}
		}
	}

	if !args.strictmode && len(cmds) > 0 {
		if args.nopage && !device.SessionPreparation() {
			log.Printf("[%s]Failed to init execute envirment. Try to execute command directly.\n", host)
		}
		for _, cmd := range cmds {
			o, err := device.ExecCommand(cmd)
			if args.prettyoutput {
				o = nwssh.SanitizeRespone(o, true, true)
			}
			output += o
			if err != nil {
				log.Printf("[%s]Failed to exec cmd '%s'. Error: %v\n", host, cmd, err)
				log.Printf("[%s]Exit execution!\n", host)
				break
			}
			time.Sleep(time.Second * time.Duration(args.cmdinterval))
		}
	}

	if args.transcation != "" {
		if args.nopage && !device.SessionPreparation() {
			log.Printf("[%s]Failed to init executable envirment. Try to execute commands directly.\n", host)
		}
		output, err = device.RunTranscation(args.transcation)
		if err != nil {
			log.Printf("[%s]Failed exec transcation '%s'. Error: %v\n", host, args.transcation, err)
		}
	}

	if args.saveconfig {
		if !device.SaveRuningConfig() {
			log.Printf("[%s]Failed save configuration.\n", host)
		}
	}

	if args.logdir != "" {
		writefile(args.logdir+host, output)
	} else {
		mutex.Lock()
		os.Stdout.Write([]byte(output + "\n"))
		mutex.Unlock()
	}

	log.Printf("[%s]Execution completed!\n", host)
}

func runRepeatedly(host, port string, sshoptions nwssh.SSHOptions, cmds []string, args *Args, basiscmd map[string][]string) {
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
			log.Printf("[%s]Failed to parse device's vendor automatically.\n", host)
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
	var outputFile *os.File
	var startTime time.Time

	duration := time.Duration(args.repeatduration) * time.Second

	if args.logdir != "" {
		outputFile, err = os.OpenFile(args.logdir+host, os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Printf("[%s]Failed to create the output file '%s', task stopped.", host, args.logdir+host)
			return
		}
		defer outputFile.Close()
	}

	if args.nopage && !device.SessionPreparation() {
		log.Printf("[%s]Failed to init executable envirment. Try to execute commands directly.\n", host)
	}

	startTime = time.Now()
REPEAT:
	if args.strictmode && len(cmds) > 0 {
		for _, cmd := range cmds {
			o, err := device.ExecCommandExpectPrompt(cmd, time.Second*time.Duration(args.cmdtimeout))
			if args.prettyoutput == true {
				o = nwssh.SanitizeRespone(o, true, true)
			}
			output += o
			if err != nil {
				log.Printf("[%s]Failed to exec cmd '%s'. Error: %v.\n", host, cmd, err)
				log.Printf("[%s]Exit execution!\n", host)
				break
			}
		}
	}

	if !args.strictmode && len(cmds) > 0 {
		for _, cmd := range cmds {
			o, err := device.ExecCommand(cmd)
			if args.prettyoutput {
				o = nwssh.SanitizeRespone(o, true, true)
			}
			output += o
			if err != nil {
				log.Printf("[%s]Failed to exec cmd '%s'. Error: %v\n", host, cmd, err)
				log.Printf("[%s]Exit execution!\n", host)
				break
			}
			time.Sleep(time.Second * time.Duration(args.cmdinterval))
		}
	}

	if args.transcation != "" {
		output, err = device.RunTranscation(args.transcation)
		if err != nil {
			log.Printf("[%s]Failed exec transcation '%s'. Error: %v\n", host, args.transcation, err)
		}
	}

	if args.saveconfig {
		if !device.SaveRuningConfig() {
			log.Printf("[%s]Failed save configuration.\n", host)
		}
	}

	if args.logdir != "" {
		outputFile.Write([]byte(output))

	} else {
		mutex.Lock()
		os.Stdout.Write([]byte(output + "\n"))
		mutex.Unlock()
	}

	if args.repeatduration == 0 || time.Now().Sub(startTime) < duration {
		output = ""
		time.Sleep(time.Duration(args.repeatinterval) * time.Second)
		goto REPEAT
	}

	log.Printf("[%s]Execution completed!\n", host)
}

func csvModeRunning(args *Args) {

	var cmds []string
	var err error
	basiscmd := make(map[string][]string)

	maxThread := 500
	threadchan := make(chan struct{}, maxThread)
	wait := sync.WaitGroup{}

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
	} else if args.cmdfile != "" {
		cmds, err = readlines(args.cmdfile)
		if err != nil {
			log.Printf("%v\n", err)
			os.Exit(0)
		}
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

	records, err := readcsv(args.csvfile)
	if err != nil {
		fmt.Println("Fialed to read host infomation from csv file.")
	}

	for _, record := range records {
		_args := args
		_args.host = record[0]
		_args.swvendor = record[1]
		_args.username = record[2]
		_args.password = record[3]

		wait.Add(1)
		go func(host string) {
			threadchan <- struct{}{}
			if args.repeat {
				runRepeatedly(host, args.port, sshoptions, cmds, _args, basiscmd)
			} else {
				run(host, args.port, sshoptions, cmds, _args, basiscmd)
			}
			<-threadchan
			wait.Done()
		}(_args.host)

	}
	wait.Wait()
}

func main() {
	initflag()
	if args.help {
		fmt.Println("Usage of CLI: swssh [args]")
		flag.PrintDefaults()
		os.Exit(0)
	}

	if args.csvfile != "" {
		csvModeRunning(&args)
		os.Exit(1)
	}

	if args.host == "" && args.hostfile == "" && args.conffiledir == "" {
		fmt.Println("Traget host is expected but got none. See help docs.")
		flag.PrintDefaults()
		os.Exit(0)
	}

	if args.username == "" {
		fmt.Println("Username is expected but got none. See help docs.")
		flag.PrintDefaults()
		os.Exit(0)
	}

	if args.password == "" {
		fmt.Printf("Please input the password:")
		fmt.Scanf("%s", &args.password)
	}

	var err error
	if args.logdir != "" {
		if !strings.HasSuffix(args.logdir, "/") {
			args.logdir += "/"
		}
		err = createPath(args.logdir)
		if err != nil {
			fmt.Println("Failed to create logpath '%s', error: %v", args.logdir, err)
			os.Exit(0)
		}

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

	basiscmd := make(map[string][]string)

	maxThread := 500
	threadchan := make(chan struct{}, maxThread)
	wait := sync.WaitGroup{}
	if args.conffiledir != "" {
		fileinfo, err := os.ReadDir(args.conffiledir)
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
					if args.repeat {
						runRepeatedly(host, args.port, sshoptions, cmds, &args, basiscmd)
					} else {
						run(host, args.port, sshoptions, cmds, &args, basiscmd)
					}

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
	} else if args.cmdfile != "" {
		cmds, err = readlines(args.cmdfile)
		if err != nil {
			log.Printf("%v\n", err)
			os.Exit(0)
		}
	}

	for _, host := range hosts {
		wait.Add(1)
		go func(host string) {
			threadchan <- struct{}{}
			if args.repeat {
				runRepeatedly(host, args.port, sshoptions, cmds, &args, basiscmd)
			} else {
				run(host, args.port, sshoptions, cmds, &args, basiscmd)
			}
			<-threadchan
			wait.Done()
		}(host)
	}
	wait.Wait()
}

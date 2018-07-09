package main

import (
	"bufio"
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
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
	}
	config.Config.Ciphers = append(config.Config.Ciphers, "aes128-cbc")

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

type SSHBaseInterface interface {
	Connect()
	Close()
	SendCmd()
	SendCmdExpect()
	SendCmdTiminig()
}

type SSHBase struct {
	host           string
	port           string
	sshconfig      ssh.ClientConfig
	sessiontimeout time.Duration
	client         ssh.Client
	session        ssh.Session
	ReturnChar     string
	termheight     int
	termwidth      int
	termtype       string
	alive          bool
	InChannel      io.WriteCloser
	OutChannel     io.Reader
	ErrChannel     io.Reader
	send           chan string
}

type SSHOptions struct {
	PrivateKeyFile string
	DefaultEnter   string // Character(s) to send to correspond to enter key (default: '\n').
	ResponeReturn  string //Character(s) to use in normalized return data to represent enter key (default: '\n').
	SessionTimeout time.Duration
	IgnorHostKey   bool
	BannerCallback ssh.BannerCallback
}

func SSH(host, port, username, password string, timeout time.Duration, sshopts SSHOptions) (*SSHBase, error) {

	config := &ssh.ClientConfig{
		User:    username,
		Timeout: timeout,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		BannerCallback: sshopts.BannerCallback,
	}

	config.Config.Ciphers = append(config.Config.Ciphers, "aes128-cbc")

	if !sshopts.PrivateKeyFile {

		key, err := ioutil.ReadFile(sshopts.PrivateKeyFile)
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			log.Fatalf("Unable to parse private key: %v", err)
			return nil, errors.New("Unable to parse private key.")
		}
		config.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}
	}

	if sshopts.IgnorHostKey {
		//Not recommend for product enveriment.
		config.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	} else {

		config.HostKeyCallback = ssh.FixedHostKey(getHostKey(host))
	}

	return &SSHBase{sshconfig: config, sessiontimeout: sshopts.SessionTimeout}, nil
}

func getHostKey(host string) ssh.PublicKey {
	// parse OpenSSH known_hosts file
	file, err := os.Open(filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts"))
	if err != nil {
		log.Fatal("Unable to load known_hosts file, %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var hostKey ssh.PublicKey
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), " ")
		if len(fields) != 3 {
			continue
		}
		if strings.Contains(fields[0], host) {
			var err error
			hostKey, _, _, _, err = ssh.ParseAuthorizedKey(scanner.Bytes())
			if err != nil {
				log.Fatalf("Error parsing %q: %v", fields[2], err)
			}
			break
		}
	}

	if hostKey == nil {
		log.Fatalf("No hostkey found for %s", host)
	}

	return hostKey
}

func (s *SSHBase) Connect() error {
	if s.alive {
		log.Fatalf("SSH Connection is opened.")
	}

	client, err := ssh.Dial("tcp", s.host+":"+s.port, s.sshconfig)
	if err != nil {
		log.Fatalf("Unable to connect to remote host: %v", err)
		return err
	}
	defer client.Close()

	sess, err := client.NewSession()

	if err != nil {
		log.Fatalf("Failed to create a session: %v", err)
		return err
	}
	defer sess.Close()

	term_mode := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	err := sess.RequestPty(s.termtype, s.termheight, s.termwidth, term_mode)
	if err != nil {
		log.Fatalf("Failed to init a pseudo terminal: %v", err)
		return err
	}

	stdin, err := sess.StdinPipe()
	if err != nil {
		log.Fatalf("Failed to init input stream: %v", err)
		return err
	}

	stdout, err := sess.StdoutPipe()
	if err != nil {
		log.Fatalf("Failed to init output stream: %v", err)
		return err
	}

	stderr, err := sess.StderrPipe()
	if err != nil {
		log.Fatalf("Failed to init error stream: %v", err)
		return err
	}

	err := sess.Shell()
	if err != nil {
		log.Fatalf("Failed to init remote shell: %v", err)
		return err
	}

	s.alive = true
	s.InChannel = stdin
	s.OutChannel = stdout
	s.ErrChannel = stderr
	s.session = sess
	s.client = client

	return nil
}

func (s *SSHBase) Close() {
	if !s.alive {
		return nil
	}

	s.session.Close()
	s.client.Close()
}
func (s *SSHBase) readRespone() (string, error) {

}

func (s *SSHBase) ExecCommand(cmd string) (respone string, errorinfo string, err error){

}

func (s *SSHBase) sendcommand(cmd string) (int, error) {
	return s.InChannel.Write(cmd + s.ReturnChar)
}


func main() {
	LogAndRun("172.16.130.178")

}

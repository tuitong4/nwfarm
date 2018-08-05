package nwssh

import (
	"bufio"
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const MaxBuffer = 1024 * 10

type SSHBASE interface {
	Connect() error
	Close()
	SessionPreparation() bool
	ExecCommand(string) (string, error)
	ExecCommandTiming(string, time.Duration) (string, error)
	ExecCommandExpect(string, string, time.Duration) (string, error)
	ExecCommandExpectPrompt(string, time.Duration) (string, error)
	SaveRuningConfig() bool
	RunTranscation(string) (string, error)
}

type SSHBase struct {
	host         string
	port         string
	sshconfig    *ssh.ClientConfig
	client       *ssh.Client
	session      *ssh.Session
	termheight   int
	termwidth    int
	termtype     string
	alive        bool
	InChannel    io.WriteCloser
	OutChannel   io.Reader
	respchan     chan string
	readwaittime time.Duration
	WelecomInfo  string
}

type SSHOptions struct {
	PrivateKeyFile string
	IgnorHostKey   bool
	BannerCallback ssh.BannerCallback
	TermType       string
	TermHeight     int
	TermWidht      int
	ReadWaitTime   time.Duration //Read data from a ssh channel timeout
}

func SSH(host, port, username, password string, timeout time.Duration, sshopts SSHOptions) (*SSHBase, error) {

	config := &ssh.ClientConfig{
		User:    username,
		Timeout: timeout, //time.duration should be lager than 1 second.
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		BannerCallback: sshopts.BannerCallback,
	}

	config.Config.Ciphers = append(config.Config.Ciphers, "aes128-cbc")
	config.Config.Ciphers = append(config.Config.Ciphers, "aes128-ctr")

	if sshopts.PrivateKeyFile != "" {

		key, err := ioutil.ReadFile(sshopts.PrivateKeyFile)
		if err != nil {
			return nil, fmt.Errorf("Unable to opne file '%s'.%v", sshopts.PrivateKeyFile, err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("Unable to parse private key.%v", err)
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

	ssh_client := &SSHBase{
		host:         host,
		port:         port,
		sshconfig:    config,
		termheight:   sshopts.TermHeight,
		termwidth:    sshopts.TermWidht,
		termtype:     sshopts.TermType,
		readwaittime: sshopts.ReadWaitTime,
	}
	return ssh_client, nil
}

func getHostKey(host string) ssh.PublicKey {
	// parse OpenSSH known_hosts file
	file, err := os.Open(filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts"))
	if err != nil {
		return nil
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
				return nil
			}
			break
		}
	}

	return hostKey
}

func (s *SSHBase) Connect() error {
	if s.alive {
		return errors.New("SSH Connection is opened.")
	}
	client, err := ssh.Dial("tcp", s.host+":"+s.port, s.sshconfig)
	if err != nil {
		return fmt.Errorf("Unable to connect to remote host: %v", err)
	}

	sess, err := client.NewSession()

	if err != nil {
		return fmt.Errorf("Failed to create a session: %v", err)
	}

	s.session = sess

	err = s.invokeShell()
	if err != nil {
		return fmt.Errorf("Failed to create a shell: %v", err)
	}

	s.alive = true
	s.client = client
	s.WelecomInfo = s.readChannel()

	return nil
}

func (s *SSHBase) invokeShell() error {

	term_mode := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	err := s.session.RequestPty(s.termtype, s.termheight, s.termwidth, term_mode)

	if err != nil {
		return fmt.Errorf("Failed to init a pseudo terminal: %v", err)
	}

	stdin, err := s.session.StdinPipe()
	if err != nil {
		return fmt.Errorf("Failed to init input stream: %v", err)
	}

	stdout, err := s.session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("Failed to init output stream: %v", err)
	}

	err = s.session.Shell()
	if err != nil {
		return fmt.Errorf("Failed to init a remote shell: %v", err)
	}

	respchan_t := make(chan string)

	go func() {
		buf := make([]byte, MaxBuffer)
		for {
			n, err := stdout.Read(buf)
			if err != nil {
				return
			}
			respchan_t <- string(buf[:n])
		}
	}()
	s.respchan = respchan_t
	s.InChannel = stdin
	s.OutChannel = stdout

	return nil
}

func (s *SSHBase) Close() {
	if s.alive {
		s.session.Close()
		s.client.Close()
		s.alive = false
	}
}

func (s *SSHBase) readChannel() string {

	respone := ""
	delayed := false
READEND:
	for {
		select {
		case resp := <-s.respchan:
			respone += normalizeLineFeeds(resp)
			delayed = false

		default:
			if delayed {
				break READEND
			}
			time.Sleep(s.readwaittime)
			delayed = true
		}
	}
	return respone
}

func (s *SSHBase) readChannelTiming(timeout time.Duration) string {
	// timeouted read until timeout reached.
	respone := ""
	timer := time.NewTimer(timeout)
READEND:
	for {
		select {
		case resp := <-s.respchan:
			respone += normalizeLineFeeds(resp)

		case <-timer.C:
			break READEND
		}
	}
	return respone
}

func (s *SSHBase) readChannelExpect(expect string, timeout time.Duration) (respone string, err error) {
	// Expect string or break until timeout reached.
	respone = ""
	timer := time.NewTimer(timeout)
READEND:
	for {
		select {
		case resp := <-s.respchan:
			respone += normalizeLineFeeds(resp)
			if strings.Contains(respone, expect) {
				break READEND
			}

		case <-timer.C:
			err = fmt.Errorf("Timed-out reading channel, pattern '%s' not found in output.", expect)
			break READEND
		}
	}
	return
}

func (s *SSHBase) readChannelExpectPrompt(timeout time.Duration) (respone string, err error) {
	// Expect string or break until timeout reached.
	respone = ""
	timer := time.NewTimer(timeout)
READEND:
	for {
		select {
		case resp := <-s.respchan:
			respone += normalizeLineFeeds(resp)
		case <-timer.C:
			err = fmt.Errorf("Timed-out reading channel, prompt not found in output.")
			break READEND
		default:
			lines := strings.Split(strings.TrimSpace(respone), "\n")
			if len_lines := len(lines); len_lines > 1 {
				if findPrompt(lines[len_lines-1]) {
					break READEND
				}
			}
		}
	}
	return respone, err
}

func Normalize(s string) string {
	return strings.TrimSpace(s) + "\n"
}

func normalizeLineFeeds(s string) string {
	//strip all "\r" in line like "\r\r\n", "\r\n", etc.
	return strings.Replace(s, "\r", "", -1)
}

func findPrompt(s string) bool {
	/*Prompt Characters:
		< > a-z A-Z 0-9 ~ * / _ - [] ()
	 Prompt sample:
		BJ_YF_311-F-02_N7718-1#
		<BJ_YF_305-A-15_CE5810>
		<BJ-YZH-101-1109/1111-LVS-S5500>
		<BJ_YF_311_F-12-13_LVS_S5560>
		[~BJ_YF_320-I-10_CE5810]

	NOT SAFE OPERATION!!!!!!!!!!!!
	*/

	r := regexp.MustCompile(`[<\[\w~@_\-\(\)\.\*\/]+(>|%|#|\]|\$)`)
	return r.MatchString(s)
}

func SanitizeRespone(resp string, stripcmd bool, stripprompt bool) string {
	resp = strings.TrimSpace(resp)
	if stripcmd && stripprompt {
		lines := strings.Split(resp, "\n")
		return strings.Join(lines[1:len(lines)-1], "\n")
	}
	if stripcmd {
		lines := strings.SplitN(resp, "\n", 1)
		return strings.Join(lines[1:], "")
	}
	if stripprompt {
		lines := strings.Split(resp, "\n")
		return strings.Join(lines[:len(lines)-1], "\n")
	}

	return resp
}

func (s *SSHBase) clearBuffer() {
READEND:
	for {
		select {
		case <-s.respchan:
		default:
			break READEND
		}
	}
}

func (s *SSHBase) preparateWriting() bool {
	if findPrompt(s.WelecomInfo) {
		return true
	}

	if findPrompt(s.readChannel()) {
		return true
	}

	for i := 0; i < 3; i++ {
		time.Sleep(time.Millisecond * 2)
		resp, err := s.ExecCommand("")
		if err != nil {
			return false
		}
		if findPrompt(resp) {
			return true
		}
	}
	return false
}

func (s *SSHBase) disablePaging(cmd string) bool {
	if _, err := s.ExecCommand(cmd); err != nil {
		return false
	}
	return true
}

func (s *SSHBase) SessionPreparation() bool {

	if !s.preparateWriting() {
		return false
	}

	if !s.disablePaging("") {
		return false
	}

	s.clearBuffer()

	return true
}

func (s *SSHBase) sendCommand(cmd string) (int, error) {

	n, err := s.InChannel.Write([]byte(Normalize(cmd)))
	if err != nil {
		err = fmt.Errorf("Failed to send command `%s` to remote.%v", cmd, err)
	}

	return n, err
}

func (s *SSHBase) ExecCommand(cmd string) (respone string, err error) {
	s.clearBuffer()
	_, err = s.sendCommand(cmd)
	if err != nil {
		return "", err
	}
	respone = s.readChannel()
	err = nil
	return
}

func (s *SSHBase) ExecCommandTiming(cmd string, timeout time.Duration) (respone string, err error) {
	s.clearBuffer()
	_, err = s.sendCommand(cmd)
	if err != nil {
		return "", err
	}
	respone = s.readChannelTiming(timeout)
	return
}

func (s *SSHBase) ExecCommandExpect(cmd string, expect string, timeout time.Duration) (respone string, err error) {
	s.clearBuffer()
	_, err = s.sendCommand(cmd)
	if err != nil {
		return "", err
	}

	respone, err = s.readChannelExpect(expect, timeout)
	return
}

func (s *SSHBase) ExecCommandExpectPrompt(cmd string, timeout time.Duration) (respone string, err error) {
	s.clearBuffer()
	_, err = s.sendCommand(cmd)
	if err != nil {
		return "", err
	}
	respone, err = s.readChannelExpectPrompt(timeout)
	return
}


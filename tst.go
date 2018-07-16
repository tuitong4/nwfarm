package main

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"net"
	"strings"
	"time"
)

type SSHSession struct {
	session *ssh.Session
	in      chan string
	out     chan string
	stdout  io.Reader
	errchan chan string
}

func NewSSHSession(user, password, ipPort string) (*SSHSession, error) {
	sshSession := new(SSHSession)
	if err := sshSession.createConnection(user, password, ipPort); err != nil {
		return nil, err
	}
	if err := sshSession.muxShell(); err != nil {
		return nil, err
	}
	if err := sshSession.start(); err != nil {
		return nil, err
	}
	return sshSession, nil
}

func (this *SSHSession) muxShell() error {
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}
	if err := this.session.RequestPty("vt100", 80, 40, modes); err != nil {
		return err
	}
	w, err := this.session.StdinPipe()
	if err != nil {
		return err
	}
	r, err := this.session.StdoutPipe()
	if err != nil {
		return err
	}

	e, err := this.session.StderrPipe()
	if err != nil {
		return err
	}

	in := make(chan string, 0)
	out := make(chan string, 0)
	errs := make(chan string, 0)

	go func() {
		for cmd := range in {
			w.Write([]byte(cmd + "\n"))
		}
	}()

	go func() {
		var (
			buf [65 * 1024]byte
			t   int
		)
		for {
			n, err := r.Read(buf[t:])
			if err != nil {
				return
			}
			t += n
			out <- string(buf[:t])
			t = 0
		}
	}()

	go func() {
		var (
			buf [65 * 1024]byte
			t   int
		)
		for {
			n, err := e.Read(buf[t:])
			if err != nil {
				return
			}
			t += n
			if n == 0 {
				fmt.Println("Error: read 0 byte!")
				errs <- string("")
			} else {
				fmt.Println("Error: read OK!")
				errs <- string(buf[:t])
			}
			t = 0
		}
	}()

	this.in = in
	this.out = out
	this.stdout = r
	this.errchan = errs
	return nil
}

func (this *SSHSession) start() error {
	if err := this.session.Shell(); err != nil {
		return err
	}
	return nil
}

func (this *SSHSession) createConnection(user, password, ipPort string) error {

	client, err := ssh.Dial("tcp", ipPort, &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Config: ssh.Config{
			Ciphers: []string{"aes128-cbc"},
		},
	})
	if err != nil {
		return err
	}

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	this.session = session
	return nil
}

func (this *SSHSession) WriteChannel(cmds ...string) {
	for _, cmd := range cmds {
		this.in <- cmd
	}
}

func (this *SSHSession) ReadChannelExpect(maxIntervalTime float32, expects ...string) string {
	result := ""
	isDelay := false
ExitLoop:
	for {
		select {
		case sout := <-this.out:
			isDelay = false
			result = result + sout
			for _, expect := range expects {
				if strings.Contains(sout, expect) {
					break ExitLoop
				}
			}
		default:
			//如果已经延迟过了，则直接返回
			if isDelay {
				break ExitLoop
			}
			time.Sleep(time.Duration(maxIntervalTime) * time.Second)
			isDelay = true
		}
	}
	return result
}

func foo() (ok bool, err string) {
	return true, "Error"
}

func main() {
	host := "172.19.1.1"
	port := "22"
	username := ""
	pass := ""

	d, err := NewSSHSession(username, pass, host+":"+port)
	if err != nil {
		fmt.Println(err)
	}

	d.session.Run("show versinon")


	//d.in <- "show version"
	//time.Sleep(time.Second * 2)

	output_t := ""
	timeout := make(chan struct{})
	go func() {
		time.Sleep(time.Duration(1) * time.Second)
		timeout <- struct{}{}
	}()

STDOUT:
	for {
		select {
		case output := <-d.out:
			output_t += output

		case <-timeout:
			break STDOUT
		}

	}

	errtimeout := make(chan struct{})
	go func() {
		time.Sleep(time.Duration(1) * time.Second)
		errtimeout <- struct{}{}
	}()

	erroutput_t := ""
STDERR:
	for {
		select {
		case output := <-d.out:
			erroutput_t += output

		case <-errtimeout:
			break STDERR
		}

	}
	fmt.Println("STDOUT: ", output_t)
	fmt.Println("STDError: ", erroutput_t)


}

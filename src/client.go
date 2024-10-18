package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"ircc/src/builtin"
	"ircc/src/guard"
	"net"
	"net/textproto"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const CHATSIZ = 1 << 13

const (
	Msgtype__Privmsg = iota
	Msgtype__None
)

type Message struct {
	Timestamp time.Time
	Username  string
	Data      string
}

type Mode int

const (
	Mode__Normal Mode = iota
	Mode__Insert
)

type Server struct {
	name   string
	chat   []Message
	reader *textproto.Reader
}

type Client struct {
	exit       bool
	port       int
	address    string
	nickname   string
	realname   string
	server     *Server
	serverList map[string]*Server
	stream     net.Conn
	mode       Mode
}

func NewClient(addr string, port int, nickname string, realname string) Client {
	return Client{
		address:    addr,
		port:       port,
		nickname:   nickname,
		realname:   realname,
		mode:       Mode__Normal,
		serverList: map[string]*Server{},
	}
}

func (self *Client) NewServer(name string, reader *textproto.Reader) *Server {
	return &Server{name: name, reader: reader, chat: []Message{}}
}

func (self *Client) parsemsg(msg string) Message {
	var data string
	var username string

	ptr0 := strings.Index(msg, " :")

	if ptr0 != -1 {
		data = msg[ptr0+2:]
	} else {
		data = msg
	}

	ptr1 := strings.Index(msg, "PRIVMSG")
	ptr2 := strings.Index(msg, "!")

	if ptr1 != -1 && ptr2 != -1 && ptr2 < ptr1 {
		username := msg[:ptr2]
		username = username[1:]
	} else {
		username = "INFO"
	}

	return Message{
		Timestamp: time.Now(),
		Username:  username,
		Data:      data,
	}
}

// Only works with TLS servers
func (self *Client) Connect(config *Config) {
	cert, err := tls.LoadX509KeyPair("irc.pem", "irc.pem")
	guard.Err(err)

	tlsConfig := tls.Config{Certificates: []tls.Certificate{cert}}

	stream, err := tls.Dial("tcp", fmt.Sprintf("%v:%v", self.address, self.port), &tlsConfig)
	guard.Err(err)

	self.stream = stream
	self.Authenticate(config.Password)

	reader := bufio.NewReader(self.stream)
	textProto := textproto.NewReader(reader)

	server := self.NewServer("-", textProto)

	self.server = server
	self.serverList["-"] = server
}

func (self *Client) Send(stream net.Conn, msg string) {
	_, err := fmt.Fprintf(stream, "%v\r\n", msg)

	guard.Err(err)
}

func (self *Client) SendPrivmsg(msg string) {
	self.Send(self.stream, fmt.Sprintf("PRIVMSG #%v :%v", self.server.name, msg))
}

func (self *Client) Compute() {
	for !self.exit {
		data, err := self.server.reader.ReadLine()
		guard.Err(err)

		self.Pong(data)

		msg := self.parsemsg(data)

		self.server.chat = append(self.server.chat, msg)
		IFACE.Rehydrate()
	}
}

func (self *Client) Authenticate(passwd string) {
	self.Send(self.stream, fmt.Sprintf("PASS %v", passwd))
	self.Send(self.stream, fmt.Sprintf("NICK %v", self.nickname))
	self.Send(self.stream, fmt.Sprintf("USER %v 0 * :%v", self.nickname, self.realname))
}

func (self *Client) Pong(msg string) {
	if self.exit {
		return
	}

	if strings.HasPrefix(msg, "PING") {
		server := strings.TrimPrefix(msg, "PING ")
		self.Send(self.stream, fmt.Sprintf("PONG %v", server))
	}
}

func (self *Client) Join(serverName string) {
	self.Send(self.stream, fmt.Sprintf("JOIN #%v", serverName))

	reader := bufio.NewReader(self.stream)
	textProto := textproto.NewReader(reader)

	server := self.NewServer(serverName, textProto)

	self.server = server
	self.serverList[serverName] = server
}

func (self *Client) Disconnect() {
	channel := make(chan os.Signal, 1)

	signal.Notify(channel, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for range channel {
			self.exit = true
			self.stream.Close()
			os.Exit(0)
		}
	}()
}

func (self *Client) Run(config *Config) {
	self.Connect(config)
	go self.Compute()
}

func (self *Client) SendMessage(data string) {
	msg := Message{
		Timestamp: time.Now(),
		Username:  self.nickname,
		Data:      data,
	}

	self.SendPrivmsg(data)

	self.server.chat = append(self.server.chat, msg)

	IFACE.Rehydrate()
}

func (self *Client) GetMode() string {
	modes := map[Mode]string{
		Mode__Insert: "INSERT",
		Mode__Normal: "NORMAL",
	}

	return modes[self.mode]
}

func (self *Client) GetChat() *[]Message {
	return &self.server.chat
}

func (self *Client) ChangeServer(args ...string) []string {
	if len(args) < 1 {
		return []string{"Missing server parameter, could not change the server"}
	}

	server := args[0]

	if _, ok := self.serverList[server]; ok {
		self.server = self.serverList[server]
		return []string{fmt.Sprintf("You are currently on server: %s", server)}
	}

	return []string{fmt.Sprintf("Could not change to server: %s", server)}
}

func (self *Client) ExeBin(cmd string, args ...string) {
	bin := map[string]func(args ...string) []string{
		"help":       builtin.Help,
		"list-flags": builtin.ListFlags,
		"change":     self.ChangeServer,
	}

	res := bin[cmd](args...)

	for _, line := range res {
		msg := Message{
			Timestamp: time.Now(),
			Username:  "INFO",
			Data:      line,
		}

		self.server.chat = append(self.server.chat, msg)
	}

	IFACE.Rehydrate()
}

func (self *Client) ExeCmd(cmd string) {
	if !strings.HasPrefix(cmd, "JOIN") {
		self.Send(self.stream, cmd)
	} else {
		self.Join(cmd[5:])
	}
}

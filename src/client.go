package main

import (
	"bufio"
	"fmt"
	"ircc/src/guard"
	"net"
	"net/textproto"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const (
	Msgtype__privmsg = iota
	Msgtype__none
)

type Message struct {
	Timestamp time.Time
	Username  string
	Data      string
}

type Client struct {
	exit     bool
	port     int
	address  string
	nickname string
	server   []string
	chat     []Message
	stream   net.Conn
	reader   *textproto.Reader
	Iface    UI
}

func NewClient(addr string, port int, nickname string) Client {
	return Client{
		address:  addr,
		port:     port,
		nickname: nickname,
		Iface:    NewScreen(),
	}
}

func (self *Client) parsemsg(msg string) Message {
	var chat Message

	chat.Timestamp = time.Now()

	index := strings.Index(msg, " :")

	if index != -1 {
		chat.Data = msg[index+2:]
	} else {
		chat.Data = msg
	}

	if strings.Contains(msg, "PRIVMSG") {
		index = strings.Index(msg, "!")

		username := msg[:index]
		chat.Username = username[1:]
	} else {
		chat.Username = "INFO"
	}

	return chat
}

func (self *Client) Connect(config *Config) {
	stream, err := net.Dial("tcp", fmt.Sprintf("%v:%v", self.address, self.port))
	guard.Err(err)

	self.stream = stream
	self.Authenticate(config.Password)
}

func (self *Client) Send(stream net.Conn, msg string) {
	_, err := fmt.Fprintf(stream, "%v\r\n", msg)

	guard.Err(err)
}

func (self *Client) SendPrivMsg(msg string) {
	self.Send(self.stream, fmt.Sprintf("PRIVMSG #%v :%v", self.server, msg))
}

func (self *Client) Compute() {
	for !self.exit {
		data, err := self.reader.ReadLine()
		guard.Err(err)

		self.Pong(data)

		msg := self.parsemsg(data)

		if strings.Compare(msg.Username, "INFO") == 0 {
			self.Iface.AddLine(fmt.Sprintf("[%v] %v: %v\n", msg.Timestamp.Format("15:04"), msg.Username, msg.Data))
		} else {
			self.Iface.AddLine(fmt.Sprintf("[%v] @%v: %v\n", msg.Timestamp.Format("15:04"), msg.Username, msg.Data))
		}

		self.chat = append(self.chat, msg)
	}
}

func (self *Client) Authenticate(passwd string) {
	self.Send(self.stream, fmt.Sprintf("PASS %v", passwd))
	self.Send(self.stream, fmt.Sprintf("NICK %v", self.nickname))
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

func (self *Client) Join(server string) {
	self.Send(self.stream, fmt.Sprintf("JOIN #%v", server))
	self.server = append(self.server, server)

	reader := bufio.NewReader(self.stream)
	textProto := textproto.NewReader(reader)

	self.reader = textProto
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
	self.Join(config.Server)
	go self.Compute()

	self.Iface.Run()
}

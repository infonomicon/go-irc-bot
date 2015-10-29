package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"

	"github.com/thoj/go-ircevent"
)

var config = flag.String("config", "", "configuration file")

type Config struct {
	Irc struct {
		Ssl           bool     `json:"ssl"`
		SslVerifySkip bool     `json:"ssl_verify_skip"`
		Port          string   `json:"port"`
		Nickname      string   `json:"nickname"`
		Channels      []string `json:"channels"`
		Host          string   `json:"host"`
		Password      string   `json:"password"`
	} `json:"irc"`
	Database struct {
		Karma string `json:"karma"`
	} `json:"database"`
	Logging struct {
		Location string `json:"location"`
	} `json:"logging"`
}

func (c *Config) Load(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &c); err != nil {
		return err
	}

	if c.Irc.Nickname == "" {
		c.Irc.Nickname = "derwanksta"
	}

	if c.Irc.Host == "" {
		return errors.New("host is required.")
	}

	return nil
}

func main() {
	flag.Parse()
	c := &Config{}
	if err := c.Load(*config); err != nil {
		log.Fatal(err)
	}
	ircbot := irc.IRC(c.Irc.Nickname, c.Irc.Nickname)
	ircbot.UseTLS = c.Irc.Ssl
	if c.Irc.SslVerifySkip {
		ircbot.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}
	ircbot.Password = c.Irc.Password
	fmt.Println("connecting...")
	err := ircbot.Connect(net.JoinHostPort(c.Irc.Host, c.Irc.Port))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("connected")
	ircbot.AddCallback("001", func(e *irc.Event) {
		for _, channel := range c.Irc.Channels {
			ircbot.Join(channel)
			log.Println(fmt.Sprintf("Joined: %v", channel))
		}
	})
	ircbot.AddCallback("PRIVMSG", func(e *irc.Event) {
		channel := e.Arguments[0]
		ircbot.Privmsg(channel, e.Message())
		log.Println(fmt.Sprintf("Received Msg: %v", e.Message()))
	})
	ircbot.Loop()
}

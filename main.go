package main

import (
	"log"
	"os"
	"text/template"

	"io/ioutil"

	"bytes"

	"strings"

	"github.com/BurntSushi/toml"
	"github.com/nlopes/slack"
)

type Release struct {
	Date string
	Day  string
}

type Config struct {
	Token string
}

func main() {
	f, err := ioutil.ReadFile("config.toml")
	if err != nil {
		log.Fatalln(err)
	}
	var config Config
	if _, err := toml.Decode(string(f), &config); err != nil {
		log.Fatalln(err)
	}
	api := slack.New(config.Token)
	os.Exit(run(api))
}

func run(api *slack.Client) int {
	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for {
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.MessageEvent:
				if strings.Contains(ev.Text, "mocha-release") {
					post, err := post()
					if err != nil {
						log.Println(err)
					}
					rtm.SendMessage(rtm.NewOutgoingMessage(post, ev.Channel))
				}
			case *slack.InvalidAuthEvent:
				log.Println("Error")
				return 1
			}
		}
	}
}

func post() (string, error) {
	args, err := ioutil.ReadFile("args.toml")
	if err != nil {
		return "", err
	}
	var release Release
	if _, err := toml.Decode(string(args), &release); err != nil {
		return "", err
	}

	t, err := template.ParseFiles("template.txt")
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	t.Execute(&buf, release)
	return buf.String(), nil
}

package controllers

import (
	"fmt"
	"github.com/andboson/chebot/models"
	"github.com/labstack/gommon/log"
	"github.com/maciekmm/messenger-platform-go-sdk"
	"github.com/maciekmm/messenger-platform-go-sdk/template"
	"net/http"
)

var FbMess *messenger.Messenger

type FaceBookCheck struct {
	HubMode      string `json:"hub.mode"`
	HubChallenge string `json:"hub.challenge"`
	HubToken     string `json:"hub.verify_token"`
}

func InitFb() {
	FbMess = &messenger.Messenger{
		VerifyToken: models.Conf.FbVerifyToken,
		AppSecret:   models.Conf.FbAppSecret,
		AccessToken: models.Conf.FbPageToken,
	}
	FbMess.MessageReceived = MessageReceived
	go func() {
		http.HandleFunc("/facebook.hook", FbMess.Handler)
		log.Fatal(http.ListenAndServe(":1324", nil))
	}()
}

func MessageReceived(event messenger.Event, opts messenger.MessageOpts, msg messenger.ReceivedMessage) {
	profile, err := FbMess.GetProfile(opts.Sender.ID)
	if err != nil {
		fmt.Println(err)
		return
	}
	resp, err := FbMess.SendSimpleMessage(opts.Sender.ID, fmt.Sprintf("Hello, %s %s, %s", profile.FirstName, profile.LastName, msg.Text))
	if err != nil {
		fmt.Println(err)
	}

	FbMess.SendAction(messenger.Recipient{ID: opts.Sender.ID}, messenger.SenderActionTypingOn)
	btns := template.ButtonTemplate{
		Text: "Выберите кинотеатр",
		Buttons: []template.Button{
			{
				Title:   "Любава",
				Type:    "postback",
				Payload: "lyubava",
			},			{
				Title:   "Днепроплаза",
				Type:    "postback",
				Payload: "plaza",
			},
		},
	}

	mq := messenger.MessageQuery{}
	mq.Template(btns)
	mq.RecipientID( opts.Sender.ID)
	FbMess.SendMessage(mq)
	fmt.Printf("%+v", resp)
	log.Printf("[fb] %#v", event)
}

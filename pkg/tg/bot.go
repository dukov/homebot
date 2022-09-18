package tg

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/dukov/homebot/pkg/prometheus"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

type cmdFunc func(int64) error

type Bot struct {
	client     *tgbotapi.BotAPI
	promClient *prometheus.Client
	commands   map[string]cmdFunc
}

type BotConfig struct {
	Token    string
	PromAddr string
}

func NewBot(cfg BotConfig) (Bot, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return Bot{}, err
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)
	prom, err := prometheus.NewClient(cfg.PromAddr)
	if err != nil {
		return Bot{}, err
	}
	b := Bot{client: bot, promClient: prom}
	b.commands = map[string]cmdFunc{
		"pi_temperature": b.GetTemp,
	}
	return b, nil
}

func (b Bot) GetTemp(chatID int64) error {
	rng := v1.Range{
		Start: time.Now().Add(-15 * time.Minute),
		End:   time.Now(),
		Step:  15 * time.Second,
	}
	data, err := b.promClient.QueryRange("avg_over_time(node_hwmon_temp_celsius[5m])", rng)
	if err != nil {
		return err
	}
	writerTo, err := b.promClient.Render(data, rng)
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	_, err = writerTo.WriteTo(buf)
	if err != nil {
		return err
	}

	file := tgbotapi.FileReader{
		Name:   "temp.png",
		Reader: buf,
	}
	photo := tgbotapi.NewPhoto(chatID, file)
	_, err = b.client.Send(photo)
	return err
}

func (b Bot) processMessage(upd tgbotapi.Update) error {
	if upd.Message != nil {
		log.Printf("Got message from %d", upd.Message.Chat.ID)
		if upd.Message.IsCommand() {
			cmd, found := b.commands[upd.Message.Command()]
			if !found {
				return b.notImplemented(upd.Message.Chat.ID)
			}
			return cmd(upd.Message.Chat.ID)
		}
	}
	return nil
}

func (b Bot) errorMessage(chatID int64, err error) error {
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("error %s", err))
	_, respErr := b.client.Send(msg)
	return respErr
}

func (b Bot) notImplemented(chatID int64) error {
	msg := tgbotapi.NewMessage(chatID, "Not Implemented")
	_, respErr := b.client.Send(msg)
	return respErr
}

func (b Bot) Run() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.client.GetUpdatesChan(u)
	for update := range updates {
		if err := b.processMessage(update); err != nil {
			if sndErr := b.errorMessage(update.FromChat().ChatConfig().ChatID, err); sndErr != nil {
				log.Printf("ERROR wile sending error %s message %s", sndErr, err)
			}
		}
	}

}

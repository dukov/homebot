package main

import (
	"flag"
	"os"
	"strconv"
	"strings"

	"github.com/dukov/homebot/pkg/tg"
)

func main() {

	token := ""
	addr := ""
	chats := ""
	flag.StringVar(&token, "token", os.Getenv("PROM_BOT_TOKEN"), "Telegram Bot token")
	flag.StringVar(&addr, "prom-address", os.Getenv("PROM_ADDR"), "Url to prometheus API e.g. http://localhost:9090/")
	flag.StringVar(&chats, "chats", "", "List of chat IDs to interact with")
	flag.Parse()
	cfg := tg.BotConfig{Token: token, PromAddr: addr}
	if chats != "" {
		for _, cid := range strings.Split(chats, " ") {
			id, err := strconv.Atoi(cid)
			if err != nil {
				panic(err)
			}
			cfg.AllowedCharIDs = append(cfg.AllowedCharIDs, int64(id))
		}
	}
	b, err := tg.NewBot(cfg)
	if err != nil {
		panic(err)
	}
	b.Run()
}

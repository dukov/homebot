package main

import (
	"flag"
	"os"

	"github.com/dukov/homebot/pkg/tg"
)

func main() {

	token := ""
	addr := ""
	flag.StringVar(&token, "token", os.Getenv("PROM_BOT_TOKEN"), "Telegram Bot token")
	flag.StringVar(&addr, "prom-address", os.Getenv("PROM_ADDR"), "Url to prometheus API e.g. http://localhost:9090/")
	flag.Parse()
	b, err := tg.NewBot(tg.BotConfig{Token: token, PromAddr: addr})
	if err != nil {
		panic(err)
	}
	b.Run()
}

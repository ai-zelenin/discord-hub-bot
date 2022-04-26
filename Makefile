include .env
export

build:
	go build ./cmd/discord-hub-bot/

run: build
	BOT_TOKEN=$(BOT_TOKEN) BOT_APP_ID=$(BOT_APP_ID) ./discord-hub-bot -c res/bot.yml
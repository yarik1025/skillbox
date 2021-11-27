package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)


type binanceResp struct {
	Price float64 `json:"price,string" :"price"`
	Code  int64   `json:"code"`
}
type walett map[string]float64
var db = map[int64]walett{}



func main() {
	bot, err := tgbotapi.NewBotAPI("Your_key")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true


	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		log.Println(update.Message.Text)
	msgArr := strings.Split(update.Message.Text, " ")

	switch msgArr[0]{
	case "ADD":
		summ, err := strconv.ParseFloat(msgArr[2], 64)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Невозможно сконвертировать сумму"))
			continue
		}

		if _, ok := db[update.Message.Chat.ID]; !ok {
			db[update.Message.Chat.ID] =walett{}
		}

			db[update.Message.Chat.ID][msgArr[1]] += summ

		msg := fmt.Sprintf("Баланс: %s %f", msgArr[1],db[update.Message.Chat.ID][msgArr[1]])
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
	case "SUB":
		summ, err := strconv.ParseFloat(msgArr[2], 64)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Невозможно сконвертировать сумму"))
			continue
		}

		if _, ok := db[update.Message.Chat.ID]; !ok {
			db[update.Message.Chat.ID] =walett{}
		}

		db[update.Message.Chat.ID][msgArr[1]] -= summ

		msg := fmt.Sprintf("Баланс: %s %f", msgArr[1],db[update.Message.Chat.ID][msgArr[1]])
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))

	case "DEL":
		if  len(msgArr) > 1 {
			delete(db[update.Message.Chat.ID], msgArr[1])
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Валюта удалена"))
		} else {
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Вы ввели неправильные параметры"))
		}

	case "SHOW":
		if len(msgArr) > 1 {
			msg := "Баланс:\n"
			var usdSumm float64
			for key, value := range db[update.Message.Chat.ID] {
				coinPrice, err := getPrice(key,msgArr[1])
				if err != nil {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
				}
				usdSumm += value * coinPrice
				msg += fmt.Sprintf(" %s: %f [%.2f %s] \n", key, value, value*coinPrice, msgArr[1])
			}
			msg += fmt.Sprintf("Сумма:%2f\n", usdSumm)
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
		}		else {
			msg := "Баланс:\n"
		var usdSumm float64
		for key, value := range db[update.Message.Chat.ID] {
			coinPrice, err := getPrice(key,"USDT")
			if err != nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
			}
			usdSumm += value * coinPrice
			msg += fmt.Sprintf(" %s: %f [%.2f USDT] \n", key, value, value*coinPrice)
		}
		msg += fmt.Sprintf("Сумма:%2f\n", usdSumm)
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
		}



	default:
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда"))


	}
	}
	}
func getPrice(inputcoin string,output string) (price float64, err error) {
	resp, err :=http.Get(fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%s%s",inputcoin, output))
	if err != nil{
		return
	}
	defer resp.Body.Close()

	var jsonResp binanceResp
	err = json.NewDecoder(resp.Body).Decode(&jsonResp)
	if err !=nil {
		return
	}
	if jsonResp.Code !=0 {
		err = errors.New("Некорректная валюта")
		return
	}

	price = jsonResp.Price



	return
}
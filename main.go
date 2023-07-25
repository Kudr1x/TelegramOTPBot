package main

import (
	"context"
	"flag"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

var token string
var password string
var user string

func main() {
	token, user, password = mustData()
	createBD()
	getUpdates(createBot())
}

func createBot() (tgbotapi.UpdatesChannel, tgbotapi.BotAPI) {
	bot, _ := tgbotapi.NewBotAPI(token)

	bot.Debug = false

	updateConfig := tgbotapi.NewUpdate(0)

	updateConfig.Timeout = 30

	updates, err := bot.GetUpdatesChan(updateConfig)

	if err != nil {
		log.Fatal("Error1: " + err.Error())
	}

	return updates, *bot
}

func getUpdates(updates tgbotapi.UpdatesChannel, bot tgbotapi.BotAPI) {
	for update := range updates {
		if update.Message == nil {
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

		phoneKeyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButtonContact("Предоставить номер телефона"),
			),
		)

		msg.ReplyMarkup = phoneKeyboard

		rText := ""

		switch msg.Text {
		case "/start":
			rText = "Здравствуйте, " + update.Message.From.UserName
			msg.ReplyMarkup = phoneKeyboard
		case "":
			if update.Message.Contact != nil {
				rText = "Вы успешно зарегистрированы"
				msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)

				if !checkPhoneNumber("+" + update.Message.Contact.PhoneNumber) {
					saveUser(update.Message.Contact.PhoneNumber, msg.ChatID)
				} else {
					rText = "Вы уже зарегистрированы. Не нужно ничего подтверждать"
				}
			} else {
				rText = "Предоставьте номер телефона"
			}
		default:
			rText = "Предоставьте номер телефона"
		}

		msg.Text = rText

		if _, err := bot.Send(msg); err != nil {
			log.Fatal("Error2: " + err.Error())
		}
	}
}

func mustData() (string, string, string) {
	token := flag.String("token", "", "token for bot")
	user := flag.String("user", "", "mongodb user")
	password := flag.String("password", "", "mongodb password")

	flag.Parse()

	if *token == "" {
		log.Fatal("net token'a")
	}

	if *user == "" {
		log.Fatal("net user'a")
	}

	if *password == "" {
		log.Fatal("net password'a")
	}

	return *token, *user, *password
}

func createBD() {
	uri := "mongodb+srv://" + user + ":" + password + "@otp.aujeywe.mongodb.net/?retryWrites=true&w=majority"

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, _ = mongo.Connect(ctx, options.Client().ApplyURI(uri))
}

func saveUser(phone string, id int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := client.Database("TelegramOTPbot").Collection("Users")
	_, err := collection.InsertOne(ctx, bson.D{{Key: "phone", Value: "+" + phone}, {Key: "chat_id", Value: id}})

	if err != nil {
		log.Fatal("Error3: " + err.Error())
	}
}

func checkPhoneNumber(phone string) bool {
	filter := bson.M{"phone": phone}

	collection := client.Database("TelegramOTPbot").Collection("Users")

	var result bson.M

	err := collection.FindOne(context.Background(), filter).Decode(&result)

	if err == nil {
		return true
	} else {
		return false
	}
}

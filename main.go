package main

import (
	"fmt"
	tPack "mytorrentbot/cmd/torrent"
	"os"
	"strconv"
	"strings"

	"github.com/anacrolix/torrent"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func init() {
	log.Out = os.Stdout
	log.Formatter = &logrus.TextFormatter{
		FullTimestamp: true,
	}
}

func main() {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	log.Infof("TokenIs %v", botToken)
	if botToken == "" {
		log.Fatal("Токен бота не найден. Укажите TELEGRAM_BOT_TOKEN в переменных окружения.")
	}
	envAllowedUser := os.Getenv("ALLOWED_USERS")
	allowedUser, _ := strconv.ParseInt(envAllowedUser, 10, 64)
	if envAllowedUser == "" {
		log.Fatal("Доступные пользователи не найдены. Сожалею.")
	}
	log.Infof("Allowed USERS: %v", allowedUser)
	allowedUserIDs := []int64{allowedUser}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Авторизован как %s\n", bot.Self.UserName)
	config := torrent.NewDefaultClientConfig()
	config.DataDir = os.Getenv("DATA_DIR")
	if config.DataDir == "" {
		log.Fatal("Директория для загрузки не задана.")
	}
	client, err := torrent.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil {
			continue
		}

		userID := update.Message.Chat.ID
		if !isUserAllowed(userID, allowedUserIDs) {
			textDisallow := fmt.Sprintf("Пользователь с ID %d идет нахуй, потому что я так сказал.\n", userID)
			reply := tgbotapi.NewMessage(update.Message.Chat.ID, textDisallow)
			bot.Send(reply)
			continue
		}

		// Ваш код для обработки сообщения от разрешенного пользователя
		fmt.Printf("Получено сообщение от разрешенного пользователя %d: %s\n", userID, update.Message.Text)

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		if update.Message.IsCommand() {
			cmd := update.Message.Command()
			switch cmd {
			case "status":
				for _, t := range client.Torrents() {
					<-t.GotInfo()
					tPack.SendTorrents(bot, update.Message.Chat.ID, t)
				}
				// Обработка команды "/status"
				// ...
			default:
				// Обработка других команд
				// ...
			}
		}

		if strings.HasPrefix(update.Message.Text, "magnet:") {
			log.Infof("Magnet detected")
			go tPack.DownloadTorrent(client, bot, update.Message.Chat.ID, update.Message.Text)
			reply := tgbotapi.NewMessage(update.Message.Chat.ID, "Запущено скачивание торрента.")
			bot.Send(reply)
		}
	}

}

// Функция для проверки, является ли пользователь разрешенным
func isUserAllowed(userID int64, allowedUserIDs []int64) bool {
	for _, allowedID := range allowedUserIDs {
		if userID == allowedID {
			return true
		}
	}
	return false
}

package main

import (
	"log"
	"os"
	"strings"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/anacrolix/torrent"
)

func main() {
	// Загрузка токена бота из переменной окружения или файла .env
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("Токен бота не найден. Укажите TELEGRAM_BOT_TOKEN в переменных окружения.")
	}

	// Инициализация Telegram бота
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Авторизован как %s\n", bot.Self.UserName)

	// Инициализация торрент-клиента
	client, err := torrent.NewClient(nil)
	if err != nil {
		log.Fatal(err)
	}

	// Обработка входящих сообщений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		if update.Message.IsCommand() {
			continue
		}

		if strings.HasPrefix(update.Message.Text, "magnet:") {
			go downloadTorrent(client, update.Message.Text)
			reply := tgbotapi.NewMessage(update.Message.Chat.ID, "Запущено скачивание торрента.")
			bot.Send(reply)
		}
	}
}

func downloadTorrent(client *torrent.Client, magnetLink string) {
	torrent, err := client.AddMagnet(magnetLink)
	if err != nil {
		log.Println(err)
		return
	}

	// Ожидание завершения загрузки
	<-torrent.GotInfo()
	log.Printf("Торрент %s: Получена информация, начало загрузки...\n", torrent.Info().Name)

	// Ожидание завершения загрузки
	torrent.DownloadAll()

	log.Printf("Торрент %s: Загрузка завершена.\n", torrent.Info().Name)

	// Закрытие торрент-сессии
	client.Close()
}

package torrentPack

import (
	"fmt"
	"os"
	"time"

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

func DownloadTorrent(client *torrent.Client, bot *tgbotapi.BotAPI, chatID int64, magnetLink string) {
	torrent, err := client.AddMagnet(magnetLink)
	if err != nil {
		log.Error(err)
		return
	}
	// Ожидание завершения загрузки информации о торренте
	<-torrent.GotInfo()
	log.Infof("Торрент %s: Получена информация, начало загрузки...\n", torrent.Info().Name)

	// Логирование информации о файлах в торренте
	for i, file := range torrent.Info().Files {
		log.Infof("Файл %d: %s, Размер: %d байт\n", i+1, file.Path, file.Length)
	}

	// Создание таймера для периодического обновления статуса загрузки
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	torrent.DownloadAll()

	// Начало слушателя событий
	for {
		leng := torrent.Length() / 1024
		select {
		case <-torrent.Closed():
			log.Infof("Торрент %s: Загрузка завершена\n", torrent.Info().Name)
			return
		case <-ticker.C:
			// Отправка результатов Torrents
			kByte := torrent.BytesCompleted() / 1024
			if kByte >= leng {
				torrent.Drop()
			}
		}
	}
}

func SendTorrents(bot *tgbotapi.BotAPI, chatID int64, client *torrent.Torrent) {
	t := client
	messageText := "Список торрентов:\n"
	kByte := t.BytesCompleted() / 1024
	leng := t.Length() / 1024
	log.Infof("Длина: %v и скачано %v", leng, kByte)

	progress := "неизвестен"
	if t.Length() > 0 {
		progress = fmt.Sprintf("%.2f%%", float64(t.BytesCompleted())/float64(t.Length())*100)
	}
	messageText += "Status\n"
	messageText += fmt.Sprintf("Название: %s, Загружено: %s\n", t.Info().Name, progress)
	messageText += "\n"

	message := tgbotapi.NewMessage(chatID, messageText)
	bot.Send(message)
}

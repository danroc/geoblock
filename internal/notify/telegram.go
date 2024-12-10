package notify

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/danroc/geoblock/internal/config"
)

const (
	ServicePrefixTelegram = "telegram"
)

const (
	JSONMimeType = "application/json"
)

type SendMessageRequest struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
}

func (r SendMessageRequest) JSON() []byte {
	encoded, _ := json.Marshal(r)
	return encoded
}

type TelegramService struct {
	token string
	chats []int64
}

func NewTelegramService(config config.Telegram) *TelegramService {
	return &TelegramService{
		token: config.Token,
		chats: config.Chats,
	}
}

func (s *TelegramService) Send(message string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", s.token)

	var errs []error
	for _, chat := range s.chats {
		req := SendMessageRequest{
			ChatID: chat,
			Text:   message,
		}

		if _, err := http.Post(
			url, JSONMimeType, bytes.NewBuffer(req.JSON()),
		); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

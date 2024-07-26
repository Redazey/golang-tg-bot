package tg

import (
	"fmt"
	"strings"
	"tgssn/pkg/logger"

	"tgssn/internal/model/messages"
	"tgssn/pkg/errors"

	types "tgssn/internal/model/bottypes"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type HandlerFunc func(tgUpdate tgbotapi.Update, c *Client, msgModel *messages.Model)

func (f HandlerFunc) RunFunc(tgUpdate tgbotapi.Update, c *Client, msgModel *messages.Model) {
	f(tgUpdate, c, msgModel)
}

type Client struct {
	client                *tgbotapi.BotAPI
	handlerProcessingFunc HandlerFunc // Функция обработки входящих сообщений.
}

func New(tokenStr string, handlerProcessingFunc HandlerFunc) (*Client, error) {
	client, err := tgbotapi.NewBotAPI(tokenStr)
	if err != nil {
		return nil, errors.Wrap(err, "Ошибка NewBotAPI")
	}

	return &Client{
		client:                client,
		handlerProcessingFunc: handlerProcessingFunc,
	}, nil
}

func (c *Client) SendMessage(text string, userID int64) error {
	msg := tgbotapi.NewMessage(userID, text)
	msg.ParseMode = "markdown"
	_, err := c.client.Send(msg)
	if err != nil {
		return errors.Wrap(err, "Ошибка отправки сообщения client.Send")
	}
	return nil
}

func (c *Client) ListenUpdates(msgModel *messages.Model) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := c.client.GetUpdatesChan(u)

	logger.Info("Start listening for tg messages")

	for update := range updates {
		// Функция обработки сообщений (обернутая в middleware).
		c.handlerProcessingFunc.RunFunc(update, c, msgModel)
		//вместо ProcessingMessages(update, c, msgModel)
	}
}

// ProcessingMessages функция обработки сообщений.
func ProcessingMessages(tgUpdate tgbotapi.Update, c *Client, msgModel *messages.Model) {
	if tgUpdate.Message != nil {
		// Пользователь написал текстовое сообщение.
		logger.Info(fmt.Sprintf("[%s][%v] %s", tgUpdate.Message.From.UserName, tgUpdate.Message.From.ID, tgUpdate.Message.Text))
		err := msgModel.IncomingMessage(messages.Message{
			Text:            tgUpdate.Message.Text,
			UserID:          tgUpdate.Message.From.ID,
			UserName:        tgUpdate.Message.From.UserName,
			UserDisplayName: strings.TrimSpace(tgUpdate.Message.From.FirstName + " " + tgUpdate.Message.From.LastName),
		})
		if err != nil {
			logger.Error("error processing message:", zap.Error(err))
		}
	} else if tgUpdate.CallbackQuery != nil {
		// Пользователь нажал кнопку.
		logger.Info(fmt.Sprintf("[%s][%v] Callback: %s", tgUpdate.CallbackQuery.From.UserName, tgUpdate.CallbackQuery.From.ID, tgUpdate.CallbackQuery.Data))
		callback := tgbotapi.NewCallback(tgUpdate.CallbackQuery.ID, tgUpdate.CallbackQuery.Data)
		if _, err := c.client.Request(callback); err != nil {
			logger.Error("Ошибка Request callback:", zap.Error(err))
		}
		if err := deleteInlineButtons(c, tgUpdate.CallbackQuery.From.ID, tgUpdate.CallbackQuery.Message.MessageID, tgUpdate.CallbackQuery.Message.Text); err != nil {
			logger.Error("Ошибка удаления кнопок:", zap.Error(err))
		}
		err := msgModel.IncomingMessage(messages.Message{
			Text:            tgUpdate.CallbackQuery.Data,
			UserID:          tgUpdate.CallbackQuery.From.ID,
			UserName:        tgUpdate.CallbackQuery.From.UserName,
			UserDisplayName: strings.TrimSpace(tgUpdate.CallbackQuery.From.FirstName + " " + tgUpdate.CallbackQuery.From.LastName),
			IsCallback:      true,
			CallbackMsgID:   tgUpdate.CallbackQuery.InlineMessageID,
		})
		if err != nil {
			logger.Error("error processing message from callback:", zap.Error(err))
		}
	}
}

// ShowInlineButtons Отображение кнопок меню под сообщением с ответом.
// Их нажатие ожидает коллбек-ответ.
func (c *Client) ShowInlineButtons(msgText string, buttons types.TgRowButtons, userID int64) error {
	buttns := make([]tgbotapi.KeyboardButton, 0, len(buttons))
	for i := 0; i < len(buttons); i++ {
		button := tgbotapi.NewKeyboardButton(buttons[i].Text)
		button.Text = buttons[i].Text

		// Создаем кнопку
		buttns = append(buttns, button)
	}

	// Создаем клавиатуру
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			buttns...,
		),
	)

	// Настраиваем параметры сообщения
	msg := tgbotapi.NewMessage(userID, msgText)
	msg.ReplyMarkup = keyboard

	_, err := c.client.Send(msg)
	if err != nil {
		logger.Error("Ошибка отправки сообщения", zap.Error(err))
		return errors.Wrap(err, "client.Send with inline-buttons")
	}

	return nil
}

func deleteInlineButtons(c *Client, userID int64, msgID int, sourceText string) error {
	msg := tgbotapi.NewEditMessageText(userID, msgID, sourceText)
	_, err := c.client.Send(msg)
	if err != nil {
		logger.Error("Ошибка отправки сообщения", zap.Error(err))
		return errors.Wrap(err, "client.Send remove inline-buttons")
	}
	return nil
}

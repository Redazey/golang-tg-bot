package messages

import (
	"context"
	"fmt"
	"strings"
	"tgssn/pkg/logger"

	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"

	types "tgssn/internal/model/bottypes"
)

// Область "Внешний интерфейс": начало.

// MessageSender Интерфейс для работы с сообщениями.
type MessageSender interface {
	SendMessage(text string, userID int64) error
	ShowInlineButtons(text string, buttons []types.TgRowButtons, userID int64) (int, error)
	EditInlineButtons(text string, msgID int, buttons []types.TgRowButtons, userID int64) error
	ShowKeyboardButtons(text string, buttons types.TgKbRowButtons, userID int64) error
}

// UserDataStorage Интерфейс для работы с хранилищем данных.
type UserDataStorage interface {
	CheckIfUserExistAndAdd(ctx context.Context, userID int64) (bool, error)
	InsertUserDataRecord(ctx context.Context, userID int64, rec types.UserDataRecord) (bool, error)
	AddUserLimit(ctx context.Context, userID int64, limits float64, userName string) error
	GetUserLimit(ctx context.Context, userID int64) (float64, error)
	//----mock
	GetUserOrdersCount(ctx context.Context, userID int64) (int64, error)
}

// Model Модель бота (клиент, хранилище, последние команды пользователя)
type Model struct {
	ctx             context.Context
	tgClient        MessageSender   // Клиент.
	storage         UserDataStorage // Хранилище пользовательской информации.
	lastInlineKbMsg map[int64]int
}

// New Генерация сущности для хранения клиента ТГ и хранилища пользователей
func New(ctx context.Context, tgClient MessageSender, storage UserDataStorage) *Model {
	return &Model{
		ctx:             ctx,
		tgClient:        tgClient,
		storage:         storage,
		lastInlineKbMsg: map[int64]int{},
	}
}

// Message Структура сообщения для обработки.
type Message struct {
	Text            string
	UserID          int64
	UserName        string
	UserDisplayName string
	IsCallback      bool
	CallbackMsgID   string
}

func (s *Model) GetCtx() context.Context {
	return s.ctx
}

func (s *Model) SetCtx(ctx context.Context) {
	s.ctx = ctx
}

// IncomingMessage Обработка входящего сообщения.
func (s *Model) IncomingMessage(msg Message) error {
	span, ctx := opentracing.StartSpanFromContext(s.ctx, "IncomingMessage")
	s.ctx = ctx
	defer span.Finish()

	// Распознавание стандартных команд.
	if isNeedReturn, err := checkBotCommands(s, msg); err != nil || isNeedReturn {
		return err
	}

	if isNeedReturn, err := callbacksCommands(s, msg); err != nil || isNeedReturn {
		return err
	}

	// Отправка ответа по умолчанию.
	return s.tgClient.SendMessage(TxtUnknownCommand, msg.UserID)
}

// Область "Внешний интерфейс": конец.

// Распознавание стандартных команд бота.
func checkBotCommands(s *Model, msg Message) (bool, error) {
	span, ctx := opentracing.StartSpanFromContext(s.ctx, "checkBotCommands")
	s.ctx = ctx
	defer span.Finish()

	switch msg.Text {
	case "/start":
		displayName := msg.UserDisplayName
		if len(displayName) == 0 {
			displayName = msg.UserName
		}

		if err := s.tgClient.ShowKeyboardButtons(fmt.Sprintf(TxtStart, displayName), BtnStart, msg.UserID); err != nil {
			return false, err
		}

		return true, nil
	case "Categories":
		lastMsgID, err := s.tgClient.ShowInlineButtons(TxtCtgs, BtnCtgs, msg.UserID)
		if err != nil {
			return false, err
		}
		s.lastInlineKbMsg[msg.UserID] = lastMsgID
		return true, nil
	case "Profile":
		balance, err := s.storage.GetUserLimit(ctx, msg.UserID)
		if err != nil {
			return false, err
		}

		orders, err := s.storage.GetUserOrdersCount(ctx, msg.UserID)
		if err != nil {
			return false, err
		}

		lastMsgID, err := s.tgClient.ShowInlineButtons(
			fmt.Sprintf(TxtProfile, msg.UserID, balance, orders),
			BtnProfile,
			msg.UserID,
		)
		if err != nil {
			return false, err
		}
		s.lastInlineKbMsg[msg.UserID] = lastMsgID

		return true, nil
	case "Support":
		s.tgClient.SendMessage(TxtSup, msg.UserID)
		return true, nil
	case "/help":
		return true, s.tgClient.SendMessage(TxtHelp, msg.UserID)
	}

	// Команда не распознана.
	return false, nil
}

// callbacks
func callbacksCommands(s *Model, msg Message) (bool, error) {
	if msg.IsCallback {
		span, ctx := opentracing.StartSpanFromContext(s.ctx, "checkIfCoiceCurrency")
		s.ctx = ctx
		defer span.Finish()

		if strings.Contains(msg.Text, "back") {
			lastMsgID, err := s.tgClient.ShowInlineButtons(TxtCtgs, BtnCtgs, msg.UserID)
			if err != nil {
				return false, err
			}
			s.lastInlineKbMsg[msg.UserID] = lastMsgID
			return true, nil

		} else if strings.Contains(msg.Text, "CR") {
			return true, s.tgClient.EditInlineButtons(
				fmt.Sprintf(TxtReports, "Experian", TxtCRDesc),
				s.lastInlineKbMsg[msg.UserID],
				BtnCR,
				msg.UserID,
			)

		} else if strings.Contains(msg.Text, "TU") {
			return true, s.tgClient.EditInlineButtons(
				fmt.Sprintf(TxtReports, "Trans union", TxtTUDesc),
				s.lastInlineKbMsg[msg.UserID],
				BtnTU,
				msg.UserID,
			)

		} else if strings.Contains(msg.Text, "fullz") {
			return true, s.tgClient.EditInlineButtons(
				fmt.Sprintf(TxtReports, "Ready fulls", TxtFullzDesc),
				s.lastInlineKbMsg[msg.UserID],
				BtnFullz,
				msg.UserID,
			)
		}
	}

	// Команда не распознана.
	return false, nil
}

// Получение бюджета пользователя.
func getUserLimit(s *Model, userID int64) (float64, error) {
	userLimit, err := s.storage.GetUserLimit(s.ctx, userID)
	if err != nil {
		logger.Error("Ошибка получения бюджета", zap.Error(err))
		return 0, err
	}
	return userLimit, nil
}

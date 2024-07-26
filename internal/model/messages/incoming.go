package messages

import (
	"context"
	"fmt"
	"tgssn/pkg/logger"
	"time"

	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"

	types "tgssn/internal/model/bottypes"
)

// Область "Внешний интерфейс": начало.

// MessageSender Интерфейс для работы с сообщениями.
type MessageSender interface {
	SendMessage(text string, userID int64) error
	ShowInlineButtons(msgText string, buttons types.TgRowButtons, userID int64) error
}

// UserDataStorage Интерфейс для работы с хранилищем данных.
type UserDataStorage interface {
	InsertUserDataRecord(ctx context.Context, userID int64, rec types.UserDataRecord, userName string, limitPeriod time.Time) (bool, error)
	SetUserCurrency(ctx context.Context, userID int64, currencyName string, userName string) error
	GetShopCategories(ctx context.Context, userID int64) ([]string, error)
	GetUserLimit(ctx context.Context, userID int64) (int64, error)
	SetUserLimit(ctx context.Context, userID int64, limits int64, userName string) error
	//----mock
	GetUserBalance(ctx context.Context, userID int64) (float64, error)
	GetUserOrdersCount(ctx context.Context, userID int64) (int64, error)
}

// Model Модель бота (клиент, хранилище, последние команды пользователя)
type Model struct {
	ctx             context.Context
	tgClient        MessageSender    // Клиент.
	storage         UserDataStorage  // Хранилище пользовательской информации.
	lastUserCat     map[int64]string // Последняя выбранная пользователем категория.
	lastUserCommand map[int64]string // Последняя выбранная пользователем команда.
}

// New Генерация сущности для хранения клиента ТГ и хранилища пользователей
func New(ctx context.Context, tgClient MessageSender, storage UserDataStorage) *Model {
	return &Model{
		ctx:             ctx,
		tgClient:        tgClient,
		storage:         storage,
		lastUserCat:     map[int64]string{},
		lastUserCommand: map[int64]string{},
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

	//lastUserCat := s.lastUserCat[msg.UserID]
	//lastUserCommand := s.lastUserCommand[msg.UserID]

	// Обнуление выбранной категории и команды.
	s.lastUserCat[msg.UserID] = ""
	s.lastUserCommand[msg.UserID] = ""

	// Распознавание стандартных команд.
	if isNeedReturn, err := checkBotCommands(s, msg); err != nil || isNeedReturn {
		return err
	}

	if isNeedReturn, err := categoriesBtn(s, msg); err != nil || isNeedReturn {
		return err
	}

	if isNeedReturn, err := profileBtn(s, msg); err != nil || isNeedReturn {
		return err
	}

	if isNeedReturn, err := supportBtn(s, msg); err != nil || isNeedReturn {
		return err
	}

	// Отправка ответа по умолчанию.
	return s.tgClient.SendMessage(TxtUnknownCommand, msg.UserID)
}

// Область "Внешний интерфейс": конец.

// Область "Служебные функции": начало.

// Область "Распознавание входящих команд": начало.

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

		return true, s.tgClient.ShowInlineButtons(fmt.Sprintf(TxtStart, displayName), BtnStart, msg.UserID)
	case "/help":
		return true, s.tgClient.SendMessage(TxtHelp, msg.UserID)
	}

	// Команда не распознана.
	return false, nil
}

func categoriesBtn(s *Model, msg Message) (bool, error) {
	span, ctx := opentracing.StartSpanFromContext(s.ctx, "checkBotCommands")
	s.ctx = ctx
	defer span.Finish()

	switch msg.Text {
	case "Categories":
		return true, s.tgClient.ShowInlineButtons(TxtCtgs, BtnCtgs, msg.UserID)
	case "/help":
		return true, s.tgClient.SendMessage(TxtHelp, msg.UserID)
	}

	// Команда не распознана.
	return false, nil
}

func profileBtn(s *Model, msg Message) (bool, error) {
	span, ctx := opentracing.StartSpanFromContext(s.ctx, "checkBotCommands")
	s.ctx = ctx
	defer span.Finish()

	switch msg.Text {
	case "Profile":
		balance, err := s.storage.GetUserBalance(ctx, msg.UserID)
		if err != nil {
			return false, err
		}
		orders, err := s.storage.GetUserOrdersCount(ctx, msg.UserID)
		if err != nil {
			return false, err
		}

		s.tgClient.SendMessage(fmt.Sprintf(TxtProfile, msg.UserID, balance, orders), msg.UserID)
		return true, nil
	case "/help":
		return true, s.tgClient.SendMessage(TxtHelp, msg.UserID)
	}

	// Команда не распознана.
	return false, nil
}

func supportBtn(s *Model, msg Message) (bool, error) {
	span, ctx := opentracing.StartSpanFromContext(s.ctx, "checkBotCommands")
	s.ctx = ctx
	defer span.Finish()

	switch msg.Text {
	case "Support":
		s.tgClient.SendMessage(TxtSup, msg.UserID)
		return true, nil
	case "/help":
		return true, s.tgClient.SendMessage(TxtHelp, msg.UserID)
	}

	// Команда не распознана.
	return false, nil
}

// Получение бюджета пользователя.
func getUserLimit(s *Model, userID int64) (int64, error) {
	userLimit, err := s.storage.GetUserLimit(s.ctx, userID)
	if err != nil {
		logger.Error("Ошибка получения бюджета", zap.Error(err))
		return 0, err
	}
	return userLimit, nil
}

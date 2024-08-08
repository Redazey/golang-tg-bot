package messages

import (
	"context"

	"github.com/opentracing/opentracing-go"

	types "tgssn/internal/model/bottypes"
)

// Область "Внешний интерфейс": начало.

// MessageSender Интерфейс для работы с сообщениями.
type MessageSender interface {
	SendMessage(text string, userID int64) (int, error)
	ShowInlineButtons(text string, buttons []types.TgRowButtons, userID int64) (int, error)
	EditInlineButtons(text string, msgID int, userID int64, buttons []types.TgRowButtons) error
	ShowKeyboardButtons(text string, buttons types.TgKbRowButtons, userID int64) error
	DeleteInlineButtons(userID int64, msgID int, sourceText string) error
	DeleteMsg(userID int64, msgID int)
	ReplyMessage(FromUserID int64, ToUserID int64, msgID int) error
}

// UserDataStorage Интерфейс для работы с хранилищем данных.
type UserDataStorage interface {
	GetCategoryInfo(ctx context.Context, CtgID int64) (map[string]any, error)
	CheckIfUserExistAndAdd(ctx context.Context, userID int64) (bool, error)
	InsertUserDataRecord(ctx context.Context, userID int64, rec types.UserDataRecord) (bool, error)
	AddUserLimit(ctx context.Context, userID int64, limits float64) error
	GetUserLimit(ctx context.Context, userID int64) (float64, error)
	CheckIfUserRecordsExist(ctx context.Context, userID int64) (int64, error)
	GetUserOrders(ctx context.Context, userID int64) ([]types.UserDataRecord, error)
	ChangeWorkerStatus(ctx context.Context, userID int64, status bool) (bool, error)
	CheckIfWorkerExistAndAdd(ctx context.Context, userID int64, name string) (bool, error)
	GetCtgInfoFromName(ctx context.Context, name string) (map[string]any, error)
	CreateTicket(ctx context.Context, workerID int64, buyerID int64, ctgID int64) (bool, error)
	GetTicketInfo(ctx context.Context, workerID int64) (map[string]any, error)
	UpdateTicketStatus(ctx context.Context, workerID int64, status string) error
}

type userInteractions struct {
	command     string
	ticket      string
	orderPage   int
	ordersPages []types.UserDataRecord
}

// Model Модель бота (клиент, хранилище, последние команды пользователя)
type Model struct {
	ctx                 context.Context
	tgClient            MessageSender   // Клиент.
	storage             UserDataStorage // Хранилище пользовательской информации.
	lastInlineKbMsg     map[int64]int
	lastUserInteraction map[int64]*userInteractions
}

// New Генерация сущности для хранения клиента ТГ и хранилища пользователей
func New(ctx context.Context, tgClient MessageSender, storage UserDataStorage) *Model {
	return &Model{
		ctx:                 ctx,
		tgClient:            tgClient,
		storage:             storage,
		lastInlineKbMsg:     map[int64]int{},
		lastUserInteraction: map[int64]*userInteractions{},
	}
}

// Message Структура сообщения для обработки.
type Message struct {
	Text            string
	MessageID       int
	IsDocument      bool
	UserID          int64
	UserName        string
	UserDisplayName string
	IsCallback      bool
	CallbackMsgID   int
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

	var lastUserCommand string
	if interaction, ok := s.lastUserInteraction[msg.UserID]; ok && interaction != nil {
		lastUserCommand = interaction.command
	} else {
		s.lastUserInteraction[msg.UserID] = &userInteractions{}
	}

	s.lastUserInteraction[msg.UserID].command = ""

	// Распознавание стандартных команд.
	if isNeedReturn, err := CheckBotCommands(s, msg); err != nil || isNeedReturn {
		return err
	}

	if isNeedReturn, err := CallbacksCommands(s, msg); err != nil || isNeedReturn {
		return err
	}

	if isNeedReturn, err := CheckIfEnterCmd(s, msg, lastUserCommand); err != nil || isNeedReturn {
		return err
	}

	// Отправка ответа по умолчанию.
	_, err := s.tgClient.SendMessage(TxtUnknownCommand, msg.UserID)
	return err
}

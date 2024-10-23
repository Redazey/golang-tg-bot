package messages

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"

	"tgseller/config"
	types "tgseller/internal/model/bottypes"
	"tgseller/pkg/logger"
)

// Область "Внешний интерфейс": начало.

// Struct для создания платежного счета
type CreateInvoiceRequest struct {
	CurrencyType string  `json:"currency_type"`
	Asset        string  `json:"asset"`
	Fiat         string  `json:"fiat"`
	Amount       float64 `json:"amount"`
	Description  string  `json:"description"`
	Payload      string  `json:"payload,omitempty"`
	Expires      int     `json:"expires_in"`
}

type Invoice struct {
	InvoiceID       int64   `json:"invoice_id"`
	Hash            string  `json:"hash"`
	CurrencyType    string  `json:"currency_type"`
	Asset           string  `json:"asset"`
	Fiat            string  `json:"fiat"`
	Amount          string  `json:"amount"`
	FeeAsset        string  `json:"fee_asset,omitempty"`
	FeeAmount       float64 `json:"fee_amount,omitempty"`
	PayURL          string  `json:"pay_url,omitempty" ` // deprecated
	BotInvoiceURL   string  `json:"bot_invoice_url"`
	Description     string  `json:"description,omitempty"`
	Status          string  `json:"status"`
	CreatedAt       string  `json:"created_at"`
	PaidUsdRate     string  `json:"paid_usd_rate,omitempty"`
	UsdRate         string  `json:"usd_rate,omitempty"` // deprecated
	AllowComments   bool    `json:"allow_comments"`
	AllowAnonymous  bool    `json:"allow_anonymous"`
	PaidAt          string  `json:"paid_at,omitempty"`
	PaidAnonymously bool    `json:"paid_anonymously"`
	Payload         string  `json:"payload,omitempty"`
}

type GetInvoicesParams struct {
	Status string `json:"status,omitempty"`
}

type DeleteInvoiceRequest struct {
	InvoiceID int64 `json:"invoice_id"`
}

type DeleteInvoiceResponse struct {
	Result bool
}

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
	// users
	CheckIfUserExistAndAdd(ctx context.Context, userID int64) (bool, error)
	AddUserLimit(ctx context.Context, userID int64, limits float64) error
	GetUserAccessStatus(ctx context.Context, userID int64) (bool, error)

	// refills
	InsertUserRefillRecord(ctx context.Context, userID int64, invoiceID int64, amount float64) error
	DeleteRefillRecord(ctx context.Context, invoiceID int64) error
	ChangeRefillRecordStatus(ctx context.Context, status string, invoice_id int64) error
}

type Payment interface {
	CryptoPayRequest(ctx context.Context, method string, request any) ([]byte, error)
}

// Model Модель бота (клиент, хранилище, последние команды пользователя)
type Model struct {
	ctx      context.Context
	tgClient MessageSender   // Клиент.
	storage  UserDataStorage // Хранилище пользовательской информации.
	payment  Payment         // Платёжка
	cfg      *config.Enviroment
}

// New Генерация сущности для хранения клиента ТГ и хранилища пользователей
func New(ctx context.Context, tgClient MessageSender, storage UserDataStorage, payment Payment, cfg *config.Enviroment) *Model {
	return &Model{
		ctx:      ctx,
		tgClient: tgClient,
		storage:  storage,
		payment:  payment,
		cfg:      cfg,
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

	var err error

	// Распознавание стандартных команд.
	if isNeedReturn, err := CheckBotCommands(s, msg); err != nil || isNeedReturn {
		if err != nil {
			logger.Error("Error while CheckBotCommands: ", zap.Error(err))
		}

		return err
	}

	if isNeedReturn, err := CallbacksCommands(s, msg); err != nil || isNeedReturn {
		if err != nil {
			logger.Error("Error while CallbacksCommands: ", zap.Error(err))
		}

		return err
	}
	// Отправка ответа по умолчанию.
	_, err = s.tgClient.SendMessage(TxtUnknownCommand, msg.UserID)
	return err
}

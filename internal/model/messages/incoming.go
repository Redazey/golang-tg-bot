package messages

import (
	"context"
	"fmt"
	"strings"

	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"

	"tgseller/config"
	types "tgseller/internal/model/bottypes"
	"tgseller/pkg/cache"
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

type CtgInfo struct {
	ID          int     `db:"id"`
	Name        string  `db:"name"`
	Price       float64 `db:"price"`
	Short       string  `db:"short_name"`
	Description string  `db:"description"`
	DataFormat  string  `db:"data_format"`
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
	// ctgs
	GetCtgsInfo(ctx context.Context) ([]types.CtgInfo, error)
	GetCtgShorts(ctx context.Context) ([]string, error)

	// users
	CheckIfUserExistAndAdd(ctx context.Context, userID int64) (bool, error)
	InsertUserDataRecord(ctx context.Context, userID int64, ctgInfo types.CtgInfo) (bool, error)
	AddUserLimit(ctx context.Context, userID int64, limits float64) error
	GetUserLimit(ctx context.Context, userID int64) (float64, error)
	CheckIfUserRecordsExist(ctx context.Context, userID int64) (int64, error)
	GetUserOrders(ctx context.Context, userID int64) ([]types.UserDataRecord, error)

	// refills
	InsertUserRefillRecord(ctx context.Context, userID int64, invoiceID int64, amount float64) error
	DeleteRefillRecord(ctx context.Context, invoiceID int64) error
	ChangeRefillRecordStatus(ctx context.Context, status string, invoice_id int64) error
	GetRefillRecords(ctx context.Context) ([]int, error)
	GetRefillHistory(ctx context.Context, userID int64) ([]types.UserRefillRecord, error)

	// workers
	ChangeWorkerStatus(ctx context.Context, userID int64, status bool) (bool, error)
	CheckIfWorkerExistAndAdd(ctx context.Context, userID int64, name string) (bool, error)

	// tickets
	CreateTicket(ctx context.Context, workerID int64, buyerID int64, ctgID int64) (bool, error)
	GetTicketInfo(ctx context.Context, workerID int64) (map[string]any, error)
	UpdateTicketStatus(ctx context.Context, workerID int64, status string) error
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

	var PaymentCtgs []string
	cacheCtgs, err := cache.ReadCache("ctgShorts")
	if err != nil {
		return err
	}

	if cacheCtgs == "" {
		PaymentCtgs, err = s.storage.GetCtgShorts(ctx)
		if err != nil {
			return err
		}

		if err := cache.SaveCache("ctgShorts", strings.Join(PaymentCtgs, " ")); err != nil {
			return err
		}
	} else {
		PaymentCtgs = strings.Split(cacheCtgs, " ")
	}

	var PaymentCtgsInfo []types.CtgInfo
	var cacheCtgsInfo []types.CtgInfo
	err = cache.ReadMapCache("ctgsInfo", &PaymentCtgsInfo)
	if err != nil {
		return err
	}

	if PaymentCtgsInfo == nil {
		PaymentCtgsInfo, err = s.storage.GetCtgsInfo(ctx)
		if err != nil {
			return err
		}

		if err := cache.SaveMapCache("ctgsInfo", cacheCtgsInfo); err != nil {
			return err
		}
	} else {
		PaymentCtgsInfo = cacheCtgsInfo
	}

	lastUserCommand, err := cache.ReadCache(fmt.Sprintf("%v_command", msg.UserID))
	if err != nil {
		return err
	}

	// Распознавание стандартных команд.
	if isNeedReturn, err := CheckBotCommands(s, msg, PaymentCtgs); err != nil || isNeedReturn {
		if err != nil {
			logger.Error("Error while CheckBotCommands: ", zap.Error(err))
		}

		return err
	}

	if isNeedReturn, err := CallbacksCommands(s, msg, PaymentCtgs, PaymentCtgsInfo); err != nil || isNeedReturn {
		if err != nil {
			logger.Error("Error while CallbacksCommands: ", zap.Error(err))
		}

		return err
	}

	if isNeedReturn, err := CheckIfEnterCmd(s, msg, PaymentCtgs, lastUserCommand, PaymentCtgsInfo); err != nil || isNeedReturn {
		if err != nil {
			logger.Error("Error while CheckIfEnterCmd: ", zap.Error(err))
		}

		if err := cache.SaveCache(fmt.Sprintf("%v_command", msg.UserID), ""); err != nil {
			return err
		}

		return err
	}

	// Отправка ответа по умолчанию.
	_, err = s.tgClient.SendMessage(TxtUnknownCommand, msg.UserID)
	return err
}

func GetCtgInfoFromName(ctgName string, ctgsInfo []types.CtgInfo) types.CtgInfo {
	for _, ctg := range ctgsInfo {
		if ctg.Name == ctgName {
			return ctg
		}
	}

	return types.CtgInfo{}
}

func GetCtgInfoFromID(ctgID int64, ctgsInfo []types.CtgInfo) types.CtgInfo {
	for _, ctg := range ctgsInfo {
		if ctg.ID == ctgID {
			return ctg
		}
	}

	return types.CtgInfo{}
}

package messages

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/opentracing/opentracing-go"

	"tgssn/cmd/payment"
	types "tgssn/internal/model/bottypes"
	"tgssn/pkg/errors"
)

// Область "Внешний интерфейс": начало.

// MessageSender Интерфейс для работы с сообщениями.
type MessageSender interface {
	SendMessage(text string, userID int64) error
	ShowInlineButtons(text string, buttons []types.TgRowButtons, userID int64) (int, error)
	EditInlineButtons(text string, msgID int, userID int64, buttons []types.TgRowButtons) error
	ShowKeyboardButtons(text string, buttons types.TgKbRowButtons, userID int64) error
	DeleteInlineButtons(userID int64, msgID int, sourceText string) error
}

// UserDataStorage Интерфейс для работы с хранилищем данных.
type UserDataStorage interface {
	CheckIfUserExistAndAdd(ctx context.Context, userID int64) (bool, error)
	InsertUserDataRecord(ctx context.Context, userID int64, rec types.UserDataRecord) (bool, error)
	AddUserLimit(ctx context.Context, userID int64, limits float64) error
	GetUserLimit(ctx context.Context, userID int64) (float64, error)
	CheckIfUserRecordsExist(ctx context.Context, userID int64) (int64, error)
	GetUserOrders(ctx context.Context, userID int64) (types.UserDataRecord, error)
}

// Model Модель бота (клиент, хранилище, последние команды пользователя)
type Model struct {
	ctx             context.Context
	tgClient        MessageSender   // Клиент.
	storage         UserDataStorage // Хранилище пользовательской информации.
	lastInlineKbMsg map[int64]int
	lastUserCommand map[int64]string
}

// New Генерация сущности для хранения клиента ТГ и хранилища пользователей
func New(ctx context.Context, tgClient MessageSender, storage UserDataStorage) *Model {
	return &Model{
		ctx:             ctx,
		tgClient:        tgClient,
		storage:         storage,
		lastInlineKbMsg: map[int64]int{},
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

	lastUserCommand := s.lastUserCommand[msg.UserID]

	// Распознавание стандартных команд.
	if isNeedReturn, err := checkBotCommands(s, msg); err != nil || isNeedReturn {
		return err
	}

	if isNeedReturn, err := callbacksCommands(s, msg); err != nil || isNeedReturn {
		return err
	}

	if isNeedReturn, err := checkIfEnterRefillBalance(s, msg, lastUserCommand); err != nil || isNeedReturn {
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
			return true, err
		}

		return true, nil
	case "Categories":
		lastMsgID, err := s.tgClient.ShowInlineButtons(TxtCtgs, BtnCtgs, msg.UserID)
		if err != nil {
			return true, err
		}
		s.lastInlineKbMsg[msg.UserID] = lastMsgID
		return true, nil
	case "Profile":
		if _, err := s.storage.CheckIfUserExistAndAdd(ctx, msg.UserID); err != nil {
			return true, err
		}

		balance, err := s.storage.GetUserLimit(ctx, msg.UserID)
		if err != nil {
			return true, err
		}

		orders, err := s.storage.CheckIfUserRecordsExist(ctx, msg.UserID)
		if err != nil {
			return true, err
		}

		lastMsgID, err := s.tgClient.ShowInlineButtons(
			fmt.Sprintf(TxtProfile, msg.UserID, balance, orders),
			BtnProfile,
			msg.UserID,
		)
		if err != nil {
			return true, err
		}
		s.lastInlineKbMsg[msg.UserID] = lastMsgID

		return true, nil
	case "Support":
		s.tgClient.SendMessage(TxtSup, msg.UserID)
		s.lastInlineKbMsg[msg.UserID] = 0
		return true, nil
	case "/help":
		s.lastInlineKbMsg[msg.UserID] = 0
		return true, s.tgClient.SendMessage(TxtHelp, msg.UserID)
	}

	// Команда не распознана.
	return false, nil
}

// callbacks
func callbacksCommands(s *Model, msg Message) (bool, error) {
	if msg.IsCallback {
		span, ctx := opentracing.StartSpanFromContext(s.ctx, "callbacksCommands")
		s.ctx = ctx
		defer span.Finish()

		// Дерево callbacks начинающихся с Categories
		if msg.Text == "backToCtg" {
			err := s.tgClient.EditInlineButtons(
				TxtCtgs,
				s.lastInlineKbMsg[msg.UserID],
				msg.UserID,
				BtnCtgs,
			)
			if err != nil {
				return true, err
			}

			return true, nil

			// выбор категории товара

		} else if msg.Text == "CR" {
			return true, s.tgClient.EditInlineButtons(
				fmt.Sprintf(TxtReports, "Experian", TxtCRDesc),
				s.lastInlineKbMsg[msg.UserID],
				msg.UserID,
				BtnCR,
			)

		} else if msg.Text == "TU" {
			return true, s.tgClient.EditInlineButtons(
				fmt.Sprintf(TxtReports, "Trans union", TxtTUDesc),
				s.lastInlineKbMsg[msg.UserID],
				msg.UserID,
				BtnTU,
			)

		} else if msg.Text == "fullz" {
			return true, s.tgClient.EditInlineButtons(
				fmt.Sprintf(TxtReports, "Ready fulls", TxtFullzDesc),
				s.lastInlineKbMsg[msg.UserID],
				msg.UserID,
				BtnFullz,
			)

			// покупка товара

		} else if strings.Contains(msg.Text, "buy") {
			if msg.Text == "buy TU" {
				s.tgClient.EditInlineButtons(TxtPaymentDesc, s.lastInlineKbMsg[msg.UserID], msg.UserID, BackToCtgBtn)
				s.lastUserCommand[msg.UserID] = "buy TU"
				return true, nil

			} else if msg.Text == "buy CR" {
				s.tgClient.EditInlineButtons(TxtPaymentDesc, s.lastInlineKbMsg[msg.UserID], msg.UserID, BackToCtgBtn)
				s.lastUserCommand[msg.UserID] = "buy CR"
				return true, nil

			} else {
				s.tgClient.DeleteInlineButtons(msg.UserID, s.lastInlineKbMsg[msg.UserID], TxtPaymentQuestion)
				s.lastUserCommand[msg.UserID] = "buy Fullz"
				return true, nil
			}

			// Дерево callbacks начинающихся с Profile
		} else if msg.Text == "backToProfile" {
			if _, err := s.storage.CheckIfUserExistAndAdd(ctx, msg.UserID); err != nil {
				return true, err
			}

			balance, err := s.storage.GetUserLimit(ctx, msg.UserID)
			if err != nil {
				return true, err
			}

			orders, err := s.storage.CheckIfUserRecordsExist(ctx, msg.UserID)
			if err != nil {
				return true, err
			}

			err = s.tgClient.EditInlineButtons(
				fmt.Sprintf(TxtProfile, msg.UserID, balance, orders),
				s.lastInlineKbMsg[msg.UserID],
				msg.UserID,
				BtnProfile,
			)
			if err != nil {
				return true, err
			}

			return true, nil

		} else if msg.Text == "refill" {
			s.lastUserCommand[msg.UserID] = "refill"
			return true, s.tgClient.EditInlineButtons(TxtPaymentQuestion, s.lastInlineKbMsg[msg.UserID], msg.UserID, BackToProfileBtn)

		} else if msg.Text == "orders" {
			orders, err := s.storage.GetUserOrders(ctx, msg.UserID)
			if err != nil {
				return true, err
			}

			return true, s.tgClient.EditInlineButtons(
				fmt.Sprintf(TxtOrderHistory, orders.RedcordID, orders.Period, orders.Category, orders.Sum),
				s.lastInlineKbMsg[msg.UserID],
				msg.UserID,
				BackToProfileBtn,
			)
		}
	}

	// Команда не распознана.
	return false, nil
}

func checkIfEnterRefillBalance(s *Model, msg Message, lastUserCommand string) (bool, error) {
	if lastUserCommand != "" {
		span, ctx := opentracing.StartSpanFromContext(s.ctx, "checkIfEnterRefillBalance")
		s.ctx = ctx
		defer span.Finish()

		if lastUserCommand == "refill" {
			userInput, err := strconv.Atoi(msg.Text)
			if err != nil || userInput == 0 {
				s.tgClient.SendMessage(TxtPaymentNotInt, msg.UserID)
				s.lastUserCommand[msg.UserID] = ""
				return true, errors.Wrap(err, "Пользователь ввёл неверное значение")
			}

			p := payment.New(ctx, s.storage)
			if paymentState, err := p.MockPay(int64(userInput)); err != nil {
				s.tgClient.SendMessage(TxtPaymentErr, msg.UserID)
				return true, errors.Wrap(err, "Ошибка при переводе средств")
			} else if !paymentState {
				s.tgClient.SendMessage(TxtPaymentNotEnough, msg.UserID)
				return true, nil
			}

			s.storage.AddUserLimit(ctx, msg.UserID, float64(userInput))
			s.tgClient.SendMessage(TxtPaymentSuccsessful, msg.UserID)
			s.lastUserCommand[msg.UserID] = ""

			return true, nil
		} else if lastUserCommand == "buy TU" {

			return true, nil
		} else if lastUserCommand == "buy CR" {

			return true, nil
		} else if lastUserCommand == "buy Fullz" {

			return true, nil
		}
	}

	return false, nil
}

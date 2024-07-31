package messages

import (
	"fmt"
	"strconv"
	"strings"
	"tgssn/cmd/payment"
	"tgssn/internal/model/bottypes"
	"tgssn/pkg/errors"
	"tgssn/pkg/logger"

	"github.com/opentracing/opentracing-go"
)

func CheckIfEnterCmd(s *Model, msg Message, lastUserCommand string) (bool, error) {
	if lastUserCommand != "" {
		span, ctx := opentracing.StartSpanFromContext(s.ctx, "checkIfEnterLastCommand")
		s.ctx = ctx
		defer span.Finish()

		if lastUserCommand == "refill" {
			userInput, err := strconv.Atoi(msg.Text)
			if err != nil || userInput == 0 {
				if err := s.tgClient.SendMessage(TxtPaymentNotInt, msg.UserID); err != nil {
					return true, err
				}
				return true, errors.Wrap(err, "Пользователь ввёл неверное значение")
			}

			p := payment.New(ctx, s.storage)
			if paymentState, err := p.MockPay(int64(userInput)); err != nil {
				s.tgClient.SendMessage(TxtPaymentErr, msg.UserID)
				return true, errors.Wrap(err, "Ошибка при переводе средств")

			} else if !paymentState {
				return true, s.tgClient.SendMessage(TxtPaymentNotEnough, msg.UserID)
			}

			s.storage.AddUserLimit(ctx, msg.UserID, float64(userInput))
			s.lastUserCommand[msg.UserID] = ""
			if s.lastInlineKbMsg[msg.UserID], err = s.tgClient.ShowInlineButtons(
				TxtPaymentSuccsessful,
				BackToCtgBtn,
				msg.UserID,
			); err != nil {
				return true, err
			}
			return true, nil

		} else if strings.Contains(lastUserCommand, "buy") {
			var err error
			var ctgName string
			var price float64

			if lastUserCommand == "buy TU" {
				ctgName = "Trans Union"
				price = 8

			} else if lastUserCommand == "buy CR" {
				ctgName = "Experian"
				price = 8

			}

			if succsessful, err := s.storage.InsertUserDataRecord(ctx, msg.UserID, bottypes.UserDataRecord{
				UserID:   msg.UserID,
				Category: ctgName,
				Sum:      price,
			}); err != nil {
				if err := s.tgClient.SendMessage(TxtError, msg.UserID); err != nil {
					return true, err
				}

				return true, err
			} else if !succsessful {
				s.lastUserCommand[msg.UserID] = ""
				if s.lastInlineKbMsg[msg.UserID], err = s.tgClient.ShowInlineButtons(
					TxtPaymentNotEnough,
					BtnRefill,
					msg.UserID,
				); err != nil {
					return true, err
				}

				return true, nil
			}

			s.lastUserCommand[msg.UserID] = ""

			if isValidDataInput(msg.Text) {
				s.lastUserTicket[msg.UserID] = msg.Text
			} else {
				s.tgClient.SendMessage(TxtWrongTicketFormat, msg.UserID)
				return true, nil
			}

			s.lastInlineKbMsg[WorkersChatID], err = s.tgClient.ShowInlineButtons(
				fmt.Sprintf(TxtForWorkers, ctgName),
				CreateInlineButtons(
					BtnWorkersChatDN,
					fmt.Sprintf(BtnWorkersChatVal, msg.UserID, ctgName),
				),
				WorkersChatID,
			)
			if err != nil {
				return true, err
			}

			return true, s.tgClient.SendMessage(TxtTicketInProccess, msg.UserID)

		} else if lastUserCommand == "goodTicket" {
			if err := s.storage.UpdateTicketStatus(ctx, msg.UserID, "good"); err != nil {
				s.tgClient.SendMessage(TxtErrorTicketUpd, msg.UserID)
				return true, err
			}

			if _, err := s.storage.ChangeWorkerStatus(ctx, msg.UserID, false); err != nil {
				s.tgClient.SendMessage(TxtErrorTicketUpd, msg.UserID)
				return true, err
			}

			s.lastUserCommand[msg.UserID] = ""

			return true, nil
		}
	}

	return false, nil
}

func isValidDataInput(input string) bool {
	parts := strings.Split(input, ";")
	if len(parts) != 7 {
		fmt.Printf("len - %v", len(parts))
		return false
	}

	if parts[0] == "" || parts[0] == " " {
		logger.Info("0")
		return false
	} else if parts[1] == "" || parts[1] == " " {
		logger.Info("1")
		return false
	} else if parts[2] == "" || parts[2] == " " {
		logger.Info("2")
		return false
	} else if parts[3] == "" || parts[3] == " " {
		logger.Info("3")
		return false
	} else if _, ok := strconv.Atoi(parts[4]); ok != nil || parts[4] == "" || parts[4] == " " {
		logger.Info("4")
		return false
	} else if parts[5] == "" || parts[5] == " " {
		logger.Info("5")
		return false
	} else if _, ok := strconv.Atoi(parts[6]); ok != nil || parts[6] == "" || parts[6] == " " {
		logger.Info("6")
		return false
	}

	return true
}

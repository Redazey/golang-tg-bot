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
				if _, err := s.tgClient.SendMessage(TxtPaymentNotInt, msg.UserID); err != nil {
					return true, err
				}
				return true, errors.Wrap(err, "Пользователь ввёл неверное значение")
			}

			p := payment.New(ctx, s.storage)
			if paymentState, err := p.MockPay(int64(userInput)); err != nil {
				s.tgClient.SendMessage(TxtPaymentErr, msg.UserID)
				return true, errors.Wrap(err, "Ошибка при переводе средств")

			} else if !paymentState {
				_, err := s.tgClient.SendMessage(TxtPaymentNotEnough, msg.UserID)
				return true, err
			}

			s.storage.AddUserLimit(ctx, msg.UserID, float64(userInput))

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
			var ctgInfo map[string]any
			var ctgName string

			if lastUserCommand == "buy TU" {
				ctgName = "Trans Union"
				if ctgInfo, err = s.storage.GetCtgInfoFromName(ctx, ctgName); err != nil {
					return true, err
				}

			} else if lastUserCommand == "buy CR" {
				ctgName = "Experian"
				if ctgInfo, err = s.storage.GetCtgInfoFromName(ctx, ctgName); err != nil {
					return true, err
				}
			}

			if succsessful, err := s.storage.InsertUserDataRecord(ctx, msg.UserID, bottypes.UserDataRecord{
				UserID:     msg.UserID,
				CategoryID: ctgInfo["id"].(int64),
			}); err != nil {
				s.tgClient.SendMessage(TxtError, msg.UserID)
				return true, err

			} else if !succsessful {
				if s.lastInlineKbMsg[msg.UserID], err = s.tgClient.ShowInlineButtons(
					TxtPaymentNotEnough,
					BtnRefill,
					msg.UserID,
				); err != nil {
					return true, err
				}

				return true, nil
			}

			if isValidDataInput(msg.Text) {
				s.lastUserInteraction[msg.UserID].command = msg.Text
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

			if err = s.storage.AddUserLimit(ctx, msg.UserID, -ctgInfo["price"].(float64)); err != nil {
				return true, err
			}

			_, err = s.tgClient.SendMessage(TxtTicketInProccess, msg.UserID)

			return true, err

		} else if lastUserCommand == "goodTicket" {

			if msg.IsDocument {
				ticketInfo, err := s.storage.GetTicketInfo(ctx, msg.UserID)
				if err != nil {
					return true, err
				}

				buyerID, ok := ticketInfo["buyer_tg_id"].(int64)
				if !ok {
					return true, err
				}

				if err := s.tgClient.ReplyMessage(msg.UserID, buyerID, msg.MessageID); err != nil {
					return true, err
				}

				if err := s.storage.UpdateTicketStatus(ctx, msg.UserID, "good"); err != nil {
					s.tgClient.SendMessage(TxtErrorTicketUpd, msg.UserID)
					return true, err
				}

				if _, err := s.storage.ChangeWorkerStatus(ctx, msg.UserID, false); err != nil {
					s.tgClient.SendMessage(TxtErrorTicketUpd, msg.UserID)
					return true, err
				}

				return true, nil
			}
			s.tgClient.SendMessage(TxtBadFile, msg.UserID)
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

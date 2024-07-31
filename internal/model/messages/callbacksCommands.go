package messages

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/opentracing/opentracing-go"
)

// callbacks
func CallbacksCommands(s *Model, msg Message) (bool, error) {
	if msg.IsCallback {
		span, ctx := opentracing.StartSpanFromContext(s.ctx, "callbacksCommands")
		s.ctx = ctx
		defer span.Finish()
		var err error

		// Дерево callbacks начинающихся с Categories
		if msg.Text == "backToCtg" {
			err = s.tgClient.EditInlineButtons(
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
				s.tgClient.DeleteInlineButtons(msg.UserID, s.lastInlineKbMsg[msg.UserID], TxtFullzPaymentDesc)
				return true, nil
			}

			// Дерево callbacks начинающихся с Profile
		} else if msg.Text == "backToProfile" {
			if _, err = s.storage.CheckIfUserExistAndAdd(ctx, msg.UserID); err != nil {
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
			s.lastUserCommand[msg.UserID] = ""

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
				fmt.Sprintf(TxtOrderHistory, orders.RecordID, orders.Period, orders.Category, orders.Sum),
				s.lastInlineKbMsg[msg.UserID],
				msg.UserID,
				BackToProfileBtn,
			)
			// Раздел Callback'ов, которые отправляют работники
		} else if strings.Contains(msg.Text, "takeTicket:") {
			callbackData := strings.Split(msg.Text, ":")
			buyerID, err := strconv.Atoi(callbackData[1])
			if err != nil {
				return true, err
			}

			reportType := callbackData[2]

			if _, err = s.storage.CheckIfWorkerExistAndAdd(ctx, msg.UserID); err != nil {
				return true, err
			}

			if succsessful, err := s.storage.CreateTicket(ctx, msg.UserID, int64(buyerID)); err != nil || !succsessful {
				s.tgClient.SendMessage(TxtBusyWorker, msg.UserID)
				return true, err
			}

			s.lastInlineKbMsg[msg.UserID], err = s.tgClient.ShowInlineButtons(
				fmt.Sprintf(TxtToWorker, reportType, s.lastUserTicket[int64(buyerID)]),
				BtnToWorker,
				msg.UserID,
			)
			if err != nil {
				return true, err
			}

			s.tgClient.DeleteMsg(WorkersChatID, msg.CallbackMsgID)
			return true, nil

			// Управление состоянием тикета работником
		} else if msg.Text == "badTicket" {
			if err := s.storage.UpdateTicketStatus(ctx, msg.UserID, "bad"); err != nil {
				s.tgClient.SendMessage(TxtErrorTicketUpd, msg.UserID)
				return true, err
			}

			if _, err := s.storage.ChangeWorkerStatus(ctx, msg.UserID, false); err != nil {
				s.tgClient.SendMessage(TxtErrorTicketUpd, msg.UserID)
				return true, err
			}

			s.lastUserCommand[msg.UserID] = ""

			return true, s.tgClient.DeleteInlineButtons(msg.UserID, msg.CallbackMsgID, TxtBadTicket)
		} else if msg.Text == "goodTicket" {
			s.lastUserCommand[msg.UserID] = "goodTicket"

			return true, s.tgClient.DeleteInlineButtons(msg.UserID, msg.CallbackMsgID, TxtGoodTicket)
		}
	}

	// Команда не распознана.
	return false, nil
}

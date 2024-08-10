package messages

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"tgssn/internal/model/bottypes"
	"tgssn/pkg/errors"
	"tgssn/pkg/logger"

	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
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
			return true, s.tgClient.EditInlineButtons(
				TxtCtgs,
				s.lastInlineKbMsg[msg.UserID],
				msg.UserID,
				BtnCtgs,
			)

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
				s.tgClient.EditInlineButtons(TxtPaymentDesc, s.lastInlineKbMsg[msg.UserID], msg.UserID, []bottypes.TgRowButtons{{BackToCtgBtn}})
				s.lastUserInteraction[msg.UserID].command = "buy TU"
				return true, nil

			} else if msg.Text == "buy CR" {
				s.tgClient.EditInlineButtons(TxtPaymentDesc, s.lastInlineKbMsg[msg.UserID], msg.UserID, []bottypes.TgRowButtons{{BackToCtgBtn}})
				s.lastUserInteraction[msg.UserID].command = "buy CR"
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

			return true, nil

		} else if msg.Text == "refill" {
			s.lastUserInteraction[msg.UserID].command = "refill"
			return true, s.tgClient.EditInlineButtons(
				TxtPaymentQuestion,
				s.lastInlineKbMsg[msg.UserID],
				msg.UserID,
				[]bottypes.TgRowButtons{{BackToProfileBtn}},
			)

		} else if msg.Text == "orders" {
			if s.lastUserInteraction[msg.UserID].ordersPages == nil {
				orders, err := s.storage.GetUserOrders(ctx, msg.UserID)
				if err != nil {
					return true, err
				}

				s.lastUserInteraction[msg.UserID].ordersPages = orders
			}

			orders := s.lastUserInteraction[msg.UserID].ordersPages
			s.lastUserInteraction[msg.UserID].orderPage = 0

			txtPage, err := getPages(ctx, s, msg, 0, orders)
			if err != nil {
				return true, err
			}

			return true, s.tgClient.EditInlineButtons(
				txtPage,
				s.lastInlineKbMsg[msg.UserID],
				msg.UserID,
				[]bottypes.TgRowButtons{{BtnOrderForward}, {BackToProfileBtn}},
			)

		} else if msg.Text == "deleteInvoice" {
			body, err := s.payment.CryptoPayRequest(
				s.ctx, "deleteInvoice",
				DeleteInvoiceRequest{InvoiceID: s.lastUserInteraction[msg.UserID].paymentID},
			)
			if err != nil {
				logger.Error("Failed to delete invoice", zap.Error(err))

				return true, err
			}

			var result bool
			if err = json.Unmarshal(body, &result); err != nil {
				logger.Debug("Error while unmarshaling message")

				return true, err
			}

			if !result {
				logger.Info("Failed to delete invoice")

				return true, err
			}

			if err = s.storage.DeleteRefillRecord(ctx, s.lastUserInteraction[msg.UserID].paymentID); err != nil {
				return true, err
			}

			return true, s.tgClient.EditInlineButtons(
				TxtPaymentCanceled,
				s.lastInlineKbMsg[msg.UserID],
				msg.UserID,
				[]bottypes.TgRowButtons{{BackToProfileBtn}},
			)

		} else if strings.Contains(strings.Join(PaymentMethods, " "), msg.Text) {
			amount := s.lastUserInteraction[msg.UserID].paymentVal
			if amount == 0 {
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

				return true, nil
			}

			invoiceReq := CreateInvoiceRequest{
				CurrencyType: "fiat",
				Asset:        msg.Text,
				Fiat:         "USD",
				Amount:       s.lastUserInteraction[msg.UserID].paymentVal,
				Description:  fmt.Sprintf(TxtRefillDesc, s.lastUserInteraction[msg.UserID].paymentVal, msg.Text),
				Payload:      fmt.Sprintf("%v", msg.UserID),
				Expires:      s.cfg.PaymentEX,
			}

			body, err := s.payment.CryptoPayRequest(ctx, "createInvoice", invoiceReq)
			if err != nil {
				s.tgClient.SendMessage(TxtPaymentErr, msg.UserID)
				return true, errors.Wrap(err, "Ошибка при переводе средств")

			}

			var paymentState Invoice

			err = json.Unmarshal(body, &paymentState)
			if err != nil {
				return true, err
			}

			s.lastUserInteraction[msg.UserID].paymentID = paymentState.InvoiceID
			BtnRefillRequest[0][0].URL = paymentState.BotInvoiceURL

			if err = s.storage.InsertUserRefillRecord(ctx, msg.UserID, paymentState.InvoiceID, amount); err != nil {
				return true, err
			}

			if s.lastInlineKbMsg[msg.UserID], err = s.tgClient.ShowInlineButtons(
				TxtRefillReqCreated,
				BtnRefillRequest,
				msg.UserID,
			); err != nil {
				return true, err
			}

			return true, nil

		} else if msg.Text == "pageBack" {
			if s.lastUserInteraction[msg.UserID].ordersPages == nil {
				orders, err := s.storage.GetUserOrders(ctx, msg.UserID)
				if err != nil {
					return true, err
				}

				s.lastUserInteraction[msg.UserID].ordersPages = orders
			}

			s.lastUserInteraction[msg.UserID].orderPage -= OrdersInPage

			var (
				orders    = s.lastUserInteraction[msg.UserID].ordersPages
				page      = s.lastUserInteraction[msg.UserID].orderPage
				btnOrders []bottypes.TgRowButtons
			)

			txtPage, err := getPages(ctx, s, msg, page, orders)
			if err != nil {
				return true, err
			}

			if page == 0 {
				btnOrders = []bottypes.TgRowButtons{{BtnOrderForward}, {BackToProfileBtn}}
			} else {
				btnOrders = []bottypes.TgRowButtons{{BtnOrderBack, BtnOrderForward}, {BackToProfileBtn}}
			}

			return true, s.tgClient.EditInlineButtons(
				txtPage,
				s.lastInlineKbMsg[msg.UserID],
				msg.UserID,
				btnOrders,
			)

		} else if msg.Text == "pageForward" {
			if s.lastUserInteraction[msg.UserID].ordersPages == nil {
				orders, err := s.storage.GetUserOrders(ctx, msg.UserID)
				if err != nil {
					return true, err
				}

				s.lastUserInteraction[msg.UserID].ordersPages = orders
			}

			s.lastUserInteraction[msg.UserID].orderPage += OrdersInPage

			var (
				orders    = s.lastUserInteraction[msg.UserID].ordersPages
				page      = s.lastUserInteraction[msg.UserID].orderPage
				btnOrders []bottypes.TgRowButtons
			)

			txtPage, err := getPages(ctx, s, msg, page, orders)
			if err != nil {
				return true, err
			}

			if page+OrdersInPage >= len(orders) {
				btnOrders = []bottypes.TgRowButtons{{BtnOrderBack}, {BackToProfileBtn}}
			} else {
				btnOrders = []bottypes.TgRowButtons{{BtnOrderBack, BtnOrderForward}, {BackToProfileBtn}}
			}

			return true, s.tgClient.EditInlineButtons(
				txtPage,
				s.lastInlineKbMsg[msg.UserID],
				msg.UserID,
				btnOrders,
			)

			// Раздел Callback'ов, которые отправляют работники
		} else if strings.Contains(msg.Text, "takeTicket:") {
			callbackData := strings.Split(msg.Text, ":")
			buyerID, err := strconv.Atoi(callbackData[1])
			if err != nil {
				return true, err
			}

			reportType := callbackData[2]
			ctgInfo, err := s.storage.GetCtgInfoFromName(ctx, reportType)
			if err != nil {
				return true, err
			}

			if _, err = s.storage.CheckIfWorkerExistAndAdd(ctx, msg.UserID, msg.UserDisplayName); err != nil {
				return true, err
			}

			if succsessful, err := s.storage.CreateTicket(ctx, msg.UserID, int64(buyerID), ctgInfo["id"].(int64)); err != nil || !succsessful {
				s.tgClient.SendMessage(TxtBusyWorker, msg.UserID)
				return true, err
			}

			s.lastInlineKbMsg[msg.UserID], err = s.tgClient.ShowInlineButtons(
				fmt.Sprintf(TxtToWorker, reportType, s.lastUserInteraction[int64(buyerID)].ticket),
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
			ticketInfo, err := s.storage.GetTicketInfo(ctx, msg.UserID)
			if err != nil {
				s.tgClient.SendMessage(TxtErrorTicketUpd, msg.UserID)
				return true, err
			}

			if err := s.storage.UpdateTicketStatus(ctx, msg.UserID, "bad"); err != nil {
				s.tgClient.SendMessage(TxtErrorTicketUpd, msg.UserID)
				return true, err
			}

			if _, err := s.storage.ChangeWorkerStatus(ctx, msg.UserID, false); err != nil {
				s.tgClient.SendMessage(TxtErrorTicketUpd, msg.UserID)
				return true, err
			}

			buyerID, ok := ticketInfo["buyer_tg_id"].(int64)
			if !ok {
				s.tgClient.SendMessage(TxtErrorTicketUpd, msg.UserID)
				return true, err
			}

			ctgID, ok := ticketInfo["category_id"].(int64)
			if !ok {
				s.tgClient.SendMessage(TxtErrorTicketUpd, msg.UserID)
				return true, errors.New("Ошибка при конвертации типов")
			}

			ctgInfo, err := s.storage.GetCategoryInfo(ctx, ctgID)
			if err != nil {
				s.tgClient.SendMessage(TxtErrorTicketUpd, msg.UserID)
				return true, err
			}

			ctgName, ok := ctgInfo["name"].(string)
			if !ok {
				s.tgClient.SendMessage(TxtErrorTicketUpd, msg.UserID)
				return true, errors.New("Ошибка при конвертации типов")
			}

			ctgPrice, ok := ctgInfo["price"].(float64)
			if !ok {
				s.tgClient.SendMessage(TxtErrorTicketUpd, msg.UserID)
				return true, errors.New("Ошибка при конвертации типов")
			}

			if _, err := s.tgClient.SendMessage(
				fmt.Sprintf(TxtBadTicketUsr, ctgName, ctgPrice),
				msg.UserID,
			); err != nil {
				s.tgClient.SendMessage(TxtErrorTicketUpd, msg.UserID)
				return true, err
			}

			if err := s.storage.AddUserLimit(ctx, buyerID, ctgPrice); err != nil {
				s.tgClient.SendMessage(TxtErrorTicketUpd, msg.UserID)
				return true, err
			}

			return true, s.tgClient.DeleteInlineButtons(msg.UserID, msg.CallbackMsgID, TxtBadTicket)

		} else if msg.Text == "goodTicket" {
			s.lastUserInteraction[msg.UserID].command = "goodTicket"

			return true, s.tgClient.DeleteInlineButtons(msg.UserID, msg.CallbackMsgID, TxtSendFile)

		}
	}

	// Команда не распознана.
	return false, nil
}

func getPages(ctx context.Context, s *Model, msg Message, from int, orders []bottypes.UserDataRecord) (string, error) {
	var (
		txtOrders strings.Builder
		ctgsInfo  = make(map[int64]map[string]any)
		err       error
	)

	txtOrders.WriteString(fmt.Sprintf("Page: %v\n", from/OrdersInPage+1))

	for _, order := range orders[from:min(from+OrdersInPage, len(orders))] {
		if ctgsInfo[order.CategoryID] == nil {
			ctgsInfo[order.CategoryID], err = s.storage.GetCategoryInfo(ctx, order.CategoryID)
			if err != nil {
				s.tgClient.SendMessage(TxtErrorTicketUpd, msg.UserID)
				return "", err
			}
		}
		ctgInfo := ctgsInfo[order.CategoryID]

		ctgName, ok := ctgInfo["name"].(string)
		if !ok {
			s.tgClient.SendMessage(TxtErrorTicketUpd, msg.UserID)
			return "", errors.New("Ошибка при конвертации типов")
		}

		ctgPrice, ok := ctgInfo["price"].(float64)
		if !ok {
			s.tgClient.SendMessage(TxtErrorTicketUpd, msg.UserID)
			return "", errors.New("Ошибка при конвертации типов")
		}

		txtOrders.WriteString(fmt.Sprintf(TxtOrderHistory, order.RecordID, order.Period, ctgName, ctgPrice))
	}

	return txtOrders.String(), nil
}

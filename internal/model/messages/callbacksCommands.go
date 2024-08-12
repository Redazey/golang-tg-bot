package messages

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	types "tgssn/internal/model/bottypes"
	"tgssn/pkg/cache"
	"tgssn/pkg/errors"
	"tgssn/pkg/logger"

	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

// callbacks
func CallbacksCommands(s *Model, msg Message, paymentCtgs []string, paymentCtgsInfo []types.CtgInfo) (bool, error) {
	cacheInlinekbMsg, err := cache.ReadCache(fmt.Sprintf("%v_inlinekbMsg", msg.UserID))
	if err != nil {
		return true, err
	}

	lastInlinekbMsg, err := strconv.Atoi(cacheInlinekbMsg)
	if err != nil {
		return true, err
	}

	if msg.IsCallback {
		span, ctx := opentracing.StartSpanFromContext(s.ctx, "callbacksCommands")
		s.ctx = ctx
		defer span.Finish()
		var err error

		// Дерево callbacks начинающихся с Categories
		if msg.Text == "backToCtg" {
			var btns []types.TgRowButtons
			for _, ctg := range paymentCtgs {
				btns = append(btns, types.TgRowButtons{types.TgInlineButton{DisplayName: ctg, Value: ctg}})
			}

			if lastInlinekbMsg == 0 {
				lastMsgID, err := s.tgClient.ShowInlineButtons(TxtCtgs, btns, msg.UserID)
				if err != nil {
					return true, err
				}

				if err := cache.SaveCache(fmt.Sprintf("%v_inlinekbMsg", msg.UserID), lastMsgID); err != nil {
					return true, err
				}
			}

			return true, s.tgClient.EditInlineButtons(
				TxtCtgs,
				lastInlinekbMsg,
				msg.UserID,
				btns,
			)

			// выбор категории товара

		} else if strings.Contains(strings.Join(paymentCtgs, " "), msg.Text) && !strings.Contains(msg.Text, "buy") {
			for _, ctg := range paymentCtgsInfo {
				if msg.Text != ctg.Short {
					continue
				}

				BtnBuying[0][0] = types.TgInlineButton{
					DisplayName: fmt.Sprintf(TxtBtnBuy, ctg.Price),
					Value:       "buy " + ctg.Short,
				}

				if lastInlinekbMsg == 0 {
					lastMsgID, err := s.tgClient.ShowInlineButtons(
						fmt.Sprintf(TxtReports, ctg.Name, ctg.Description),
						BtnBuying,
						msg.UserID,
					)
					if err != nil {
						return true, err
					}

					if err := cache.SaveCache(fmt.Sprintf("%v_inlinekbMsg", msg.UserID), lastMsgID); err != nil {
						return true, err
					}
				}

				return true, s.tgClient.EditInlineButtons(
					fmt.Sprintf(TxtReports, ctg.Name, ctg.Description),
					lastInlinekbMsg,
					msg.UserID,
					BtnBuying,
				)
			}

			// покупка товара

		} else if strings.Contains(msg.Text, "buy") {
			for _, ctg := range paymentCtgsInfo {
				if msg.Text != "buy "+ctg.Short {
					continue
				}
				if msg.Text == "buy Fullz" {
					s.tgClient.DeleteInlineButtons(msg.UserID, lastInlinekbMsg, TxtFullzPaymentDesc)
					return true, nil
				}

				if err := cache.SaveCache(fmt.Sprintf("%v_command", msg.UserID), "buy "+ctg.Short); err != nil {
					return true, err
				}

				if lastInlinekbMsg == 0 {
					lastMsgID, err := s.tgClient.ShowInlineButtons(
						TxtPaymentDesc,
						[]types.TgRowButtons{{BackToCtgBtn}},
						msg.UserID,
					)
					if err != nil {
						return true, err
					}

					if err := cache.SaveCache(fmt.Sprintf("%v_inlinekbMsg", msg.UserID), lastMsgID); err != nil {
						return true, err
					}
				}
				return true, s.tgClient.EditInlineButtons(
					fmt.Sprintf(TxtPaymentDesc, ctg.DataFormat),
					lastInlinekbMsg,
					msg.UserID,
					[]types.TgRowButtons{{BackToCtgBtn}},
				)
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

			if lastInlinekbMsg == 0 {
				lastMsgID, err := s.tgClient.ShowInlineButtons(
					fmt.Sprintf(TxtProfile, msg.UserID, balance, orders),
					BtnProfile,
					msg.UserID,
				)
				if err != nil {
					return true, err
				}

				if err := cache.SaveCache(fmt.Sprintf("%v_inlinekbMsg", msg.UserID), lastMsgID); err != nil {
					return true, err
				}
			}

			return true, s.tgClient.EditInlineButtons(
				fmt.Sprintf(TxtProfile, msg.UserID, balance, orders),
				lastInlinekbMsg,
				msg.UserID,
				BtnProfile,
			)

		} else if msg.Text == "refill" {
			if err := cache.SaveCache(fmt.Sprintf("%v_command", msg.UserID), "refill"); err != nil {
				return true, err
			}

			if lastInlinekbMsg == 0 {
				lastMsgID, err := s.tgClient.ShowInlineButtons(
					TxtPaymentQuestion,
					[]types.TgRowButtons{{BackToProfileBtn}},
					msg.UserID,
				)
				if err != nil {
					return true, err
				}

				if err := cache.SaveCache(fmt.Sprintf("%v_inlinekbMsg", msg.UserID), lastMsgID); err != nil {
					return true, err
				}
			}
			return true, s.tgClient.EditInlineButtons(
				TxtPaymentQuestion,
				lastInlinekbMsg,
				msg.UserID,
				[]types.TgRowButtons{{BackToProfileBtn}},
			)

		} else if msg.Text == "orders" {
			var orders []types.UserDataRecord
			if err := cache.ReadMapCache(fmt.Sprintf("%v_orders", msg.UserID), &orders); err != nil {
				return true, err
			}
			if orders == nil {
				orders, err = s.storage.GetUserOrders(ctx, msg.UserID)
				if err != nil {
					return true, err
				}

				if err := cache.SaveMapCache(fmt.Sprintf("%v_orders", msg.UserID), &orders); err != nil {
					return true, err
				}
			}

			var btns []types.TgRowButtons

			if err := cache.SaveCache(fmt.Sprintf("%v_orderPage", msg.UserID), 0); err != nil {
				return true, err
			}

			txtPage, err := getOrdersPages(0, orders, paymentCtgsInfo)
			if err != nil {
				return true, err
			}

			if OrdersInPage >= len(orders) {
				btns = []types.TgRowButtons{{BackToProfileBtn}}
			} else {
				btns = []types.TgRowButtons{{BtnOrderForward}, {BackToProfileBtn}}
			}

			if lastInlinekbMsg == 0 {
				lastMsgID, err := s.tgClient.ShowInlineButtons(
					TxtPaymentDesc,
					[]types.TgRowButtons{{BackToCtgBtn}},
					msg.UserID,
				)
				if err != nil {
					return true, err
				}

				if err := cache.SaveCache(fmt.Sprintf("%v_inlinekbMsg", msg.UserID), lastMsgID); err != nil {
					return true, err
				}
			}
			return true, s.tgClient.EditInlineButtons(
				txtPage,
				lastInlinekbMsg,
				msg.UserID,
				btns,
			)

		} else if msg.Text == "refills" {
			var refills []types.UserRefillRecord
			if err := cache.ReadMapCache(fmt.Sprintf("%v_orders", msg.UserID), &refills); err != nil {
				return true, err
			}

			if refills == nil {
				refills, err := s.storage.GetRefillHistory(ctx, msg.UserID)
				if err != nil {
					return true, err
				}

				if err := cache.SaveMapCache(fmt.Sprintf("%v_orders", msg.UserID), &refills); err != nil {
					return true, err
				}
			}

			var btns []types.TgRowButtons

			if err := cache.SaveCache(fmt.Sprintf("%v_refillPage", msg.UserID), 0); err != nil {
				return true, err
			}

			txtPage := getRefillPages(0, refills)

			if OrdersInPage >= len(refills) {
				btns = []types.TgRowButtons{{BackToProfileBtn}}
			} else {
				btns = []types.TgRowButtons{{BtnRefillForward}, {BackToProfileBtn}}
			}

			if lastInlinekbMsg == 0 {
				lastMsgID, err := s.tgClient.ShowInlineButtons(
					txtPage,
					btns,
					msg.UserID,
				)
				if err != nil {
					return true, err
				}

				if err := cache.SaveCache(fmt.Sprintf("%v_inlinekbMsg", msg.UserID), lastMsgID); err != nil {
					return true, err
				}
			}
			return true, s.tgClient.EditInlineButtons(
				txtPage,
				lastInlinekbMsg,
				msg.UserID,
				btns,
			)

		} else if msg.Text == "deleteInvoice" {
			cachePaymentID, err := cache.ReadCache(fmt.Sprintf("%v_paymentID", msg.UserID))
			if err != nil {
				return true, err
			}

			paymentID, err := strconv.Atoi(cachePaymentID)
			if err != nil {
				return true, err
			}

			body, err := s.payment.CryptoPayRequest(
				s.ctx, "deleteInvoice",
				DeleteInvoiceRequest{InvoiceID: int64(paymentID)},
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

			if err = s.storage.DeleteRefillRecord(ctx, int64(paymentID)); err != nil {
				return true, err
			}

			if lastInlinekbMsg == 0 {
				lastMsgID, err := s.tgClient.ShowInlineButtons(
					TxtPaymentCanceled,
					[]types.TgRowButtons{{BackToProfileBtn}},
					msg.UserID,
				)
				if err != nil {
					return true, err
				}

				if err := cache.SaveCache(fmt.Sprintf("%v_inlinekbMsg", msg.UserID), lastMsgID); err != nil {
					return true, err
				}
			}
			return true, s.tgClient.EditInlineButtons(
				TxtPaymentCanceled,
				lastInlinekbMsg,
				msg.UserID,
				[]types.TgRowButtons{{BackToProfileBtn}},
			)

		} else if strings.Contains(strings.Join(PaymentMethods, " "), msg.Text) {
			cacheAmount, err := cache.ReadCache(fmt.Sprintf("%v_amount", msg.UserID))
			if err != nil {
				return true, err
			}

			amount, err := strconv.Atoi(cacheAmount)
			if err != nil {
				return true, err
			}

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
					lastInlinekbMsg,
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
				Amount:       float64(amount),
				Description:  fmt.Sprintf(TxtRefillDesc, amount, msg.Text),
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

			if err := cache.SaveCache(fmt.Sprintf("%v_paymentID", msg.UserID), paymentState.InvoiceID); err != nil {
				return true, err
			}
			BtnRefillRequest[0][0].URL = paymentState.BotInvoiceURL

			if err = s.storage.InsertUserRefillRecord(ctx, msg.UserID, paymentState.InvoiceID, float64(amount)); err != nil {
				return true, err
			}

			lastMsgID, err := s.tgClient.ShowInlineButtons(
				TxtRefillReqCreated,
				BtnRefillRequest,
				msg.UserID,
			)
			if err != nil {
				return true, err
			}

			if err := cache.SaveCache(fmt.Sprintf("%v_inlinekbMsg", msg.UserID), lastMsgID); err != nil {
				return true, err
			}

			return true, nil

		} else if msg.Text == "pageBack" {
			cacheOrderPage, err := cache.ReadCache(fmt.Sprintf("%v_orderPage", msg.UserID))
			if err != nil {
				return true, err
			}

			orderPage, err := strconv.Atoi(cacheOrderPage)
			if err != nil {
				return true, err
			}

			var orders []types.UserDataRecord
			if err = cache.ReadMapCache(fmt.Sprintf("%v_orderPage", msg.UserID), &orders); err != nil {
				return true, err
			}

			if orders == nil {
				orders, err = s.storage.GetUserOrders(ctx, msg.UserID)
				if err != nil {
					return true, err
				}

				if err := cache.SaveMapCache(fmt.Sprintf("%v_orders", msg.UserID), &orders); err != nil {
					return true, err
				}
			}

			if err := cache.SaveCache(fmt.Sprintf("%v_orderPage", msg.UserID), orderPage-OrdersInPage); err != nil {
				return true, err
			}

			var btnOrders []types.TgRowButtons

			txtPage, err := getOrdersPages(orderPage, orders, paymentCtgsInfo)
			if err != nil {
				return true, err
			}

			if orderPage == 0 {
				btnOrders = []types.TgRowButtons{{BtnOrderForward}, {BackToProfileBtn}}
			} else if orderPage <= len(orders) {
				btnOrders = []types.TgRowButtons{{BackToProfileBtn}}
			} else {
				btnOrders = []types.TgRowButtons{{BtnOrderBack, BtnOrderForward}, {BackToProfileBtn}}
			}

			if lastInlinekbMsg == 0 {
				lastMsgID, err := s.tgClient.ShowInlineButtons(
					txtPage,
					btnOrders,
					msg.UserID,
				)
				if err != nil {
					return true, err
				}

				if err := cache.SaveCache(fmt.Sprintf("%v_inlinekbMsg", msg.UserID), lastMsgID); err != nil {
					return true, err
				}
			}
			return true, s.tgClient.EditInlineButtons(
				txtPage,
				lastInlinekbMsg,
				msg.UserID,
				btnOrders,
			)

		} else if msg.Text == "pageForward" {
			cacheOrderPage, err := cache.ReadCache(fmt.Sprintf("%v_orderPage", msg.UserID))
			if err != nil {
				return true, err
			}

			orderPage, err := strconv.Atoi(cacheOrderPage)
			if err != nil {
				return true, err
			}

			var orders []types.UserDataRecord
			if err = cache.ReadMapCache(fmt.Sprintf("%v_orderPage", msg.UserID), &orders); err != nil {
				return true, err
			}

			if orders == nil {
				orders, err = s.storage.GetUserOrders(ctx, msg.UserID)
				if err != nil {
					return true, err
				}

				if err := cache.SaveMapCache(fmt.Sprintf("%v_orders", msg.UserID), &orders); err != nil {
					return true, err
				}
			}

			if err := cache.SaveCache(fmt.Sprintf("%v_orderPage", msg.UserID), orderPage-OrdersInPage); err != nil {
				return true, err
			}

			var btnOrders []types.TgRowButtons

			txtPage, err := getOrdersPages(orderPage, orders, paymentCtgsInfo)
			if err != nil {
				return true, err
			}

			if orderPage+OrdersInPage >= len(orders) {
				btnOrders = []types.TgRowButtons{{BtnOrderBack}, {BackToProfileBtn}}
			} else {
				btnOrders = []types.TgRowButtons{{BtnOrderBack, BtnOrderForward}, {BackToProfileBtn}}
			}

			if lastInlinekbMsg == 0 {
				lastMsgID, err := s.tgClient.ShowInlineButtons(
					txtPage,
					btnOrders,
					msg.UserID,
				)
				if err != nil {
					return true, err
				}

				if err := cache.SaveCache(fmt.Sprintf("%v_inlinekbMsg", msg.UserID), lastMsgID); err != nil {
					return true, err
				}
			}
			return true, s.tgClient.EditInlineButtons(
				txtPage,
				lastInlinekbMsg,
				msg.UserID,
				btnOrders,
			)

		} else if msg.Text == "refillPageBack" {
			cacheRefillPage, err := cache.ReadCache(fmt.Sprintf("%v_refillPage", msg.UserID))
			if err != nil {
				return true, err
			}

			refillPage, err := strconv.Atoi(cacheRefillPage)
			if err != nil {
				return true, err
			}

			var refills []types.UserRefillRecord
			if err = cache.ReadMapCache(fmt.Sprintf("%v_refills", msg.UserID), &refills); err != nil {
				return true, err
			}

			if refills == nil {
				refills, err = s.storage.GetRefillHistory(ctx, msg.UserID)
				if err != nil {
					return true, err
				}

				if err := cache.SaveMapCache(fmt.Sprintf("%v_refills", msg.UserID), &refills); err != nil {
					return true, err
				}
			}

			if err := cache.SaveCache(fmt.Sprintf("%v_refillPage", msg.UserID), refillPage-OrdersInPage); err != nil {
				return true, err
			}

			var btnOrders []types.TgRowButtons

			txtPage := getRefillPages(refillPage, refills)

			if refillPage == 0 {
				btnOrders = []types.TgRowButtons{{BtnRefillForward}, {BackToProfileBtn}}
			} else {
				btnOrders = []types.TgRowButtons{{BtnRefillBack, BtnRefillForward}, {BackToProfileBtn}}
			}

			if lastInlinekbMsg == 0 {
				lastMsgID, err := s.tgClient.ShowInlineButtons(
					txtPage,
					btnOrders,
					msg.UserID,
				)
				if err != nil {
					return true, err
				}

				if err := cache.SaveCache(fmt.Sprintf("%v_inlinekbMsg", msg.UserID), lastMsgID); err != nil {
					return true, err
				}
			}
			return true, s.tgClient.EditInlineButtons(
				txtPage,
				lastInlinekbMsg,
				msg.UserID,
				btnOrders,
			)

		} else if msg.Text == "refillPageForward" {
			cacheRefillPage, err := cache.ReadCache(fmt.Sprintf("%v_refillPage", msg.UserID))
			if err != nil {
				return true, err
			}

			refillPage, err := strconv.Atoi(cacheRefillPage)
			if err != nil {
				return true, err
			}

			var refills []types.UserRefillRecord
			if err = cache.ReadMapCache(fmt.Sprintf("%v_refills", msg.UserID), &refills); err != nil {
				return true, err
			}

			if refills == nil {
				refills, err = s.storage.GetRefillHistory(ctx, msg.UserID)
				if err != nil {
					return true, err
				}

				if err := cache.SaveMapCache(fmt.Sprintf("%v_refills", msg.UserID), &refills); err != nil {
					return true, err
				}
			}

			if err := cache.SaveCache(fmt.Sprintf("%v_refillPage", msg.UserID), refillPage-OrdersInPage); err != nil {
				return true, err
			}

			var btnOrders []types.TgRowButtons

			txtPage := getRefillPages(refillPage, refills)

			if refillPage+OrdersInPage >= len(refills) {
				btnOrders = []types.TgRowButtons{{BtnRefillBack}, {BackToProfileBtn}}
			} else {
				btnOrders = []types.TgRowButtons{{BtnRefillBack, BtnRefillForward}, {BackToProfileBtn}}
			}

			if lastInlinekbMsg == 0 {
				lastMsgID, err := s.tgClient.ShowInlineButtons(
					txtPage,
					btnOrders,
					msg.UserID,
				)
				if err != nil {
					return true, err
				}

				if err := cache.SaveCache(fmt.Sprintf("%v_inlinekbMsg", msg.UserID), lastMsgID); err != nil {
					return true, err
				}
			}
			return true, s.tgClient.EditInlineButtons(
				txtPage,
				lastInlinekbMsg,
				msg.UserID,
				btnOrders,
			)

			// Раздел Callback'ов, которые отправляют работники
		} else if strings.Contains(msg.Text, "takeTicket:") {
			callbackData := strings.Split(msg.Text, ":")
			reportType := callbackData[2]
			buyerID, err := strconv.Atoi(callbackData[1])
			if err != nil {
				return true, err
			}

			ctgInfo := GetCtgInfoFromName(reportType, paymentCtgsInfo)

			if _, err = s.storage.CheckIfWorkerExistAndAdd(ctx, msg.UserID, msg.UserDisplayName); err != nil {
				return true, err
			}

			if succsessful, err := s.storage.CreateTicket(ctx, msg.UserID, int64(buyerID), ctgInfo.ID); err != nil || !succsessful {
				s.tgClient.SendMessage(TxtBusyWorker, msg.UserID)
				return true, err
			}

			ticketData, err := cache.ReadCache(fmt.Sprintf("%v_ticket", msg.UserID))
			if err != nil {
				return true, err
			}

			lastMsgID, err := s.tgClient.ShowInlineButtons(
				fmt.Sprintf(TxtToWorker, reportType, ticketData),
				BtnToWorker,
				msg.UserID,
			)
			if err != nil {
				return true, err
			}

			if err := cache.SaveCache(fmt.Sprintf("%v_inlinekbMsg", msg.UserID), lastMsgID); err != nil {
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

			ctgInfo := GetCtgInfoFromID(ctgID, paymentCtgsInfo)

			if _, err := s.tgClient.SendMessage(
				fmt.Sprintf(TxtBadTicketUsr, ctgInfo.Name, ctgInfo.Price),
				msg.UserID,
			); err != nil {
				s.tgClient.SendMessage(TxtErrorTicketUpd, msg.UserID)
				return true, err
			}

			if err := s.storage.AddUserLimit(ctx, buyerID, ctgInfo.Price); err != nil {
				s.tgClient.SendMessage(TxtErrorTicketUpd, msg.UserID)
				return true, err
			}

			return true, s.tgClient.DeleteInlineButtons(msg.UserID, msg.CallbackMsgID, TxtBadTicket)

		} else if msg.Text == "goodTicket" {
			if err := cache.SaveCache(fmt.Sprintf("%v_command", msg.UserID), "goodTicket"); err != nil {
				return true, err
			}

			return true, s.tgClient.DeleteInlineButtons(msg.UserID, msg.CallbackMsgID, TxtSendFile)
		}
	}

	// Команда не распознана.
	return false, nil
}

func getOrdersPages(from int, orders []types.UserDataRecord, ctgs []types.CtgInfo) (string, error) {
	var txtOrders strings.Builder

	txtOrders.WriteString(fmt.Sprintf("Page: %v\n", from/OrdersInPage+1))

	for _, order := range orders[from:min(from+OrdersInPage, len(orders))] {
		ctgInfo := GetCtgInfoFromID(order.CategoryID, ctgs)

		txtOrders.WriteString(fmt.Sprintf(TxtOrderHistory, order.RecordID, order.Period, ctgInfo.Name, ctgInfo.Price))
	}

	return txtOrders.String(), nil
}

func getRefillPages(from int, refills []types.UserRefillRecord) string {
	var txtRefills strings.Builder

	txtRefills.WriteString(fmt.Sprintf("Page: %v\n", from/OrdersInPage+1))

	for _, refill := range refills[from:min(from+OrdersInPage, len(refills))] {
		txtRefills.WriteString(fmt.Sprintf(TxtRefillsHistory, refill.InvoiceID, refill.Period, refill.Amount))
	}

	return txtRefills.String()
}

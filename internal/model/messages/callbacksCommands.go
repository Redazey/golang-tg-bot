package messages

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"tgssn/internal/model/bottypes"
	"tgssn/pkg/errors"

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
				s.tgClient.EditInlineButtons(TxtPaymentDesc, s.lastInlineKbMsg[msg.UserID], msg.UserID, BackToCtgBtn)
				s.lastUserInteraction[msg.UserID].command = "buy TU"
				return true, nil

			} else if msg.Text == "buy CR" {
				s.tgClient.EditInlineButtons(TxtPaymentDesc, s.lastInlineKbMsg[msg.UserID], msg.UserID, BackToCtgBtn)
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
			return true, s.tgClient.EditInlineButtons(TxtPaymentQuestion, s.lastInlineKbMsg[msg.UserID], msg.UserID, BackToProfileBtn)

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
				[]bottypes.TgRowButtons{{BtnOrderForward}, {BackToProfileBtn[0][0]}},
			)

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
				btnOrders = []bottypes.TgRowButtons{{BtnOrderForward}, {BackToProfileBtn[0][0]}}
			} else {
				btnOrders = []bottypes.TgRowButtons{{BtnOrderBack, BtnOrderForward}, {BackToProfileBtn[0][0]}}
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
				btnOrders = []bottypes.TgRowButtons{{BtnOrderBack}, {BackToProfileBtn[0][0]}}
			} else {
				btnOrders = []bottypes.TgRowButtons{{BtnOrderBack, BtnOrderForward}, {BackToProfileBtn[0][0]}}
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

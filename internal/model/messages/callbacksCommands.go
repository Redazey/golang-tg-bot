package messages

import (
	"encoding/json"
	"fmt"
	"strconv"
	types "tgseller/internal/model/bottypes"
	"tgseller/pkg/cache"
	"tgseller/pkg/errors"
	"tgseller/pkg/logger"

	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

// callbacks
func CallbacksCommands(s *Model, msg Message) (bool, error) {
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

		switch msg.Text {
		case "backToCtg":
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
		case "buy":
			if err := cache.SaveCache(fmt.Sprintf("%v_command", msg.UserID), "buy"); err != nil {
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
				fmt.Sprintf(TxtPaymentDesc),
				lastInlinekbMsg,
				msg.UserID,
				[]types.TgRowButtons{{BackToCtgBtn}},
			)
		case "backToProfile":
			if _, err = s.storage.CheckIfUserExistAndAdd(ctx, msg.UserID); err != nil {
				return true, err
			}

			access, err := s.storage.GetUserAccessStatus(ctx, msg.UserID)
			if err != nil {
				return true, err
			}

			var access_status string
			if access {
				access_status = "активна!"
			} else {
				access_status = "неактивна"
			}

			if lastInlinekbMsg == 0 {
				lastMsgID, err := s.tgClient.ShowInlineButtons(
					fmt.Sprintf(TxtProfile, msg.UserID, access_status),
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
				fmt.Sprintf(TxtProfile, msg.UserID, access_status),
				lastInlinekbMsg,
				msg.UserID,
				BtnProfile,
			)
		case "confirm_buy":
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
		case "deleteInvoice":
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
		}
	}
	// Команда не распознана.
	return false, nil
}

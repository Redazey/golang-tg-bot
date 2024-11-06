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
				lastMsgID, err := s.tgClient.ShowInlineButtons(
					TxtStart,
					[]types.TgRowButtons{
						{
							BtnSubscribe,
						},
					}, msg.UserID)
				if err != nil {
					return true, err
				}

				if err := cache.SaveCache(fmt.Sprintf("%v_inlinekbMsg", msg.UserID), lastMsgID); err != nil {
					return true, err
				}
			}

			return true, s.tgClient.EditInlineButtons(
				TxtStart,
				lastInlinekbMsg,
				msg.UserID,
				[]types.TgRowButtons{
					{
						BtnSubscribe,
					},
				},
			)
		case "backToProfile":
			if _, err = s.storage.CheckIfUserExistAndAdd(ctx, msg.UserID); err != nil {
				return true, err
			}

			access, err := s.storage.GetUserAccessStatus(ctx, msg.UserID)
			if err != nil {
				return true, err
			}

			var accessed_at string
			var access_status string
			if access {
				access_status = "активна"
				accessed_at_time, err := s.storage.GetUserAccessData(s.ctx, msg.UserID)
				accessed_at = accessed_at_time.Format("2006 02 January")
				if err != nil {
					return true, err
				}
			} else {
				access_status = "неактивна"
				accessed_at = "\\-"
			}

			if lastInlinekbMsg == 0 {
				lastMsgID, err := s.tgClient.ShowInlineButtons(
					fmt.Sprintf(TxtProfile, msg.UserID, access_status, accessed_at),
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
				fmt.Sprintf(TxtProfile, msg.UserID, access_status, accessed_at),
				lastInlinekbMsg,
				msg.UserID,
				BtnProfile,
			)
		case "buy":
			invoiceReq := CreateInvoiceRequest{
				CurrencyType: "fiat",
				Asset:        "USDT",
				Fiat:         "USD",
				Amount:       constAmount,
				Description:  fmt.Sprintf(TxtRefillDesc, 2),
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

			record := types.Records{
				ID:      paymentState.InvoiceID,
				User_id: msg.UserID,
				Amount:  constAmount,
			}

			if succeed, err := s.storage.InsertUserDataRecord(ctx, msg.UserID, record); err != nil || !succeed {
				return true, err
			}

			BtnRefillRequest[0][0].URL = paymentState.BotInvoiceURL

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

			if err = s.storage.DeleteUserRecord(ctx, int64(paymentID)); err != nil {
				return true, err
			}

			if lastInlinekbMsg == 0 {
				lastMsgID, err := s.tgClient.ShowInlineButtons(
					TxtPaymentCanceled,
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
				TxtPaymentCanceled,
				lastInlinekbMsg,
				msg.UserID,
				[]types.TgRowButtons{{BackToCtgBtn}},
			)
		}
	}
	// Команда не распознана.
	return false, nil
}

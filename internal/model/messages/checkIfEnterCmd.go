package messages

import (
	"fmt"
	"strconv"
	"strings"
	types "tgseller/internal/model/bottypes"
	"tgseller/pkg/cache"
	"tgseller/pkg/errors"

	"github.com/opentracing/opentracing-go"
)

func CheckIfEnterCmd(s *Model, msg Message, paymentCtgs []string, lastUserCommand string, paymentCtgsInfo []types.CtgInfo) (bool, error) {
	cacheInlinekbMsg, err := cache.ReadCache(fmt.Sprintf("%v_inlinekbMsg", msg.UserID))
	if err != nil {
		return true, err
	}

	lastInlinekbMsg, err := strconv.Atoi(cacheInlinekbMsg)
	if err != nil {
		return true, err
	}

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

			if err = cache.SaveCache(fmt.Sprintf("%v_amount", msg.UserID), float64(userInput)); err != nil {
				return true, err
			}

			var btns []types.TgRowButtons
			btns = append(btns, types.TgRowButtons{BackToProfileBtn})
			for _, method := range PaymentMethods {
				btns = append(btns, types.TgRowButtons{types.TgInlineButton{DisplayName: method, Value: method}})
			}

			if lastInlinekbMsg == 0 {
				s.tgClient.ShowInlineButtons(TxtPaymentQuestion, btns, msg.UserID)
			}
			return true, s.tgClient.EditInlineButtons(TxtPaymentQuestion, lastInlinekbMsg, msg.UserID, btns)

		} else if strings.Contains(lastUserCommand, "buy") {
			var err error

			for _, ctg := range paymentCtgsInfo {
				if lastUserCommand != "buy "+ctg.Short {
					continue
				}

				if succsessful, err := s.storage.InsertUserDataRecord(ctx, msg.UserID, ctg); err != nil {
					s.tgClient.SendMessage(TxtError, msg.UserID)
					return true, err

				} else if !succsessful {
					if lastInlinekbMsg, err = s.tgClient.ShowInlineButtons(
						TxtPaymentNotEnough,
						BtnRefill,
						msg.UserID,
					); err != nil {
						return true, err
					}

					if err := cache.SaveCache(fmt.Sprintf("%v_inlinekbMsg", msg.UserID), lastInlinekbMsg); err != nil {
						return true, err
					}

					return true, nil
				}

				if isValidDataInput(msg.Text) {
					if err := cache.SaveCache(fmt.Sprintf("%v_ticket", msg.UserID), lastInlinekbMsg); err != nil {
						return true, err
					}
				} else {
					s.tgClient.SendMessage(TxtWrongTicketFormat, msg.UserID)
					return true, nil
				}

				lastInlinekbMsg, err = s.tgClient.ShowInlineButtons(
					fmt.Sprintf(TxtForWorkers, ctg.Name),
					CreateInlineButtons(
						BtnWorkersChatDN,
						fmt.Sprintf(BtnWorkersChatVal, msg.UserID, ctg.Name),
					),
					WorkersChatID,
				)
				if err != nil {
					return true, err
				}

				if err := cache.SaveCache(fmt.Sprintf("%v_inlinekbMsg", msg.UserID), lastInlinekbMsg); err != nil {
					return true, err
				}

				if err = s.storage.AddUserLimit(ctx, msg.UserID, -ctg.Price); err != nil {
					return true, err
				}

				_, err = s.tgClient.SendMessage(TxtTicketInProccess, msg.UserID)

				return true, err
			}

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
	if len(parts) == 7 {
		if parts[0] == "" || parts[0] == " " {
			return false
		} else if parts[1] == "" || parts[1] == " " {
			return false
		} else if parts[2] == "" || parts[2] == " " {
			return false
		} else if parts[3] == "" || parts[3] == " " {
			return false
		} else if _, ok := strconv.Atoi(parts[4]); ok != nil || parts[4] == "" || parts[4] == " " {
			return false
		} else if parts[5] == "" || parts[5] == " " {
			return false
		} else if _, ok := strconv.Atoi(parts[6]); ok != nil || parts[6] == "" || parts[6] == " " {
			return false
		}
	} else if len(parts) == 5 {
		if parts[0] == "" || parts[0] == " " {
			return false
		} else if parts[1] == "" || parts[1] == " " {
			return false
		} else if parts[2] == "" || parts[2] == " " {
			return false
		} else if parts[3] == "" || parts[3] == " " {
			return false
		} else if _, ok := strconv.Atoi(parts[4]); ok != nil || parts[4] == "" || parts[4] == " " {
			return false
		}
	} else {
		return false
	}

	return true
}

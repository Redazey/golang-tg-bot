package messages

import (
	"fmt"
	"tgseller/internal/model/bottypes"
	"tgseller/pkg/cache"

	"github.com/opentracing/opentracing-go"
)

// Распознавание стандартных команд бота.
func CheckBotCommands(s *Model, msg Message) (bool, error) {
	span, ctx := opentracing.StartSpanFromContext(s.ctx, "checkBotCommands")
	s.ctx = ctx
	defer span.Finish()

	switch msg.Text {
	case "/start":
		lastMsgID, err := s.tgClient.ShowKeyboardButtons(
			TxtStart,
			BtnStart,
			msg.UserID,
		)
		if err != nil {
			return true, err
		}
		err = cache.SaveCache(fmt.Sprintf("%v_inlinekbMsg", msg.UserID), lastMsgID)
		if err != nil {
			return true, err
		}

		return true, s.tgClient.EditInlineButtons(
			TxtStart,
			lastMsgID,
			msg.UserID,
			[]bottypes.TgRowButtons{
				{
					BtnSubscribe,
				},
			},
		)
	case "subscribe":
		if err := cache.SaveCache(fmt.Sprintf("%v_command", msg.UserID), "buy"); err != nil {
			return true, err
		}
		lastMsgID, err := s.tgClient.ShowInlineButtons(
			TxtPaymentStart,
			[]bottypes.TgRowButtons{
				{
					BtnSubscribe,
				},
			},
			msg.UserID,
		)
		if err != nil {
			return true, err
		}

		return true, cache.SaveCache(fmt.Sprintf("%v_inlinekbMsg", msg.UserID), lastMsgID)

	case "Profile":
		if _, err := s.storage.CheckIfUserExistAndAdd(ctx, msg.UserID); err != nil {
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

		lastMsgID, err := s.tgClient.ShowInlineButtons(
			fmt.Sprintf(TxtProfile, msg.UserID, access_status, accessed_at),
			BtnProfile,
			msg.UserID,
		)
		if err != nil {
			return true, err
		}

		return true, cache.SaveCache(fmt.Sprintf("%v_inlinekbMsg", msg.UserID), lastMsgID)
	case "/help":
		if err := cache.SaveCache(fmt.Sprintf("%v_inlinekbMsg", msg.UserID), 0); err != nil {
			return true, err
		}

		_, err := s.tgClient.SendMessage(TxtHelp, msg.UserID)
		return true, err
	case "Support":
		if err := cache.SaveCache(fmt.Sprintf("%v_inlinekbMsg", msg.UserID), 0); err != nil {
			return true, err
		}

		_, err := s.tgClient.SendMessage(TxtSupport, msg.UserID)
		return true, err
	}

	// Команда не распознана.
	return false, nil
}

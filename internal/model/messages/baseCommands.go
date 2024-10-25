package messages

import (
	"fmt"
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
		displayName := msg.UserDisplayName
		if len(displayName) == 0 {
			displayName = msg.UserName
		}

		lastMsgID, err := s.tgClient.ShowKeyboardButtons(
			fmt.Sprintf(TxtStart, displayName),
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
			fmt.Sprintf(TxtStart, displayName),
			lastMsgID,
			msg.UserID,
			BtnSubscribe,
		)
	case "Subscribe":
		if err := cache.SaveCache(fmt.Sprintf("%v_command", msg.UserID), "buy"); err != nil {
			return true, err
		}
		lastMsgID, err := s.tgClient.ShowInlineButtons(
			TxtPaymentStart,
			BtnSubscribe,
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

		var access_status string
		if access {
			access_status = "активна!"
		} else {
			access_status = "неактивна"
		}

		lastMsgID, err := s.tgClient.ShowInlineButtons(
			fmt.Sprintf(TxtProfile, msg.UserID, access_status),
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
	}

	// Команда не распознана.
	return false, nil
}

package messages

import (
	"fmt"

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

		if err := s.tgClient.ShowKeyboardButtons(fmt.Sprintf(TxtStart, displayName), BtnStart, msg.UserID); err != nil {
			return true, err
		}

		return true, nil
	case "Categories":
		lastMsgID, err := s.tgClient.ShowInlineButtons(TxtCtgs, BtnCtgs, msg.UserID)
		if err != nil {
			return true, err
		}
		s.lastInlineKbMsg[msg.UserID] = lastMsgID
		return true, nil
	case "Profile":
		if _, err := s.storage.CheckIfUserExistAndAdd(ctx, msg.UserID); err != nil {
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

		lastMsgID, err := s.tgClient.ShowInlineButtons(
			fmt.Sprintf(TxtProfile, msg.UserID, balance, orders),
			BtnProfile,
			msg.UserID,
		)
		if err != nil {
			return true, err
		}
		s.lastInlineKbMsg[msg.UserID] = lastMsgID

		return true, nil
	case "Support":
		s.tgClient.SendMessage(TxtSup, msg.UserID)
		s.lastInlineKbMsg[msg.UserID] = 0
		return true, nil
	case "/help":
		s.lastInlineKbMsg[msg.UserID] = 0
		return true, s.tgClient.SendMessage(TxtHelp, msg.UserID)
	}

	// Команда не распознана.
	return false, nil
}

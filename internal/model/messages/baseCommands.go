package messages

import (
	"fmt"
	types "tgssn/internal/model/bottypes"
	"tgssn/pkg/cache"

	"github.com/opentracing/opentracing-go"
)

// Распознавание стандартных команд бота.
func CheckBotCommands(s *Model, msg Message, paymentCtgs []string) (bool, error) {
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
		var btns []types.TgRowButtons
		for _, ctg := range paymentCtgs {
			btns = append(btns, types.TgRowButtons{types.TgInlineButton{DisplayName: ctg, Value: ctg}})
		}

		lastMsgID, err := s.tgClient.ShowInlineButtons(TxtCtgs, btns, msg.UserID)
		if err != nil {
			return true, err
		}

		if err := cache.SaveCache(fmt.Sprintf("%v_inlinekbMsg", msg.UserID), lastMsgID); err != nil {
			return true, err
		}

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

		if err := cache.SaveCache(fmt.Sprintf("%v_inlinekbMsg", msg.UserID), lastMsgID); err != nil {
			return true, err
		}

		return true, nil
	case "Support":
		s.tgClient.SendMessage(TxtSup, msg.UserID)
		if err := cache.SaveCache(fmt.Sprintf("%v_inlinekbMsg", msg.UserID), 0); err != nil {
			return true, err
		}

		return true, nil
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

package messages

import (
	types "tgseller/internal/model/bottypes"
)

// Команды стартовых действий.
var BtnStart = types.TgKbRowButtons{
	types.TgKeyboardButton{Text: "Profile"},
	types.TgKeyboardButton{Text: "Support"},
}

// Отказ от оплаты
var BackToCtgBtn = types.TgInlineButton{DisplayName: "❌НАЗАД", Value: "backToCtg"}

// Покупка/возвращение назад
var BtnBuying = []types.TgRowButtons{
	{
		types.TgInlineButton{DisplayName: "ЧЕРЕЗ CryptoBot", Value: "buy"},
	},
	{
		BackToCtgBtn,
	},
}

// Кнопка пополнения
var BtnSubscribe = []types.TgRowButtons{
	{
		types.TgInlineButton{DisplayName: "❤️ ПОДПИСАТЬСЯ ❤️", Value: "subscribe"},
		BackToCtgBtn,
	},
}

var BtnRefillRequest = []types.TgRowButtons{
	{
		types.TgInlineButton{DisplayName: "✅ОПЛАТИТЬ", Value: "", URL: "%v"},
	},
	{
		types.TgInlineButton{DisplayName: "❌ОТМЕНИТЬ ОПЛАТУ", Value: "deleteInvoice"},
	},
}

// Кнопки для вкладки профиль
var BtnProfile = []types.TgRowButtons{
	{
		types.TgInlineButton{DisplayName: "❤️ ПРОДЛИТЬ ПОДПИСКУ ❤️", Value: "subscribe"},
	},
}

// Функция для создания кнопок,
// нужна для моментов, когда требуется использовать callback, в котором возвращаются какие то данные
// через fmt.sprintf
func CreateInlineButtons(DisplayName string, value string) []types.TgRowButtons {
	return []types.TgRowButtons{
		{
			types.TgInlineButton{
				DisplayName: DisplayName,
				Value:       value,
			},
		},
	}
}

// Область "Константы и переменные": начало.
var PaymentMethods = []string{"USDT", "TON", "GRAM", "NOT", "MY", "BTC", "LTC", "ETH", "BNB", "TRX", "USDC"}

const (
	TxtStart = `Приветик %v~
	Похоже, что ты зашёл, что бы поглядеть на красивых девочек, однако телеграмм не позволит нам показать их
	красоту просто так, так что заходи в нашу приватку, где мы можем постить ваших любимых девочек без ограничений!❤️
	❤️Напоминаем, что мы собираем только самый лучший контент!❤️
	При возникновении каких то проблем пиши сюда @ForeyDaxel~`
	TxtPaymentStart        = "В наш канал можно попасть, только оплатив вход криптовалютой 👉👈"
	TxtCtgs                = "📰 Choose a category that interests you:"
	TxtBtnBuy              = "buy for %v $"
	TxtProfile             = "📰 ID: %v\n💎 Подписка: %v"
	TxtSup                 = "For assistance, please contact technical support: "
	TxtUnknownCommand      = "Unfortunately, this command is unknown to me. To get started, please enter /start."
	TxtChoosePaymentMethod = "Choose a payment method:"
	TxtRefillDesc          = "Покупка подписки на сумму %v$"
	TxtRefillReqCreated    = "Перейдите по ссылке ниже для оплаты!"
	TxtPaymentSuccsessful  = `Подписка была успешно оформленна!
	Подавайте заявку и она будет автоматически принята в течении минуты :3
	СЮДА ССЫЛКУ`
	TxtPaymentCanceled  = "Оплата была отменена!"
	TxtPaymentErr       = "😱 Произошла ошибка при переводе средств! Перепроверьте введенные данные или обратитесь в поддержку @ForeyDaxel 😱"
	TxtPaymentNotEnough = "You have insufficient funds in your account, please top up"
	TxtError            = "😱 Произошла неожиданная ошибка! Пожалуйста, обратитесь в поддержку! 😱"
	TxtHelp             = "Это бот для приваточки канала Blue Archive, пиши /start и гляди что у нас есть :3"
)

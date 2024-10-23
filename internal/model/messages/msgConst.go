package messages

import (
	types "tgseller/internal/model/bottypes"
)

// Команды стартовых действий.
var BtnStart = types.TgKbRowButtons{
	types.TgKeyboardButton{Text: "Categories"},
	types.TgKeyboardButton{Text: "Profile"},
	types.TgKeyboardButton{Text: "Support"},
}

// Отказ от оплаты
var BackToCtgBtn = types.TgInlineButton{DisplayName: "Back to categories", Value: "backToCtg"}

// Покупка/возвращение назад
var BtnBuying = []types.TgRowButtons{
	{
		types.TgInlineButton{},
	},
	{
		BackToCtgBtn,
	},
}

// Кнопка пополнения
var BtnRefill = []types.TgRowButtons{
	{
		types.TgInlineButton{DisplayName: "Refill balance", Value: "refill"},
		BackToCtgBtn,
	},
}

var BtnRefillRequest = []types.TgRowButtons{
	{
		types.TgInlineButton{DisplayName: "Proceed to payment", Value: "", URL: "%v"},
	},
	{
		types.TgInlineButton{DisplayName: "Cancel request", Value: "deleteInvoice"},
	},
}

// Кнопки для истории заказов
var BtnOrderBack = types.TgInlineButton{DisplayName: "Back", Value: "pageBack"}
var BtnOrderForward = types.TgInlineButton{DisplayName: "Forward", Value: "pageForward"}

// Кнопки для истории пополнений
var BtnRefillBack = types.TgInlineButton{DisplayName: "Back", Value: "refillPageBack"}
var BtnRefillForward = types.TgInlineButton{DisplayName: "Forward", Value: "refillPageForward"}

// Кнопки для вкладки профиль
var BtnProfile = []types.TgRowButtons{
	{
		types.TgInlineButton{DisplayName: "Refill balance", Value: "refill"},
	},
	{
		types.TgInlineButton{DisplayName: "Orders history", Value: "orders"},
		types.TgInlineButton{DisplayName: "Refill history", Value: "refills"},
	},
}

// Возвращение к профилю
var BackToProfileBtn = types.TgInlineButton{DisplayName: "Back to profile", Value: "backToProfile"}

// Кнопки для чата работников

// Сообщение работникам в чате

var (
	BtnWorkersChatDN  = "Взять тикет"
	BtnWorkersChatVal = "takeTicket:%v:%v"
)

// Сообщение работнику в ЛС
var BtnToWorker = []types.TgRowButtons{
	{
		types.TgInlineButton{DisplayName: "❌", Value: "badTicket"},
		types.TgInlineButton{DisplayName: "✅", Value: "goodTicket"},
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
	TxtStart               = "Hello, %v 👋.This is a simple test seller golang-bot"
	TxtCtgs                = "📰 Choose a category that interests you:"
	TxtBtnBuy              = "buy for %v $"
	TxtProfile             = "📰 ID: %v\n💎 Подписка: %v"
	TxtSup                 = "For assistance, please contact technical support: "
	TxtUnknownCommand      = "Unfortunately, this command is unknown to me. To get started, please enter /start."
	TxtChoosePaymentMethod = "Choose a payment method:"
	TxtRefillDesc          = "Top-up of %v $ via %v payment method"
	TxtRefillReqCreated    = "Your top-up request was created!"
	TxtPaymentSuccsessful  = "Account topped up by %v $! 💵"
	TxtPaymentCanceled     = "Payment was succsessfully canceled!"
	TxtPaymentErr          = "An error occurred while transferring funds! Please double-check your information or contact support"
	TxtPaymentNotEnough    = "You have insufficient funds in your account, please top up"
	TxtError               = "Unexcepted error occured! Please contact support"
	TxtHelp                = "Это бот для приваточки канала Blue Archive, пиши /start и гляди что у нас есть :3"
)

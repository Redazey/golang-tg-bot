package messages

import types "tgssn/internal/model/bottypes"

// Команды стартовых действий.
var BtnStart = types.TgKbRowButtons{
	types.TgKeyboardButton{Text: "Categories"},
	types.TgKeyboardButton{Text: "Profile"},
	types.TgKeyboardButton{Text: "Support"},
}

// Кнопки с категориями
var BtnCtgs = []types.TgRowButtons{
	{
		types.TgInlineButton{DisplayName: "💵 CR", Value: "CR"},
		types.TgInlineButton{DisplayName: "📁 TU", Value: "TU"},
		types.TgInlineButton{DisplayName: "📔 Ready Fulls", Value: "fullz"},
	},
}

// Покупка/возвращение назад
var BtnCR = []types.TgRowButtons{
	{
		types.TgInlineButton{DisplayName: "BUY FOR 6$ ❗️20% SALE❗️", Value: "buy CR"},
	},
	{
		types.TgInlineButton{DisplayName: "Back to categories", Value: "backToCtg"},
	},
}

var BtnTU = []types.TgRowButtons{
	{
		types.TgInlineButton{DisplayName: "BUY FOR 8$ ❗️20% SALE❗️", Value: "buy TU"},
	},
	{
		types.TgInlineButton{DisplayName: "Back to categories", Value: "backToCtg"},
	},
}

var BtnFullz = []types.TgRowButtons{
	{
		types.TgInlineButton{DisplayName: "Buy", Value: "buy Fullz"},
	},
	{
		types.TgInlineButton{DisplayName: "Back to categories", Value: "backToCtg"},
	},
}

// Отказ от оплаты
var BackToCtgBtn = []types.TgRowButtons{
	{
		types.TgInlineButton{DisplayName: "Back to categories", Value: "backToCtg"},
	},
}

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
var BackToProfileBtn = []types.TgRowButtons{
	{
		types.TgInlineButton{DisplayName: "Back to profile", Value: "backToProfile"},
	},
}

// Область "Константы и переменные": начало.

const (
	TxtStart     = "Hello, %v 👋.This is a bot for Experian and Trans union reports (cr,tu)"
	TxtCtgs      = "📰 Choose a category that interests you:"
	TxtReports   = "📁 Category: %v reports\nDescription: %v\n"
	TxtCRDesc    = "CR"
	TxtTUDesc    = "TU"
	TxtFullzDesc = `Fullz with ready experian in format
	{name;address;city;state;zip;dob;dl;dl issue date;expiration date}
	credit score 700+`
	TxtPaymentDesc = `Please send the data in the format:
	{Full name;address;city;state;ZIP;DOB;SSN}
	After that, the payment will be processed, and the money will be debited from your account
	❗️ In case we are unable to find information based on your data, the money will be refunded ❗️`
	TxtProfile            = "📰 ID: %v\n💎 Balance: %v$\n📊 Orders: %v"
	TxtSup                = "For assistance, please contact technical support: "
	TxtUnknownCommand     = "Unfortunately, this command is unknown to me. To get started, please enter /start."
	TxtPaymentQuestion    = "Enter the amount you wish to add to your account: "
	TxtPaymentNotInt      = "Please use only whole numbers and numbers that are above zero for input 😉"
	TxtPaymentSuccsessful = "Account topped up! 💵"
	TxtPaymentErr         = "An error occurred while transferring funds! Please double-check your information or contact support"
	TxtPaymentNotEnough   = "You have insufficient funds in your account, please top up"
	TxtOrderHistory       = "💡 Order № %v\n🕐 Date %v\n📁 Category %v\n💰 Sum %v\n➖➖➖➖➖➖➖➖➖➖➖➖"
	TxtHelp               = "This is a bot for Experian and Trans union reports (cr,tu). Enter /start"
)

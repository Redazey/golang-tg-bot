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

var BtnCR = []types.TgRowButtons{
	{
		types.TgInlineButton{DisplayName: "BUY FOR 6$ ❗️20% SALE❗️", Value: "buyCR"},
	},
	{
		types.TgInlineButton{DisplayName: "Back to categories", Value: "back"},
	},
}

var BtnTU = []types.TgRowButtons{
	{
		types.TgInlineButton{DisplayName: "BUY FOR 8$ ❗️20% SALE❗️", Value: "buyTU"},
	},
	{
		types.TgInlineButton{DisplayName: "Back to categories", Value: "back"},
	},
}

var BtnFullz = []types.TgRowButtons{
	{
		types.TgInlineButton{DisplayName: "Buy", Value: "buyFullz"},
	},
	{
		types.TgInlineButton{DisplayName: "Back to categories", Value: "back"},
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

// Область "Константы и переменные": начало.

const (
	TxtStart     = "Hello, %v 👋.This is a bot for Experian and Trans union reports (cr,tu)"
	TxtCtgs      = "📰 Choose a category that interests you:"
	TxtReports   = "Category: %v reports\nDescription: %v\n"
	TxtCRDesc    = "CR"
	TxtTUDesc    = "TU"
	TxtFullzDesc = `Fullz with ready experian in format {name:address:city;state;zip;dob;dl:dl} issue date,
	expiration date, credit score 700+`
	TxtProfile            = "📰 ID: %v\n💎 Balance: %v\n📊 Orders: %v"
	TxtSup                = "For assistance, please contact technical support: "
	TxtUnknownCommand     = "Unfortunately, this command is unknown to me. To get started, please enter /start."
	TxtPaymentQuestion    = "Enter the amount you wish to add to your account: "
	TxtPaymentNotInt      = "Please use only whole numbers for input 😉"
	TxtPaymentSuccsessful = "Account topped up! 💵"
	TxtPaymentErr         = "An error occurred while transferring funds! Please double-check your information or contact support"
	TxtPaymentNotEnough   = "You have insufficient funds in your account, please top up"
	TxtOrderHistory       = "💡 Order № %v\n🕐 Date %v\n📁 Category %v\n💰 Sum %v\n➖➖➖➖➖➖➖➖➖➖➖➖"
	TxtHelp               = "This is a bot for Experian and Trans union reports (cr,tu). Enter /start"
)

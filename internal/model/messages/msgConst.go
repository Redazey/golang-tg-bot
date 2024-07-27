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
		types.TgInlineButton{DisplayName: "Refill balance", Value: "deposit"},
	},
	{
		types.TgInlineButton{DisplayName: "Orders history", Value: "orders"},
		types.TgInlineButton{DisplayName: "Refill history", Value: "deposits"},
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
	TxtProfile          = "📰 ID: %v\n💎 Balance: %v\n📊 Orders: %v"
	TxtSup              = "For assistance, please contact technical support: "
	TxtUnknownCommand   = "Unfortunately, this command is unknown to me. To get started, please enter /start."
	TxtReportError      = "Can't get a report."
	TxtReportWait       = "Creating report, please wait..."
	TxtCatChoice        = "Выбрана категория *%v*. Введите сумму (только число). Для отмены введите 0. Используемая валюта: *%v*"
	TxtCatSave          = "Категория успешно сохранена."
	TxtCatEmpty         = "Пока нет категорий, сначала добавьте хотя бы одну категорию."
	TxtRecSave          = "Запись успешно сохранена."
	TxtRecOverLimit     = "Запись не сохранена: превышен бюджет раходов в текущем месяце."
	TxtRecTbl           = "Для загрузки истории расходов введите таблицу в следующем формате (дата сумма категория):\n`YYYY-MM-DD 0.00 XXX`\nНапример: \n`2022-09-20 1500 Кино`\n`2022-07-12 350.50 Продукты, еда`\n`2022-08-30 8000 Одежда и обувь`\n`2022-09-01 60 Бензин`\n`2022-09-27 425 Такси`\n`2022-09-26 1500 Бензин`\n`2022-09-26 950 Кошка`\n`2022-09-25 50 Бензин`\nИспользуемая валюта: *%v*"
	TxtReportQP         = "За какой период будем смотреть отчет? Команды периодов: /report_w - неделя, /report_m - месяц, /report_y - год"
	TxtHelp             = "Я - бот, помогающий вести учет расходов. Для начала работы введите /start"
	TxtCurrencyChoice   = "В качестве основной задана валюта: *%v*. Для изменения выберите другую валюту."
	TxtCurrencySet      = "Валюта изменена на *%v*."
	TxtCurrencySetError = "Ошибка сохранения валюты."
	TxtLimitInfo        = "Текущий ежемесячный бюджет: *%v*. Для изменения введите число, например, 80000."
	TxtLimitSet         = "Бюджет изменен на *%v*."
)

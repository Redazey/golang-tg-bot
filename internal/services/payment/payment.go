package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	consts "tgssn/internal/model/messages"
	"tgssn/pkg/errors"
	"tgssn/pkg/logger"
	"time"

	"go.uber.org/zap"
)

// MessageSender Интерфейс для работы с сообщениями.
type MessageSender interface {
	SendMessage(text string, userID int64) (int, error)
}

// UserDataStorage Интерфейс для работы с хранилищем данных.
type UserDataStorage interface {
	AddUserLimit(ctx context.Context, userID int64, limits float64) error

	ChangeRefillRecordStatus(ctx context.Context, status string, invoice_id int64) error
	GetRefillRecords(ctx context.Context) ([]int, error)
}

// Model Модель платёжки
type Model struct {
	ctx       context.Context
	storage   UserDataStorage // Хранилище пользовательской информации.
	tgClient  MessageSender   // Клиент.
	apiSecret string
}

// New Генерация сущности для хранения клиента ТГ и хранилища пользователей
func New(ctx context.Context, storage UserDataStorage, tgClient MessageSender, apiSecret string) *Model {
	return &Model{
		ctx:       ctx,
		storage:   storage,
		tgClient:  tgClient,
		apiSecret: apiSecret,
	}
}

// Struct для создания платежного счета
type CreateInvoiceRequest struct {
	Asset       string  `json:"asset"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	Payload     string  `json:"payload,omitempty"`
	Expires     int     `json:"expires_in"`
}

type Invoice struct {
	InvoiceID         int64    `json:"invoice_id"`
	Hash              string   `json:"hash"`
	CurrencyType      string   `json:"currency_type"`
	Asset             string   `json:"asset"`
	Fiat              string   `json:"fiat"`
	Amount            string   `json:"amount"`
	PaidAsset         string   `json:"paid_asset"`
	PaidAmount        string   `json:"paid_amount"`
	PaidFiatRate      string   `json:"paid_fiat_rate"`
	AcceptedAssets    []string `json:"accepted_assets"`
	FeeAsset          string   `json:"fee_asset"`
	FeeAmount         string   `json:"fee_amount"`
	Fee               string   `json:"fee"`     // Deprecated
	PayURL            string   `json:"pay_url"` // Deprecated
	BotInvoiceURL     string   `json:"bot_invoice_url"`
	MiniAppInvoiceURL string   `json:"mini_app_invoice_url"`
	WebAppInvoiceURL  string   `json:"web_app_invoice_url"`
	Description       string   `json:"description"`
	Status            string   `json:"status"`
	CreatedAt         string   `json:"created_at"`
	PaidUsdRate       string   `json:"paid_usd_rate"`
	UsdRate           string   `json:"usd_rate"` // Deprecated
	AllowComments     bool     `json:"allow_comments"`
	AllowAnonymous    bool     `json:"allow_anonymous"`
	ExpirationDate    string   `json:"expiration_date"`
	PaidAt            string   `json:"paid_at"`
	PaidAnonymously   bool     `json:"paid_anonymously"`
	Comment           string   `json:"comment"`
	HiddenMessage     string   `json:"hidden_message"`
	Payload           string   `json:"payload"`
	PaidBtnName       string   `json:"paid_btn_name"`
	PaidBtnURL        string   `json:"paid_btn_url"`
}

type GetInvoicesRequest struct {
	InvoiceIDs string `json:"invoice_ids"`
	Status     string `json:"status"`
}

type GetInvoicesResponse struct {
	Invoices []Invoice `json:"invoices"`
}
type DeleteInvoiceRequest struct {
	InvoiceID int64 `json:"invoice_id"`
}

type DeleteInvoiceResponse struct {
	Result bool
}

func (s *Model) Init() {
	go func() {
		for {
			invoiceIDs, err := s.storage.GetRefillRecords(s.ctx)
			if err != nil {
				logger.Error("Ошибка", zap.Error(err))
			}

			for i, invoiceID := range invoiceIDs {
				body, err := s.CryptoPayRequest(s.ctx, "getInvoices", GetInvoicesRequest{
					InvoiceIDs: fmt.Sprintf("%v", invoiceID),
					Status:     "paid",
				})
				if err != nil {
					logger.Error("Fail to get CryptoPayRequest: ", zap.Error(err))
					continue
				}

				var invoices = map[string][]Invoice{}

				if err := json.Unmarshal(body, &invoices); err != nil {
					logger.Error("ОШибка", zap.Error(err))
					time.Sleep(time.Second * 10)

					continue
				}

				var invoice Invoice

				if len(invoices["items"]) > i {
					invoice = invoices["items"][i]
				} else {
					time.Sleep(time.Second * 10)

					continue
				}

				if invoice.Status == "paid" {
					userID, err := strconv.Atoi(invoice.Payload)
					if err != nil {
						logger.Info(fmt.Sprintf("Попытка превратить это в int %v", invoice.Payload))
						logger.Error("Ошибка при конвертации типов данных", zap.Error(err))

						if err = s.storage.ChangeRefillRecordStatus(s.ctx, "error", invoice.InvoiceID); err != nil {
							logger.Error("failed to ChangeRefillRecordStatus", zap.Error(err))
						}
					}

					intAmount, err := strconv.Atoi(invoice.Amount)
					if err != nil {
						logger.Info(fmt.Sprintf("Попытка превратить это в int %v", invoice.Amount))
						logger.Error("Ошибка при конвертации типов данных", zap.Error(err))

					}
					logger.Info(fmt.Sprintf("%v - %v", userID, intAmount))

					if _, err = s.tgClient.SendMessage(fmt.Sprintf(consts.TxtPaymentSuccsessful, invoice.Amount), int64(userID)); err != nil {
						logger.Error("Failed to send message", zap.Error(err))

					}
					if err = s.storage.AddUserLimit(s.ctx, int64(userID), float64(intAmount)); err != nil {
						logger.Error("Failed to send message", zap.Error(err))

					}

					if err = s.storage.ChangeRefillRecordStatus(s.ctx, "paid", invoice.InvoiceID); err != nil {
						logger.Error("failed to ChangeRefillRecordStatus", zap.Error(err))
					}

					var result bool
					if err = json.Unmarshal(body, &result); err != nil {
						logger.Debug("Error while unmarshaling message")

						continue
					}

					if !result {
						logger.Info("Failed to delete invoice")
					}
				}
			}

			time.Sleep(time.Second * 10)
		}
	}()
}

// Функция для создания платежного счета
func (s *Model) CryptoPayRequest(ctx context.Context, method string, request any) ([]byte, error) {
	// Создаем URL для запроса
	url := fmt.Sprintf("https://pay.crypt.bot/api/%v", method)

	// Создаем заголовки для запроса
	headers := map[string]string{
		"Content-Type":         "application/json",
		"Crypto-Pay-API-Token": s.apiSecret,
		"Crypto-Pay-API-Nonce": fmt.Sprintf("%d", 1643723900), // Unix timestamp
		"Crypto-Pay-API-Sign":  GenerateSignature(s.apiSecret, "POST", url, request),
	}

	// Marshal json запроса
	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	// Отправляем запрос
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonRequest)))
	if err != nil {
		return nil, err
	}

	// Добавляем заголовки
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	// Отправляем запрос и получаем ответ
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if body == nil {
		return nil, nil
	}

	var prepare = make(map[string]any)
	if err := json.Unmarshal(body, &prepare); err != nil {
		return nil, err
	}

	if !prepare["ok"].(bool) {
		return nil, errors.New(fmt.Sprintf("%v", prepare["error"]))
	}

	result := prepare["result"]
	resultBody, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	// Verify signature
	signature := req.Header.Get("Crypto-Pay-API-Sign")
	if signature != GenerateSignature(s.apiSecret, "POST", url, request) {
		return nil, errors.New("Invalid signature")
	}

	return resultBody, nil
}

// Функция для генерации подписи
func GenerateSignature(secret string, method, url string, request interface{}) string {
	// Marshal json запроса
	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return ""
	}

	// Создаем строку для подписи
	signString := fmt.Sprintf("%s\n%s\n%s\n%s", method, url, string(jsonRequest), "application/json")

	// Генерируем подпись
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(signString))
	return fmt.Sprintf("%x", h.Sum(nil))
}

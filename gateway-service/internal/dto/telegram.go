package dto

type TelegramUserDTO struct {
	TelegramID string `json:"telegramID" binding:"required"`
}

type TelegramInfo struct {
	TelegramID string `json:"telegramID"`
	UserID     string `json:"userID"`
	DeviceID   string `json:"deviceID"`
}

type TelegramProfile struct {
	TelegramID string `json:"telegramID"`
	UserID     string `json:"userID"`
	Email      string `json:"email,omitempty"`
}

type TelegramSubscription struct {
	SubscriptionID string `json:"subscriptionID"`
	UserID         string `json:"userID"`
	StartsAt       int64  `json:"startsAt"`
	ExpiresAt      int64  `json:"expiresAt"`
}

type TelegramChatDTO struct {
	TelegramID string `json:"telegramID" binding:"required"`
	Prompt     string `json:"prompt" binding:"required"`
}

type TelegramChatResponse struct {
	TelegramID string `json:"telegramID"`
	Response   string `json:"response"`
}

type TelegramProfileAnalyzeDTO struct {
	TelegramID        string `json:"telegramID" binding:"required"`
	ChatID            int64  `json:"chatID" binding:"required"`
	ProgressMessageID int64  `json:"progressMessageID,omitempty"`
	FirstName         string `json:"firstName" binding:"required"`
	LastName          string `json:"lastName,omitempty"`
	Username          string `json:"username,omitempty"`
	Bio               string `json:"bio,omitempty"`
	IsPremium         bool   `json:"isPremium"`
	LanguageCode      string `json:"languageCode,omitempty"`
	PhotoBase64       string `json:"photoBase64,omitempty"`
}

type TelegramProfileAnalyzeAcceptedResponse struct {
	JobID string `json:"jobId"`
}

type TelegramProfileAnalyzeResponse struct {
	TelegramID string `json:"telegramID"`
	Response   string `json:"response"`
}

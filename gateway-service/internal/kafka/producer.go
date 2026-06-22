package kafka

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
)

const TopicProfileAnalyzeRequests = "profile.analyze.requests"

type ProfileAnalyzeJob struct {
	JobID             string `json:"jobId"`
	TelegramID        string `json:"telegramID"`
	ChatID            int64  `json:"chatID"`
	ProgressMessageID int64  `json:"progressMessageID,omitempty"`
	FirstName         string `json:"firstName"`
	LastName          string `json:"lastName,omitempty"`
	Username          string `json:"username,omitempty"`
	Bio               string `json:"bio,omitempty"`
	IsPremium         bool   `json:"isPremium"`
	LanguageCode      string `json:"languageCode,omitempty"`
	PhotoBase64       string `json:"photoBase64,omitempty"`
}

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        TopicProfileAnalyzeRequests,
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireOne,
			BatchTimeout: 10,
		},
	}
}

func (p *Producer) PublishProfileAnalyzeJob(ctx context.Context, job ProfileAnalyzeJob) error {
	payload, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(job.TelegramID),
		Value: payload,
	})
}

func (p *Producer) Close() error {
	if p == nil || p.writer == nil {
		return nil
	}
	return p.writer.Close()
}

package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"gateway/internal/logging"
	"gateway/internal/storage"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type ServerTokenService struct {
	ctx        context.Context
	ttl        time.Duration
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	redis      *storage.RedisClient
}

type ServiceClaims struct {
	ServiceName string   `json:"svc"`
	AllowedDst  []string `json:"dst"` // какие сервисы может вызывать
	jwt.RegisteredClaims
}

const (
	redisPublicKeyKey     = "gateway:public_key"
	redisPublicKeyChannel = "gateway:public_key:updated"
)

func NewServerTokenService(ctx context.Context, ttl time.Duration, rdb *storage.RedisClient) (*ServerTokenService, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	svc := &ServerTokenService{
		ctx:        ctx,
		ttl:        ttl,
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
		redis:      rdb,
	}

	if err := svc.publishPublicKey(); err != nil {
		return nil, fmt.Errorf("failed to publish public key: %w", err)
	}

	return svc, nil
}

func (s *ServerTokenService) publishPublicKey() error {
	pubPEM, err := s.exportPublicKeyPEM()
	if err != nil {
		return err
	}

	if err := s.redis.Client.Set(s.ctx, redisPublicKeyKey, pubPEM, 0).Err(); err != nil {
		return fmt.Errorf("failed to write public key to redis: %w", err)
	}

	if err := s.redis.Client.Publish(s.ctx, redisPublicKeyChannel, "updated").Err(); err != nil {
		logging.Logger.Warn("failed to publish key update notification", zap.Error(err))
	}

	logging.Logger.Info("public key published to redis")
	return nil
}

func (s *ServerTokenService) exportPublicKeyPEM() ([]byte, error) {
	pubBytes, err := x509.MarshalPKIXPublicKey(s.publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}
	return pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}), nil
}

func (s *ServerTokenService) IssueServiceToken(serviceName string, allowedDst []string) (string, error) {
	claims := ServiceClaims{
		ServiceName: serviceName,
		AllowedDst:  allowedDst,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign service token: %w", err)
	}

	return signed, nil
}

func (s *ServerTokenService) ParseAccessServerToken(tokenStr string) (*ServiceClaims, error) {
	logger := logging.Logger

	var claims ServiceClaims

	token, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.publicKey, nil
	})
	if err != nil {
		logger.Error("failed to parse service token", zap.Error(err))
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		logger.Warn("service token is not valid")
		return nil, errors.New("token is not valid")
	}

	return &claims, nil
}

package script

import (
	"fmt"
	"subscription/internal/logging"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func RunMigrations(db *gorm.DB, models []any) error {
	logger := logging.Logger
	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			logger.Error("Error making migration", zap.Error(err), zap.Any("model", model))
			return fmt.Errorf("migration failed for %T: %w", model, err)

		}
		logger.Info("Migration for model done", zap.String("model", fmt.Sprintf("%T", model)))
	}
	return nil
}

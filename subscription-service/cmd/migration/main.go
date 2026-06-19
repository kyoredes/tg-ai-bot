package main

import (
	"fmt"
	"os"
	"subscription/cmd/migration/script"
	"subscription/internal/config"
	"subscription/internal/logging"
	"subscription/internal/models"
	"subscription/internal/storage"

	"github.com/subosito/gotenv"
)

func main() {
	if err := gotenv.Load(".env"); err != nil && !os.IsNotExist(err) {
		fmt.Println(err)
		return
	}
	config.Init()
	cfg := config.NewConfig()

	if err := logging.InitLogger(cfg.LoggingMode); err != nil {
		fmt.Println(err)
		return
	}

	db, err := storage.NewDatabase(config.NewDBConfig(), models.ModelsList)
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := script.RunMigrations(db, models.ModelsList); err != nil {
		fmt.Println(err)
		return
	}
}

package init_mysql

import (
	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"git.blauwelle.com/go/crate/cmd/sso/config"
	"git.blauwelle.com/go/crate/cmd/sso/model"
)

var (
	StartCmd = &cobra.Command{
		Use:          "init-mysql",
		Short:        "init-mysql",
		Example:      "init-mysql",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run()
		},
	}
)

func run() error {
	cfg, err := config.GetConfig("config/config.yaml")
	if err != nil {
		return err
	}

	db, err := gorm.Open(
		mysql.Open(cfg.Mysql.DSN),
		&gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		},
	)
	if err != nil {
		return err
	}

	err = db.Migrator().AutoMigrate(
		&model.Application{},
		&model.Role{},
		&model.User{},
		&model.UserRole{},
	)
	if err != nil {
		return err
	}

	db.Create(&model.Role{
		Model: model.Model{ID: 1},
		Name:  "admin",
	})

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("nil"), 12)
	db.Create(&model.User{
		Model:        model.Model{ID: 1},
		Username:     "bob",
		PasswordHash: string(passwordHash),
	})

	db.Create(&model.UserRole{UserID: 1, RoleID: 1})

	return nil
}

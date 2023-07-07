package sso_server

import (
	"context"
	"net/http"
	"os"
	"strconv"

	"git.blauwelle.com/go/crate/exegroup"
	"git.blauwelle.com/go/crate/exegroup/eghttp"
	"git.blauwelle.com/go/crate/log"
	"git.blauwelle.com/go/crate/log/logsdk"
	"git.blauwelle.com/go/crate/log/logsdk/logjson"
	"github.com/spf13/cobra"

	"git.blauwelle.com/go/crate/cmd/sso/config"
	"git.blauwelle.com/go/crate/cmd/sso/database"
	"git.blauwelle.com/go/crate/cmd/sso/router"
	"git.blauwelle.com/go/crate/cmd/sso/util"
)

var (
	StartCmd = &cobra.Command{
		Use:          "sso-server",
		Short:        "sso-server",
		Example:      "sso-server",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run()
		},
	}
)

func run() error {
	ctx := context.Background()

	cfg, err := config.GetConfig("config/config.yaml")
	if err != nil {
		return err
	}

	log.Logger().AddProcessor(logsdk.AllLevels, logjson.New(logjson.WithPrettyPrint(true), logjson.WithDisableHTMLEscape(true)))

	// JWT
	keyBytes, err := os.ReadFile("private.rsa")
	if err != nil {
		return err
	}

	jwt, err := util.NewJWTFromKeyBytes(keyBytes)
	if err != nil {
		return err
	}

	db, err := database.NewMysql(cfg)
	if err != nil {
		log.Error(context.TODO(), "Failed to connect to MySQL database")
		return err
	}

	redisDB, err := database.NewRedis(cfg)

	routers := router.NewRouter(db, redisDB, jwt)
	group := exegroup.Default()
	group.New().WithGoStop(eghttp.HTTPListenAndServe(eghttp.WithServerOption(func(server *http.Server) {
		server.Addr = ":" + strconv.Itoa(cfg.Listen.Port)
		server.Handler = routers
	})))

	log.Info(ctx, "Server start...")

	if err := group.Run(ctx); err != nil {
		log.Error(ctx, err.Error())
		log.Exit(1)
	}

	return nil
}

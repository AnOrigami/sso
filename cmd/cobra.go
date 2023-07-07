package cmd

import (
	"context"
	"fmt"

	"git.blauwelle.com/go/crate/log"
	"github.com/spf13/cobra"

	"git.blauwelle.com/go/crate/cmd/sso/cmd/init_mysql"
	"git.blauwelle.com/go/crate/cmd/sso/cmd/sso_server"
)

var rootCmd = &cobra.Command{
	Use:           "sso",
	Short:         "sso",
	Long:          "sso",
	SilenceErrors: true,
	SilenceUsage:  true,
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	Example: "go run main.go init-mysql\n   go run main.go sso-server",
	Run: func(cmd *cobra.Command, args []string) {
		sso()
	},
}

func sso() {
	fmt.Printf("欢迎使用sso\n")
	fmt.Printf("如果您是第一次启动，请先执行 go run main.go init-mysql 初始化数据库\n")
	fmt.Printf("初始完数据库之后，执行 go run main.go sso-server 启动sso服务\n")
	fmt.Printf("不是第一次启动执行 go run main.go sso-server 即可启动sso服务\n")
}

func init() {
	rootCmd.AddCommand(init_mysql.StartCmd)
	rootCmd.AddCommand(sso_server.StartCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(context.TODO(), err)
	}
}

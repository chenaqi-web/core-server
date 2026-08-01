package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var descRoot = strings.Join([]string{
	"这个命令行工具提供sql等子命令的工具",
	"-h:  获取命令的帮助信息",
	"sql: 有关sql的初始化操作",
}, "\n")

var rootCmd = &cobra.Command{
	Use:   "cobra",
	Short: "项目命令行工具",
	Long:  descRoot,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("cobra test")
	},
}

func Execute() error {
	return rootCmd.Execute()
}

var userLicense string

func init() {
	// 这里是添加标签
	rootCmd.PersistentFlags().Bool("viper", true, "是否采用viper作为配置文件读取")
	rootCmd.PersistentFlags().StringP("author", "a", "ChenA7", "作者名字")
	rootCmd.PersistentFlags().StringVarP(&userLicense, "license", "l", "", "授权信息")

	// 添加子命令
	rootCmd.AddCommand(wordCmd)
}

func main() {
	if err := Execute(); err != nil {
		fmt.Println(err)
	}
}

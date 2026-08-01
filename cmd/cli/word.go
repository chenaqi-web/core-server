package main

import (
	"log"
	"strings"

	"github.com/spf13/cobra"
)

const (
	MODE_UPPER = iota + 1 // 单词转化为大写
	MODE_LOWER            // 单词转化为小写
)

var str string
var mode int8

var desc = strings.Join([]string{
	"该子命令支持单词的大小写转换，模式如下:",
	"1: 全部单词转化为大写",
	"2: 全部单词转化为小写",
}, "\n")

var wordCmd = &cobra.Command{
	Use:   "word",
	Short: "单词格式转换",
	Long:  desc,
	Run: func(cmd *cobra.Command, args []string) {
		var content string
		switch mode {
		case MODE_UPPER:
			content = ToUpper(str)
		case MODE_LOWER:
			content = ToLower(str)
		default:
			log.Fatal("暂时不支持该转换模式，请执行 help word 查看帮助文档")
		}
		log.Printf("输出的结果：%s", content)
	},
}

func init() {
	wordCmd.Flags().StringVarP(&str, "str", "s", "", "请输入单词内容")
	wordCmd.Flags().Int8VarP(&mode, "mode", "m", 0, "请输入单词转换的模式")
}

// ToUpper 全部转化为大写
func ToUpper(s string) string {
	return strings.ToUpper(s)
}

// ToLower 全部转化为小写
func ToLower(s string) string {
	return strings.ToLower(s)
}

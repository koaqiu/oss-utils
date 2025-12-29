/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "oss-utils",
	Short: "阿里云OSS工具集",
	Long: `阿里云OSS工具集
	管理自定义域名、配置SSL证书等功能。`,
	Run: func(cmd *cobra.Command, args []string) {},
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// version is set at build time using -ldflags, default to dev
var version = "v1.0.2"

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.oss-utils.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// 全局 version 标志（-v, --version）
	rootCmd.PersistentFlags().BoolP("version", "v", false, "显示版本信息")

	// 在任何子命令之前处理 version 标志并优先返回版本信息
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		v, _ := cmd.Flags().GetBool("version")
		if v {
			fmt.Println(version)
			os.Exit(0)
		}
	}
}

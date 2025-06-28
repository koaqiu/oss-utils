/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"regexp"

	"slices"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"github.com/spf13/cobra"
)

func validateRegion(region string, validRegions []string) bool {
	isValidRegion := slices.Contains(validRegions, region)
	if !isValidRegion {
		fmt.Printf("无效的OSS区域: %s\n", region)
		fmt.Println("请使用以下有效的区域之一:")
		for _, validRegion := range validRegions {
			fmt.Println(validRegion)
		}
		return false
	}
	return true
}

func validateBucketName(bucket string) bool {
	bucketRegex := "^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$"
	if bucket == "" || !regexp.MustCompile(bucketRegex).MatchString(bucket) {
		fmt.Printf("无效的Bucket名称: %s\n", bucket)
		fmt.Println("Bucket名称必须满足以下规则：")
		fmt.Println("- 长度为3到63个字符")
		fmt.Println("- 只能包含小写字母、数字和连字符（-）")
		fmt.Println("- 必须以字母或数字开头和结尾")
		return false
	}
	return true
}

func readCertificateAndKeyFiles(certPath, keyPath string) (string, string, error) {
	cert, err := os.ReadFile(certPath)
	if err != nil {
		return "", "", fmt.Errorf("Error reading certificate file: %v", err)
	}
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return "", "", fmt.Errorf("Error reading key file: %v", err)
	}
	return string(cert), string(key), nil
}

func printCnameList(response *oss.ListCnameResult) {
	fmt.Println("CNAMEs for bucket:", *response.Bucket)
	if len(response.Cnames) == 0 {
		fmt.Println("没有配置CNAME。请先配置CNAME。")
		return
	}
	for _, cname := range response.Cnames {
		fmt.Println("HOST:", *cname.Domain)
		fmt.Println("SSL Status:", *cname.Status)
		fmt.Println("SSL CertId:", *cname.Certificate.CertId)
	}
}

var sslCmd = &cobra.Command{
	Use:   "ssl",
	Short: "自定义域名的SSL证书管理",
	Long:  `检索或者更新自定义域名的SSL证书。`,
	Run: func(cmd *cobra.Command, args []string) {
		if os.Getenv("OSS_ACCESS_KEY_ID") == "" || os.Getenv("OSS_ACCESS_KEY_SECRET") == "" {
			fmt.Println("请设置环境变量 OSS_ACCESS_KEY_ID 和 OSS_ACCESS_KEY_SECRET")
			return
		}

		region := cmd.Flag("region").Value.String()
		validRegions := []string{
			"cn-shanghai",
			"cn-beijing",
			"cn-hangzhou",
			"cn-qingdao",
			"cn-zhangjiakou",
			"cn-huhehaote",
			"cn-chengdu",
			"cn-hongkong",
			"us-west-1",
			"us-east-1",
			"ap-southeast-1",
			"ap-southeast-2",
			"ap-southeast-3",
			"ap-southeast-5",
			"ap-northeast-1",
			"eu-central-1",
			"me-east-1",
		}

		if !validateRegion(region, validRegions) {
			return
		}

		bucket := cmd.Flag("bucket").Value.String()
		if !validateBucketName(bucket) {
			return
		}

		fmt.Printf("当前操作的区域: %s\n", region)
		fmt.Printf("当前操作的Bucket名称: %s\n", bucket)

		certPath := cmd.Flag("cert").Value.String()
		keyPath := cmd.Flag("key").Value.String()
		if (certPath == "" && keyPath != "") || (certPath != "" && keyPath == "") {
			fmt.Println("请同时提供新的SSL证书和密钥，或不提供以使用现有的SSL证书。")
			return
		}
		isUpdateSll := false
		var certStr, keyStr string
		if certPath != "" && keyPath != "" {
			certStr2, keyStr2, err := readCertificateAndKeyFiles(certPath, keyPath)
			if err != nil {
				fmt.Println(err)
				return
			}
			certStr = certStr2
			keyStr = keyStr2
			isUpdateSll = true
		}

		cfg := oss.LoadDefaultConfig().
			WithCredentialsProvider(credentials.NewEnvironmentVariableCredentialsProvider()).
			WithRegion(region)
		client := oss.NewClient(cfg)
		request := &oss.ListCnameRequest{
			Bucket: &bucket,
		}
		response, err := client.ListCname(cmd.Context(), request)
		if err != nil {
			fmt.Printf("Error listing CNAMEs: %v\n", err)
			return
		}

		printCnameList(response)

		if len(response.Cnames) == 0 {
			return
		}
		if !isUpdateSll {
			return
		}
		fmt.Println("正在更新CNAME的SSL证书...")
		forceUpdate := true // 强制更新SSL证书
		updateRequest := &oss.PutCnameRequest{
			Bucket: &bucket,
			BucketCnameConfiguration: &oss.BucketCnameConfiguration{
				Domain: response.Cnames[0].Domain,
				CertificateConfiguration: &oss.CertificateConfiguration{
					PreviousCertId: response.Cnames[0].Certificate.CertId,
					Certificate:    &certStr,
					PrivateKey:     &keyStr,
					Force:          &forceUpdate,
				},
			},
		}

		_, err = client.PutCname(cmd.Context(), updateRequest)
		if err != nil {
			fmt.Printf("Error updating CNAME SSL certificate: %v\n", err)
			return
		}

		fmt.Println("SSL证书已成功更新。")
	},
}

func init() {
	rootCmd.AddCommand(sslCmd)

	// Here you will define your flags and configuration settings.

	// 添加region参数
	// 该参数是可选的，用于指定OSS服务的区域，默认为"cn-shanghai"
	sslCmd.Flags().String("region", "cn-shanghai", "OSS region")
	// 添加bucket参数
	// 该参数是必需的，用于指定要操作的OSS Bucket名称
	sslCmd.Flags().String("bucket", "", "Bucket name")
	sslCmd.MarkFlagRequired("bucket") // 确保bucket参数是必需的

	// 指定新的SSL证书的路径
	sslCmd.Flags().String("cert", "", "新的SSL证书的证书文件路径")
	// 指定新的SSL证书的私钥路径
	sslCmd.Flags().String("key", "", "新的SSL证书的密钥文件路径")
	// 确保cert和key参数是可选的
	sslCmd.MarkFlagFilename("cert", "pem", "crt") // 确保cert参数是文件名格式
	sslCmd.MarkFlagFilename("key", "pem", "key")  // 确保key参数是文件名格式
}

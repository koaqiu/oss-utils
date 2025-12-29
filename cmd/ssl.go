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
	"github.com/pterm/pterm" // 引入 pterm 模块用于表格输出
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
	fmt.Printf("%v 的CNAME列表\n", pterm.Yellow(*response.Bucket))
	if len(response.Cnames) == 0 {
		pterm.Warning.Println("没有配置CNAME。请先配置CNAME。")
		return
	}
	index := 1
	// 使用 pterm 创建表格输出
	tableData := pterm.TableData{
		{"索引", "HOST", "状态", "SSL CertId", "SSL 过期时间"},
	}
	for _, cname := range response.Cnames {
		var certId, expireTime string
		if cname.Certificate != nil {
			certId = *cname.Certificate.CertId
			expireTime = *cname.Certificate.ValidEndDate
		} else {
			certId = "未配置"
			expireTime = ""
		}
		tableData = append(tableData, []string{fmt.Sprintf("%d", index), *cname.Domain, *cname.Status, certId, expireTime})
		index++
	}
	// 使用 pterm.Table 渲染表格
	pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
}

func getCredentialsProvider(accessKeyID, accessKeySecret string) (credentials.CredentialsProvider, error) {
	// 如果同时提供 accessKeyID 和 accessKeySecret，使用静态凭证提供者
	if accessKeyID != "" || accessKeySecret != "" {
		if accessKeyID == "" || accessKeySecret == "" {
			return nil, fmt.Errorf("oss access id 和 secret 必须同时提供，或者都不提供以使用环境变量")
		}
		return credentials.NewStaticCredentialsProvider(accessKeyID, accessKeySecret, ""), nil
	}

	// 未通过参数提供，则检查环境变量是否存在
	if os.Getenv("OSS_ACCESS_KEY_ID") == "" || os.Getenv("OSS_ACCESS_KEY_SECRET") == "" {
		return nil, fmt.Errorf("请提供 OSS Access Key ID 和 Secret，或者设置环境变量 OSS_ACCESS_KEY_ID 和 OSS_ACCESS_KEY_SECRET")
	}

	return credentials.NewEnvironmentVariableCredentialsProvider(), nil
}

// 修改Run函数以支持quiet模式
var sslCmd = &cobra.Command{
	Use:   "ssl",
	Short: "自定义域名的SSL证书管理",
	Long:  `检索或者更新自定义域名的SSL证书。`,
	Run: func(cmd *cobra.Command, args []string) {
		// 获取quiet标志的值
		quiet, _ := cmd.Flags().GetBool("quiet")

		// 支持通过命令行参数传入 Access Key，如果未传入则尝试从环境变量读取
		accessKeyID := cmd.Flag("oss-access-id").Value.String()
		accessKeySecret := cmd.Flag("oss-access-secret").Value.String()

		// 校验并创建凭证提供者（支持传入或回退到环境变量）
		credProvider, credErr := getCredentialsProvider(accessKeyID, accessKeySecret)
		if credErr != nil {
			pterm.Error.Println(credErr)
			os.Exit(1)
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
			os.Exit(1)
			return
		}

		bucket := cmd.Flag("bucket").Value.String()
		if !validateBucketName(bucket) {
			os.Exit(1)
			return
		}

		if !quiet {
			fmt.Printf("当前操作的区域: %s\n", region)
			fmt.Printf("当前操作的Bucket名称: %s\n", bucket)
		}

		certPath := cmd.Flag("cert").Value.String()
		keyPath := cmd.Flag("key").Value.String()
		if (certPath == "" && keyPath != "") || (certPath != "" && keyPath == "") {
			pterm.Error.Println("如果需要更新SSL证书，请同时提供 --cert 和 --key 参数。")
			os.Exit(1)
			return
		}
		isUpdateSll := false
		var certStr, keyStr string
		if certPath != "" && keyPath != "" {
			certStr2, keyStr2, err := readCertificateAndKeyFiles(certPath, keyPath)
			if err != nil {
				pterm.Error.Printf("读取证书或密钥文件时出错: %v\n", err)
				os.Exit(1)
				return
			}
			certStr = certStr2
			keyStr = keyStr2
			isUpdateSll = true
		}

		cfg := oss.LoadDefaultConfig().
			WithCredentialsProvider(credProvider).
			WithRegion(region)
		client := oss.NewClient(cfg)
		request := &oss.ListCnameRequest{
			Bucket: &bucket,
		}
		response, err := client.ListCname(cmd.Context(), request)
		if err != nil {
			pterm.Error.Printf("无法列出CNAMEs: %v\n", err)
			os.Exit(1)
			return
		}

		if !quiet {
			printCnameList(response)
		}

		if len(response.Cnames) == 0 {
			if quiet {
				pterm.Error.Println("没有配置CNAME。请先配置CNAME。")
			}
			os.Exit(1)
			return
		}
		if !isUpdateSll {
			return
		}

		var domainIndex int = 0
		domain := cmd.Flag("domain").Value.String()

		if domain != "" {
			// 检查domain是否在response.Cnames列表里
			domainIndex = -1
			for i, cname := range response.Cnames {
				if *cname.Domain == domain {
					domainIndex = i
					break
				}
			}

			if domainIndex == -1 {
				fmt.Printf("指定的域名 %s 不存在于CNAME列表中。\n", domain)
				os.Exit(1)
				return
			}
		}

		if !quiet {
			fmt.Println("正在更新CNAME的SSL证书...")
		}

		forceUpdate := true // 强制更新SSL证书
		updateRequest := &oss.PutCnameRequest{
			Bucket: &bucket,
			BucketCnameConfiguration: &oss.BucketCnameConfiguration{
				Domain: response.Cnames[domainIndex].Domain,
				CertificateConfiguration: &oss.CertificateConfiguration{
					PreviousCertId: response.Cnames[domainIndex].Certificate.CertId,
					Certificate:    &certStr,
					PrivateKey:     &keyStr,
					Force:          &forceUpdate,
				},
			},
		}

		_, err = client.PutCname(cmd.Context(), updateRequest)
		if err != nil {
			fmt.Printf("Error updating CNAME SSL certificate: %v\n", err)
			os.Exit(1)
			return
		}

		if !quiet {
			fmt.Println("SSL证书已成功更新。")
		}
	},
}

func init() {
	rootCmd.AddCommand(sslCmd)
	// 添加region参数
	// 该参数是可选的，用于指定OSS服务的区域，默认为"cn-shanghai"
	sslCmd.Flags().StringP("region", "R", "cn-shanghai", "OSS region")
	// 添加bucket参数
	// 该参数是必需的，用于指定要操作的OSS Bucket名称
	sslCmd.Flags().StringP("bucket", "B", "", "Bucket name")
	sslCmd.MarkFlagRequired("bucket") // 确保bucket参数是必需的

	sslCmd.Flags().StringP("domain", "D", "", "指定要更新的CNAME域名，如果不指定，则默认使用第一个CNAME域名")

	// OSS AppId Secret参数
	// 这些参数是可选的，如果未提供，则从环境变量中读取
	sslCmd.Flags().String("oss-access-id", "", "OSS Access Key ID (可选，如果未提供则从环境变量读取)")
	sslCmd.Flags().String("oss-access-secret", "", "OSS Access Key Secret (可选，如果未提供则从环境变量读取)")

	// 指定新的SSL证书的路径
	sslCmd.Flags().String("cert", "", "新的SSL证书的证书文件路径")
	// 指定新的SSL证书的私钥路径
	sslCmd.Flags().String("key", "", "新的SSL证书的密钥文件路径")
	// 确保cert和key参数是可选的
	sslCmd.MarkFlagFilename("cert", "pem", "crt") // 确保cert参数是文件名格式
	sslCmd.MarkFlagFilename("key", "pem", "key")  // 确保key参数是文件名格式

	// 添加quiet参数
	sslCmd.Flags().BoolP("quiet", "q", false, "启用安静模式，仅输出错误信息")
}

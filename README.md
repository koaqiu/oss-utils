# OSS管理工具

目前唯一的功能就是 更新 `bucket`的自定义域名的SSL证书。

要求**必须**先配置好一个自定义域名，以及正常可用的SSL证书（`pem`和`key`）。

## 使用

``` powershell
# powershell
# 克隆
git clone https://github.com/koaqiu/oss-utils.git
cd oss-utils
go env -w GOPROXY=https://goproxy.cn,direct ; go mod tidy ; go build # 编译
# 编译成功以后
.\oss-utils.exe ssl --bucket BUCKET --region cn-shanghai --cert domain.pem --key domain.key
```

配合`certbot`等证书工具使用更佳。

## 配置

* 需要阿里云的RAM账号的 `OSS_ACCESS_KEY_ID`和`OSS_ACCESS_KEY_SECRET`，并设置到环境变量中。程序会自动读取。
* 需要为这个账号配置对于权限（参考下面的代码）[https://ram.console.aliyun.com/policies/create](https://ram.console.aliyun.com/policies/create)

自定义策略

```json
{
  "Version": "1",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "oss:PutCname",
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "yundun-cert:DescribeSSLCertificatePrivateKey",
        "yundun-cert:DescribeSSLCertificatePublicKeyDetail",
        "yundun-cert:CreateSSLCertificate"
      ],
      "Resource": "*"
    }
  ]
}
```

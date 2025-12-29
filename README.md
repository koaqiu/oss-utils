# OSS管理工具

![GitHub Release](https://img.shields.io/github/v/release/koaqiu/oss-utils)
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/koaqiu/oss-utils/.github%2Fworkflows%2Fmain.yml)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/koaqiu/oss-utils)
![GitHub License](https://img.shields.io/github/license/koaqiu/oss-utils)

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

### 版本信息与 CI 注入

构建时可以通过 `-ldflags` 注入版本号，程序内的 `-v / --version` 标志会显示该值。建议在 CI 中使用 git tag（例如 `v1.2.3`）作为版本并注入：

```powershell
# 本地示例：
go build -ldflags "-X 'github.com/koaqiu/oss-utils/cmd.version=1.2.3'"
.\\\oss-utils.exe -v  # 输出: 1.2.3
```

CI 工作流（已配置）会在基于 `v*` 标签触发时自动把 tag 值作为 VERSION 注入到构建产物中并创建 Release。请在打 tag 时使用 `v` 前缀（例如 `v1.2.3`）。

### 本地构建与验证

以下示例展示了如何在本地用不同方式注入版本并验证输出：

- PowerShell（Windows）：

```powershell
# 临时注入版本并构建
go build -ldflags "-X 'github.com/koaqiu/oss-utils/cmd.version=v1.2.3'" -o oss-utils.exe

# 运行并查看版本
.\oss-utils.exe -v  # 输出: v1.2.3
```

- Bash（Linux / macOS）：

```bash
# 注入版本并构建
go build -ldflags "-X 'github.com/koaqiu/oss-utils/cmd.version=v1.2.3'" -o oss-utils

# 运行并查看版本
./oss-utils --version  # 输出: v1.2.3
```

- 使用环境变量（例如在脚本中设置 OSS 凭证用于手动测试）：

```powershell
$env:OSS_ACCESS_KEY_ID='ID'; $env:OSS_ACCESS_KEY_SECRET='SECRET'; .\oss-utils.exe ssl --bucket my-bucket --region cn-shanghai
```

如果你需要生成跨平台二进制并同时注入版本（与 CI 行为一致），可以使用 gox 或者在 CI 中按当前 workflow 使用 `-ldflags` 注入 `cmd.version`。

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

## 更新日志

### v1.0.2

* 添加了`--oss-access-id` 和  `--oss-access-secret`，方便命令行脚本中使用

### v1.0.1

* 添加了多域名支持 `--domain`
* 添加了静默模式 `--quiet`、`-q`
* 发生错误的时候返回错误码

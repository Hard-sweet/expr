# expr

`expr` 是一个使用 Go 开发的 mDNS 资产测绘 CLI 原型。  
当前版本采用“策略二”实现路径：

- mDNS / DNS-SD 服务发现复用成熟库
- 资产归一、TXT 解析、banner enrich 由项目自身实现
- 支持 `CIDR + 端口范围` 过滤
- 支持文本输出和 JSON 输出
- 支持基于 `testdata` 的离线 golden tests

需要特别注意的是：mDNS 本身主要工作在本地链路广播域内。  
所以这个项目当前更准确的定位是：

- 在本地链路内发现 mDNS 服务
- 再按用户输入的 `CIDR` 和 `ports` 过滤结果
- 对识别到的服务做 banner 归一和基础深度识别

它不是一个跨三层网络强扫 mDNS 的通用远程扫描器。

## 项目结构

```text
expr/
├─ cmd/
│  └─ mdnsmap/
│     └─ main.go
├─ internal/
│  ├─ app/
│  │  ├─ run.go
│  │  └─ pipeline_test.go
│  ├─ cli/
│  │  └─ flags.go
│  ├─ discover/
│  │  └─ mdns/
│  │     └─ discover.go
│  ├─ enrich/
│  │  └─ build.go
│  ├─ fingerprint/
│  │  └─ engine.go
│  ├─ model/
│  │  ├─ asset.go
│  │  ├─ banner.go
│  │  ├─ mdns.go
│  │  └─ output.go
│  ├─ output/
│  │  ├─ json.go
│  │  └─ text.go
│  ├─ probe/
│  │  └─ http.go
│  └─ util/
│     ├─ iprange.go
│     └─ ports.go
├─ testdata/
│  ├─ golden/
│  │  ├─ qnap_filtered.json
│  │  ├─ qnap_filtered.txt
│  │  ├─ qnap_full.json
│  │  └─ qnap_full.txt
│  └─ samples/
│     └─ qnap_discovery.json
├─ go.mod
├─ go.sum
└─ README.md
```

### 目录职责

`cmd/mdnsmap`

- CLI 主入口
- 负责组装上下文和启动应用

`internal/app`

- 主流程编排
- 将发现、过滤、识别、输出串起来
- `pipeline_test.go` 使用离线样本验证完整链路

`internal/cli`

- 负责解析命令行参数
- 当前支持 `cidr`、`ports`、`iface`、`timeout`、`json`、`active-probe`

`internal/discover/mdns`

- mDNS/DNS-SD 发现适配层
- 隔离第三方库实现细节

`internal/enrich`

- 将发现记录转成标准化服务对象
- 处理 `CIDR + 端口范围` 过滤
- 将服务归并为资产

`internal/fingerprint`

- 服务级 banner enrich
- 当前对 `qdiscover`、`device-info`、`http`、`smb`、`afpovertcp`、`workstation` 做基础识别

`internal/probe`

- 主动探测能力
- 当前仅包含基础 HTTP 探测

`internal/output`

- 文本输出
- JSON 输出

`testdata/samples`

- 离线伪造发现样本

`testdata/golden`

- 样本对应的期望输出
- 用于 golden tests 回归

## 当前能力

当前版本支持：

- 浏览 `_services._dns-sd._udp.local`
- 枚举链路内可见的 mDNS 服务类型
- 解析服务实例的：
  - `Instance`
  - `HostName`
  - `Port`
  - `IPv4`
  - `IPv6`
  - `TXT`
  - `TTL`
- 按 `CIDR` 过滤 IPv4 结果
- 按端口范围过滤服务
- 输出服务列表和聚合资产列表
- 对 `qdiscover` 解析出较深的 banner 字段
- 对 `device-info` 解析 `model`
- 对 `http` 支持基础主动探测

## 当前边界

这版实现还有一些明确边界，建议你使用前先知道：

- mDNS 发现依赖本地链路环境，不保证跨网段发现
- `CIDR` 主要用于结果过滤，不是主动跨网段发起 mDNS 测绘
- `SMB`、`AFP` 目前还没有深入到协议握手级别的 banner 识别
- `HTTP` 主动探测当前只做轻量 GET 请求，不做复杂认证、证书分析或产品特征库比对

## 环境要求

- Go 1.26+
- 能访问本地网络的 mDNS 组播环境

当前依赖：

- `github.com/grandcat/zeroconf`

## 安装依赖

首次拉起项目时执行：

```powershell
go mod tidy
```

## 构建方法

在项目根目录执行：

```powershell
go build ./...
```

或只构建 CLI：

```powershell
go build ./cmd/mdnsmap
```

## 使用方法

### 直接运行

```powershell
go run ./cmd/mdnsmap -timeout 3s
```

### 常用参数

```text
-cidr
  CIDR 过滤范围，例如 192.168.1.0/24

-ports
  端口范围过滤，例如 1-1024 或 80,443,445,5000

-iface
  指定用于 mDNS 发现的网卡名

-timeout
  单次发现超时，例如 3s、5s

-json
  输出 JSON

-active-probe
  是否启用主动探测，默认 true
```

### 示例 1：安静网络中的最小运行

```powershell
go run ./cmd/mdnsmap -timeout 1s -ports 5000
```

如果当前链路内没有可见 mDNS 服务，可能输出：

```text
services:
answers:
PTR:
```

### 示例 2：按网段和端口过滤

```powershell
go run ./cmd/mdnsmap -cidr 192.168.1.0/24 -ports 445,5000,548 -timeout 3s
```

### 示例 3：JSON 输出

```powershell
go run ./cmd/mdnsmap -cidr 192.168.1.0/24 -ports 445,5000,548 -timeout 3s -json
```

### 示例 4：关闭主动探测

```powershell
go run ./cmd/mdnsmap -cidr 192.168.1.0/24 -ports 5000 -timeout 3s -active-probe=false
```

## 输出说明

### 文本输出

文本输出风格接近：

```text
services:
5000/tcp qdiscover:
Name=slw-nas
IPv4=192.168.1.23
IPv6=fe80::265e:beff:fe69:a313
Hostname=slw-nas.local
TTL=10
accessPort=86,accessType=https,displayModel=TS-464C,fwBuildNum=20260214,fwVer=5.2.9,model=TS-X64
answers:
PTR:
_qdiscover._tcp.local
```

### JSON 输出

JSON 输出包含：

- `scope`
- `services`
- `assets`
- `answers`
- `service_types`

其中 `services[].banner.fields` 会保留结构化 banner 字段，适合后续做数据校验和平台接入。

## 测试样本说明

当前仓库已经内置一组离线伪造样本，用于验证：

- 服务标准化
- 端口过滤
- 资产聚合
- 文本输出
- JSON 输出
- `qdiscover` banner 深度

样本文件：

- [testdata/samples/qnap_discovery.json](/d:/project/expr/testdata/samples/qnap_discovery.json:1)

golden 文件：

- [testdata/golden/qnap_full.txt](/d:/project/expr/testdata/golden/qnap_full.txt:1)
- [testdata/golden/qnap_full.json](/d:/project/expr/testdata/golden/qnap_full.json:1)
- [testdata/golden/qnap_filtered.txt](/d:/project/expr/testdata/golden/qnap_filtered.txt:1)
- [testdata/golden/qnap_filtered.json](/d:/project/expr/testdata/golden/qnap_filtered.json:1)

## 测试用例使用方法

### 运行全部测试

```powershell
go test ./...
```

### 只运行离线 golden tests

```powershell
go test ./internal/app -run TestFixture
```

### 只运行全量样本测试

```powershell
go test ./internal/app -run TestFixtureFullOutput -count=1
```

### 只运行端口过滤样本测试

```powershell
go test ./internal/app -run TestFixtureFilteredOutput -count=1
```

## 当前测试覆盖的内容

`TestFixtureFullOutput`

- 加载 `qnap_discovery.json`
- 使用 `CIDR=192.168.1.0/24`
- 使用 `ports=1-65535`
- 对完整输出做 text/json golden 校验
- 校验 `qdiscover` 的 banner depth 至少为 6
- 校验 `displayModel=TS-464C`

`TestFixtureFilteredOutput`

- 加载同一份样本
- 使用 `ports=445,5000,548`
- 验证无端口记录会被过滤掉
- 对过滤后的 text/json 输出做 golden 校验

## 如何新增样本

如果你后续想增加新的设备样本，建议按下面步骤做：

1. 在 `testdata/samples/` 下新增一份 `DiscoveryResult` 格式的 JSON
2. 在 `internal/app/pipeline_test.go` 中新增一个测试
3. 运行同样的处理链路
4. 将实际输出保存到 `testdata/golden/`
5. 将新的 golden 纳入断言

建议优先补这些类型的样本：

- 开启主动 HTTP 探测后的样本
- 多 IP / 多服务实例样本
- 无 IPv4 仅 IPv6 样本
- 非 QNAP 的 `device-info` / `http` 设备样本

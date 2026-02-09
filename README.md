# 个股贪婪恐怖指数 (Fear & Greed Index)

一个轻量、高性能的个股情绪分析工具。输入股票代码，即刻计算 0–100 的“贪婪恐怖分数”，并提供可视化的仪表盘与子指标拆解。

本项目已使用 **Go 语言** 完整重构，提供单一可执行文件，无需复杂的 Python 环境依赖。

## 功能特性

- **多维指标分析**：综合趋势、动量、RSI、MACD、波动率、回撤、量能等 7 大维度。
- **动态归一化**：使用滚动窗口分位数算法，将不同量纲的指标统一映射为 0–100 分。
- **现代化 Dashboard**：内置嵌入式 Web 界面，提供仪表盘、历史曲线、指标矩阵卡片。
- **高性能**：Go 语言原生实现，秒级启动，极低的资源占用。
- **零依赖部署**：所有静态资源（HTML/JS）均编译进二进制文件，下载即用。

## 项目截图

### 📊 主仪表盘

![主仪表盘](https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=stock%20market%20fear%20and%20greed%20index%20dashboard%20with%20large%20score%20display%20in%20center%2C%20green%20color%20for%20greed%2C%20red%20for%20fear%2C%20modern%20web%20interface%2C%20clean%20design%2C%20professional%20financial%20tool&image_size=landscape_16_9)

### 📈 历史趋势

![历史趋势](https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=stock%20market%20fear%20and%20greed%20index%20historical%20trend%20chart%20with%20line%20graph%2C%20time%20series%20data%2C%20green%20and%20red%20areas%2C%20modern%20data%20visualization%2C%20professional%20financial%20dashboard&image_size=landscape_16_9)

### 📋 指标矩阵

![指标矩阵](https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=stock%20market%20technical%20indicators%20matrix%20with%20multiple%20cards%2C%20showing%20RSI%2C%20MACD%2C%20volatility%2C%20momentum%20indicators%2C%20modern%20grid%20layout%2C%20professional%20financial%20dashboard%20design&image_size=landscape_16_9)

## 快速开始

### 🐳 Docker 部署 (推荐)

最简单的方式是直接使用 Docker 运行：

```bash
docker pull iwanlebron/stock-analysis:latest
docker run -d -p 8000:8000 --name stock-analysis iwanlebron/stock-analysis:latest
```

详细 Docker 使用指南请参考 [DOCKER.md](DOCKER.md)。

### 1. 源码运行服务

确保已安装 Go (1.16+)，然后在项目根目录执行：

```bash
# 直接运行
go run main.go

# 或编译为二进制文件运行
go build -o server main.go
./server
```

服务启动后，默认监听 `8000` 端口。

### 2. 使用

打开浏览器访问：[http://localhost:8000](http://localhost:8000)

- 在顶部搜索框输入股票代码（如 `AAPL`, `TSLA`, `NVDA`, `0700.HK`），回车即可分析。
- 点击右上角“设置”可调整分析周期、频率等参数。

### 3. API 调用

后端提供纯 JSON 接口，可供其他服务集成：

```bash
GET /fear-greed?ticker=AAPL&freq=1d&window=252
```

**响应示例**：

```json
{
  "ticker": "AAPL",
  "latest": {
    "score": 78.5,
    "label": "极度贪婪",
    "date": "2024-02-09"
  },
  "latest_subscores": {
    "trend": 85.2,
    "rsi": 91.0,
    ...
  }
}
```

## 指标构成

系统默认包含以下 7 个子指标，加权计算总分：

1. **趋势强度 (20%)**: 价格相对 MA20/MA60 的位置。
2. **动量 (15%)**: 20 日收益率。
3. **RSI (15%)**: 相对强弱指标 (14日)。
4. **MACD (10%)**: MACD 柱状图动能。
5. **回撤压力 (15%)**: 距离 252 日高点的回撤幅度（反向指标）。
6. **波动率 (15%)**: 20 日实现波动率（反向指标）。
7. **量能情绪 (10%)**: 成交量放量时的涨跌方向。

## 开发

项目结构：

```
.
├── main.go          # 服务入口
├── internal/
│   ├── api/         # HTTP API 处理与静态资源嵌入
│   ├── calc/        # 核心算法：指标计算与评分引擎
│   └── models/      # 数据结构定义
├── go.mod           # 依赖管理
├── go.sum           # 依赖校验
├── Dockerfile       # Docker 构建文件
└── docker.sh        # Docker 推送脚本
```

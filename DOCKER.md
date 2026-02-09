# Stock Fear & Greed Index (Docker Usage)

这是一个基于 Go 语言开发的股票市场贪婪恐慌指数分析工具，支持美股、港股、A股及加密货币。

## 🚀 快速开始

### 1. 拉取镜像
```bash
docker pull iwanlebron/stock-analysis:latest
```

### 2. 运行容器
```bash
docker run -d -p 8000:8000 --name stock-analysis iwanlebron/stock-analysis:latest
```

### 3. 访问服务
打开浏览器访问：[http://localhost:8000](http://localhost:8000)

---

## ⚙️ 环境变量配置 (可选)

如果需要修改端口或其他配置，可以在运行容器时传递环境变量：

```bash
docker run -d \
  -p 8080:8000 \
  -e PORT=8000 \
  --name stock-analysis \
  iwanlebron/stock-analysis:latest
```

## 🛠️ 构建自己的镜像

如果你想从源码构建：

```bash
git clone https://github.com/iwanlebron/stock-analysis.git
cd stock-analysis
docker build -t stock-analysis .
docker run -p 8000:8000 stock-analysis
```

## 📊 支持的市场
- **美股 (US)**: SPY, QQQ, NVDA, AAPL 等
- **港股 (HK)**: 恒生指数, 腾讯, 阿里 等
- **A股 (CN)**: 上证指数, 茅台 等
- **加密货币**: BTC, ETH, SOL 等

更多详情请参考 GitHub 仓库。

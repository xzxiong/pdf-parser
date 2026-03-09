# Docker 使用指南

## 构建镜像

```bash
docker build -t pdf-parser .
```

## 使用方法

### 默认行为（Sleep 模式）

容器默认以 sleep 模式运行，会持续运行 1 小时并每 60 秒打印日志：

```bash
docker run --rm pdf-parser
```

自定义报告间隔：

```bash
# 每 30 秒报告一次
docker run --rm --entrypoint /app/pdf-parser pdf-parser --sleep --report-interval 30s

# 持续 2 小时，每 5 分钟报告一次
docker run --rm --entrypoint /app/pdf-parser pdf-parser --sleep 2h --report-interval 5m
```

### 查看帮助信息

```bash
docker run --rm --entrypoint /app/pdf-parser pdf-parser --help
```

### 解析 PDF 文件

需要将 PDF 文件挂载到容器中：

```bash
# 解析当前目录下的 document.pdf
docker run --rm --entrypoint /app/pdf-parser -v $(pwd):/data pdf-parser /data/document.pdf

# 解析后进入 sleep 模式（默认 1 小时）
docker run --rm --entrypoint /app/pdf-parser -v $(pwd):/data pdf-parser --sleep /data/document.pdf

# 自定义 sleep 时间为 30 分钟
docker run --rm --entrypoint /app/pdf-parser -v $(pwd):/data pdf-parser --sleep 30m /data/document.pdf

# 自定义 sleep 时间为 2 小时
docker run --rm --entrypoint /app/pdf-parser -v $(pwd):/data pdf-parser --sleep 2h /data/document.pdf

# 仅提取标题
docker run --rm --entrypoint /app/pdf-parser -v $(pwd):/data pdf-parser --extract-title /data/document.pdf

# 仅提取正文
docker run --rm --entrypoint /app/pdf-parser -v $(pwd):/data pdf-parser --extract-body /data/document.pdf

# 以 JSON 格式输出
docker run --rm --entrypoint /app/pdf-parser -v $(pwd):/data pdf-parser --format json /data/document.pdf

# 启用调试模式
docker run --rm --entrypoint /app/pdf-parser -v $(pwd):/data pdf-parser -d /data/document.pdf
```

### 保存输出到文件

```bash
# 保存为文本文件
docker run --rm --entrypoint /app/pdf-parser -v $(pwd):/data pdf-parser /data/document.pdf > output.txt

# 保存为 JSON 文件
docker run --rm --entrypoint /app/pdf-parser -v $(pwd):/data pdf-parser --format json /data/document.pdf > output.json
```

### 批量处理

```bash
# 处理目录中的所有 PDF 文件
for pdf in *.pdf; do
  docker run --rm --entrypoint /app/pdf-parser -v $(pwd):/data pdf-parser "/data/$pdf" > "${pdf%.pdf}.txt"
done
```

### Sleep 模式说明

Sleep 模式适用于需要保持容器运行的场景（如 Kubernetes 部署）：

- 程序会先处理 PDF 文件并输出结果
- 然后进入 sleep 状态，定期打印运行状态
- 默认持续时间为 1 小时
- 默认报告间隔为 60 秒
- 可以通过参数自定义持续时间（如 30m, 2h, 1h30m）
- 可以通过 `--report-interval` 自定义报告间隔（如 30s, 1m, 5m）
```

## 镜像特点

- 基于 Alpine Linux，镜像体积小
- 多阶段构建，最终镜像只包含必要的运行时文件
- 使用非 root 用户运行，提高安全性
- 静态编译的二进制文件，无外部依赖

## 镜像大小

构建后的镜像大小约为 20-30 MB。

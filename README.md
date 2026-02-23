# bistu-wakeup

北京信息科技大学教务系统课表导出 CLI（WakeUp CSV 导入格式）。

## 下载与发布

请在 GitHub Releases 页面下载预编译二进制文件（不是源码压缩包）：

- `bistu-wakeup-windows-amd64.exe`
- `bistu-wakeup-linux-amd64`
- `bistu-wakeup-darwin-amd64`
- `bistu-wakeup-darwin-arm64`

Release 地址：

`https://github.com/yangqunsheng/bistu-wakeup/releases`

## CLI 使用说明

### 1. 交互式登录（默认）

直接运行程序后，按提示输入学号、密码并选择学期：

```bash
# Windows
./bistu-wakeup-windows-amd64.exe

# Linux
./bistu-wakeup-linux-amd64

# macOS (Intel)
./bistu-wakeup-darwin-amd64

# macOS (Apple Silicon)
./bistu-wakeup-darwin-arm64
```

运行完成后，会在当前目录生成：

`schedule_<term>.csv`

### 2. Cookie 模式（高级用法）

```bash
./bistu-wakeup-linux-amd64 --cookie "JSESSIONID=xxx; route=xxx"
```

参数说明：

- `--cookie`：直接使用浏览器中的教务系统 Cookie 登录

## 导入 WakeUp

1. 在本工具中导出 `schedule_<term>.csv`
2. 打开 WakeUp
3. 选择“导入课表”
4. 选择导出的 CSV 文件

## 源码运行（开发者）

```bash
go mod tidy
go run .
```

## 说明

- `build/` 为 release 产物目录，不纳入代码仓库。
- `.spec-workflow/` 为开发过程文档，不开源。

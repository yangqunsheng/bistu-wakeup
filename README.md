# bistu-wakeup

北京信息科技大学教务系统课表导出工具（WakeUp 课表 CSV 格式）。

## 功能

- 支持学号+密码登录（CAS）
- 支持 `--cookie` 模式登录
- 自动拉取学期并选择导出
- 导出 WakeUp 可导入的 CSV

## 运行环境

- Go 1.22+

## 快速开始

```bash
go mod tidy
go run . 
```

按提示输入学号、密码并选择学期后，会在当前目录生成 `schedule_<term>.csv`。

## Cookie 模式

```bash
go run . --cookie "JSESSIONID=xxx; route=xxx"
```

## 目录说明

- `auth/`：登录与认证
- `schedule/`：课表抓取与解析
- `export/`：CSV 导出

## 说明

- `build/` 目录用于本地 release 产物，不纳入版本控制。
- `.spec-workflow/` 为开发流程资料，不对外开源。

# 芙芙云 Web SSH Terminal (fufussh) 2.0
－ 此版本为普通版无任何后续维护，且如有[魔方财务端]需求请购买专业版3.0 芙云ssh[fufussh购买](https://www.fufuidc.com/cart?fid=22&gid=148)
一个基于 Web 的 SSH 终端管理面板，支持服务器管理、在线终端、文件管理（SFTP）、硬件监控、连接分享与直连等功能。

## 功能特性

- **Web 终端**：基于 WebSocket 的交互式 SSH 终端
- **服务器管理**：添加、编辑、分组管理服务器，支持凭据托管
- **文件管理器**：基于 SFTP 的在线文件浏览、上传、下载
- **硬件监控**：CPU、内存、磁盘、网络等实时监控
- **连接分享**：将服务器连接以受控方式分享给其他用户
- **直连模式**：无需录入服务器信息即可快速连接
- **管理后台**：用户管理、日志审计、系统设置、数据库管理、备份
- **验证码登录**：内置验证码防护

## 技术栈

| 部分 | 技术 |
| --- | --- |
| 后端 | Go + Gin + WebSocket + golang.org/x/crypto/ssh |
| 用户前端 | Vue 3 + Vite + Element Plus |
| 管理后台 | Nuxt 3 + Element Plus |
| 存储 | MySQL（默认，`MYSQL_DSN`）+ JSON 文件存储 |

## 项目结构

```
.
├── backend/          # Go 后端服务
│   ├── main.go
│   ├── handlers/     # HTTP / WebSocket 处理器
│   ├── store/        # MySQL 与 JSON 存储实现
│   ├── models/
│   └── utils/
├── frontend/         # 用户端前端（Vue 3 + Vite）
├── admin/            # 管理后台（Nuxt 3）
├── install.sh        # 一键安装脚本
└── uninstall.sh      # 卸载脚本
```

## 快速开始

### 一键安装（Linux，root）

```bash
bash install.sh
```

安装脚本会自动部署到 `/opt/web-ssh` 并完成服务配置。

### 手动运行后端

```bash
cd backend
export MYSQL_DSN="user:password@tcp(127.0.0.1:3306)/ssh-web"
export PORT=8080
go run main.go
```

- `PORT`：服务监听端口，默认 `8080`
- `MYSQL_DSN`：MySQL 连接串，需确保数据库 `ssh-web` 已创建

启动后：

- 用户端 API / 页面：`http://localhost:8080`
- 管理后台：`http://localhost:8080/admin`

### 前端开发

```bash
# 用户端
cd frontend
npm install
npm run dev

# 管理后台
cd admin
npm install
npm run dev
```

## URL 快速连接

本系统支持通过 URL 直接打开 Web 终端并连接到指定服务器，无需注册登录或录入服务器信息，适合集成到面板、工单系统等外部平台（例如在主机详情页放一个"SSH 快速连接"按钮）。

### 接口格式

```
GET {你的部署地址}/direct/connect?host={host}&port={port}&username={username}&password={password}
```

### 参数说明

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `host` | 是 | 服务器地址，IP 或域名。IPv6 或带端口的地址请使用纯地址，端口通过 `port` 传 |
| `port` | 否 | SSH 端口，默认 `22` |
| `username` | 是 | SSH 登录用户名 |
| `password` | 是 | SSH 登录密码 |

所有参数值必须经过 URL 编码（`encodeURIComponent`），因为密码中常含有 `@`、`#`、`&` 等特殊字符。

## 卸载

```bash
bash uninstall.sh
```

## 开源协议

本项目基于 [PolyForm Noncommercial License 1.0.0](./LICENSE) 开源。

仅授权非商业用途使用（个人使用、学习研究、公益组织等）。任何商业用途（包括销售、付费服务、商业产品集成等）需要另行获得版权所有者的商业授权。使用或分发本项目时请附带本协议全文。

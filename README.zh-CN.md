<div align="center">

# Antigravity Remote

**在手机上使用 Antigravity 桌面版。**

[![release](https://img.shields.io/github/v/release/AFSlayer/antigravity-remote?style=flat-square&color=4f7cff)](https://github.com/AFSlayer/antigravity-remote/releases/latest)
[![license](https://img.shields.io/badge/license-Apache--2.0-blue?style=flat-square)](LICENSE)

<img src="docs/assets/demo.gif" width="300" alt="在手机上让 agent 查看服务器状态" />

[English](README.md) · [한국어](README.ko.md) · [日本語](README.ja.md) · [Português](README.pt-BR.md) · [Español](README.es.md)

</div>

## 是什么

Antigravity 桌面版里带了一个 `language_server`。跟 Google 通信的是它，加上
`--standalone` 之后它还会把整个 Antigravity 界面当成 web 应用提供出来。它只监听
`127.0.0.1`。

`agy-remote` 在它前面加一道密码，转发到你的网络。顺便改写经过的 JS bundle 里的几个
字符串，因为桌面 IDE 直接放到手机浏览器里有些地方不好用。

做同类事情的项目都是自己做界面，或者用 CDP 镜像屏幕。这个直接提供 Antigravity 自己的
界面，所以终端、文件树、artifacts、browser agent 都能用，Google 出新功能也会自己出现。

代价是给压缩后的 bundle 打补丁很脆弱。Antigravity 一更新补丁可能就失效。`agy-remote`
每次启动都检查一遍，哪个不匹配会告诉你。

## 安装

```bash
# macOS、Linux
curl -fsSL https://raw.githubusercontent.com/AFSlayer/antigravity-remote/main/scripts/install-desktop.sh | bash
```

```powershell
# Windows
irm https://raw.githubusercontent.com/AFSlayer/antigravity-remote/main/scripts/install-desktop.ps1 | iex
```

装完就启动。控制面板会打开并显示二维码，手机扫一下就进去了，不用输密码。二维码里是一个
十分钟有效的一次性 token。

<div align="center">
<img src="docs/assets/control-panel.png" width="320" alt="控制面板" />
</div>

不用事先打开 Antigravity。没运行的话 `agy-remote` 会启动它、打开远程控制、等 language
server 就绪，再找出哪个端口在提供界面。

手机上用分享 →“添加到主屏幕”，会以 Antigravity 图标全屏打开，没有地址栏。

二进制文件没有代码签名，所以用浏览器下载压缩包会被系统隔离。macOS 右键 **打开** 再
**打开**；Windows 点 **更多信息** → **仍要运行**。上面的安装命令用 `curl`，不会被打上
隔离标记。

## 部署到服务器

放到 Linux 机器上，合上笔记本 Antigravity 也照样跑。便宜 VPS 就够，我用的是 Oracle 免费
ARM 实例。

```bash
curl -fsSL https://raw.githubusercontent.com/AFSlayer/antigravity-remote/main/scripts/install.sh | bash
```

它会问域名和工作目录，从 Google 的 `storage.googleapis.com` 下载官方构建包，只取出
165MB 的 `language_server`。本仓库不二次分发任何 Google 的二进制文件。然后写 systemd
unit、配好 Caddy 的 HTTPS、生成密码。

因为是同一套 web 界面，用电脑浏览器打开也一样，外观和操作跟桌面版没区别。对话、工作区、
正在跑的 agent 都在服务器上，所以在地铁上用手机开始的活儿，到公司用台式机可以接着做。
只有一个实例，所以根本不存在同步这回事。

有一步没法自动化。Antigravity 登录走 `localhost` 上的 OAuth 回调，远端服务器收不到。要从
一台已经在用桌面版的机器上把 token 拷过去：

```bash
scp ~/.gemini/jetski-standalone-oauth-token you@your-server:~/.gemini/
```

token 缺失时 `agy-remote` 会把这条命令打出来。

## 打了哪些补丁

一共 13 个，说明都在
[`internal/patches/registry.go`](internal/patches/registry.go)。用 `agy-remote doctor`
看哪些生效了。

| 问题 | 补丁 |
| --- | --- |
| bundle 调用 `https://127.0.0.1:<port>`，在手机上那就是手机 | 改用浏览器的 origin |
| 句子没写完按回车就发出去了 | 触屏上回车是换行，Cmd/Ctrl+回车才发送 |
| iOS 底部 home 条挡住输入框 | 处理 `safe-area-inset-bottom`，键盘弹出时去掉留白 |
| 点模型就按 medium 选中并关菜单 | 点开推理强度子菜单 |
| 第一次回复后弹出 “Enable Notifications” 横幅 | 触屏设备上跳过 |
| standalone 没有语音转写，麦克风按钮却还在 | 隐藏 |
| 新建项目从 `/` 开始 | 从配置的工作目录开始 |
| 没图标、300ms 点击延迟、浏览器外壳 | 图标、即时响应、manifest 全屏 |

想看原样界面用 `agy-remote --no-mobile-patches`。

## 安全

能进来的人可以读你的文件、执行命令，这更接近给出 shell 权限。

- 密码用 PBKDF2-SHA256 迭代 20 万次，明文不落盘。
- 会话 token 是 256 位随机数，磁盘上只有哈希，拷走 `sessions.json` 也没用。
- 每个 IP 五分钟五次失败，之后锁定时间翻倍到 30 分钟，正确密码同样被拦。
- 带二维码和关闭按钮的控制面板只在单独端口监听 loopback，不对外。

放在反向代理后面要指定可信来源，否则 forwarded 头会被忽略：

```bash
agy-remote serve --public-url https://agy.example.com --trusted-proxies 127.0.0.1/32
```

其余见 [SECURITY.md](SECURITY.md)。能用 Tailscale 就用，不能就上 HTTPS。

## 命令

```
agy-remote                     把桌面版共享到你的网络
agy-remote serve               在服务器上无界面运行
agy-remote doctor              全面检查并指出问题
agy-remote config [flags]      把选项写进 config.json
agy-remote passwd [password]   设置密码
agy-remote sessions [revoke]   查看或登出设备
```

参数用 `agy-remote help` 看，每个都有对应的 `AGY_*` 环境变量。

## 其他

只能用名字就叫 “Antigravity” 的桌面版，IDE 和 CLI 都不行，因为只有它能跑
`language_server --standalone`。你的代码不会离开本机：代理和 language server 都在同一台
机器上。开发说明、常见问题和完整的安全文档在 [英文 README](README.md) 里。

[Apache-2.0](LICENSE)。这不是 Google 的项目。请看 [DISCLAIMER.md](DISCLAIMER.md)。

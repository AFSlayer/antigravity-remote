<div align="center">

# Antigravity Remote

**在手机上使用 Antigravity 桌面版。**

[![release](https://img.shields.io/github/v/release/AFSlayer/antigravity-remote?style=flat-square&color=4f7cff)](https://github.com/AFSlayer/antigravity-remote/releases/latest)
[![ci](https://img.shields.io/github/actions/workflow/status/AFSlayer/antigravity-remote/ci.yml?branch=main&style=flat-square)](https://github.com/AFSlayer/antigravity-remote/actions/workflows/ci.yml)
[![license](https://img.shields.io/badge/license-Apache--2.0-blue?style=flat-square)](LICENSE)

<img src="docs/assets/demo.gif" width="300" alt="在手机上让 agent 查看服务器状态" />

<sub>手机 Safari 里的 Antigravity，正在服务器上执行 shell 命令。</sub>

[English](README.md) · [한국어](README.ko.md) · [日本語](README.ja.md) · [Português](README.pt-BR.md) · [Español](README.es.md)

</div>

## 这是什么

Antigravity 桌面版里带了一个叫 `language_server` 的可执行文件。真正跟 Google 通信的
是它，而且加上 `--standalone` 之后，它还会把整个 Antigravity 界面当成 web 应用提供出来。
问题是它只监听 `127.0.0.1`，所以除了桌面版自己，谁都连不上。

`agy-remote` 在它前面加了一道密码，并把服务转发到你的网络里。它还会改写经过的 JS
bundle 里的几个字符串，因为桌面版 IDE 直接放到手机浏览器里用，有些地方确实别扭。

让 Antigravity 上手机的项目已经有十几个了，做法都是自己做界面：要么写一个聊天面板，
要么用 CDP 做屏幕镜像。这个项目反过来，直接提供 Antigravity 自己的界面。所以终端能用、
文件树能用、artifacts 和 browser agent 也能用，Google 加了新功能会自己出现。这些都不是
我写的，我也不用跟着做。

代价是给压缩后的 bundle 打补丁本身就很脆弱。Antigravity 一更新，补丁可能就失效了。
`agy-remote` 每次启动都会检查所有补丁，哪个不匹配会直接告诉你。

## 安装

```bash
# macOS、Linux
curl -fsSL https://raw.githubusercontent.com/AFSlayer/antigravity-remote/main/scripts/install-desktop.sh | bash
```

```powershell
# Windows
irm https://raw.githubusercontent.com/AFSlayer/antigravity-remote/main/scripts/install-desktop.ps1 | iex
```

装完直接启动。浏览器里会打开一个控制面板，上面有二维码，手机扫一下就进去了。不用输
密码，二维码里带的是一个十分钟有效的一次性 token。

<div align="center">
<img src="docs/assets/control-panel.png" width="320" alt="控制面板" />
</div>

不需要先把 Antigravity 打开。如果它没在运行，`agy-remote` 会启动它、在设置里打开远程
控制、等 language server 就绪，然后找出它的哪个端口在提供界面。

手机上建议点分享 →“添加到主屏幕”。这样会以 Antigravity 图标全屏打开，没有浏览器地址栏。
本文里的截图就是这个状态。

### 关于安全警告

我没有买代码签名证书，Apple 和 Microsoft 都是按年收费的。所以如果你用浏览器下载压缩包，
系统会把它隔离：

- macOS 会说无法验证开发者。右键点文件选**打开**，再点一次**打开**。或者执行
  `xattr -d com.apple.quarantine agy-remote`。
- Windows 会弹 SmartScreen。点**更多信息** → **仍要运行**。

`curl` 和 `Invoke-WebRequest` 下载的文件不会被打上隔离标记，所以用上面的安装命令就完全
碰不到这些。

## 部署到服务器

放到一台 Linux 机器上，它就变成一个合上笔记本也照样干活的 Antigravity。便宜的 VPS 就够。
我自己用的是 Oracle 免费的 ARM 实例。

```bash
curl -fsSL https://raw.githubusercontent.com/AFSlayer/antigravity-remote/main/scripts/install.sh | bash
```

它会问你域名和工作目录，然后从 Google 的 `storage.googleapis.com` 下载官方 Antigravity
构建包，只取出 165MB 的 `language_server`。这个仓库不会二次分发任何 Google 的二进制文件。
接着它会写 systemd unit、配好 Caddy 自动签发 HTTPS，并生成一个密码。

有一步没法自动化。Antigravity 登录走的是 `localhost` 上的 OAuth 回调，远端服务器收不到
这个回调。你需要从一台已经在用桌面版的机器上把 token 拷过去：

```bash
scp ~/.gemini/jetski-standalone-oauth-token you@your-server:~/.gemini/
```

token 不存在时，`agy-remote` 会把这条命令直接打印出来。

## 打了哪些补丁

一共十二个，每个都在
[`internal/patches/registry.go`](internal/patches/registry.go) 里写了说明。用
`agy-remote doctor` 可以看到哪些生效了。

| 问题 | 补丁 |
| --- | --- |
| bundle 调用 `https://127.0.0.1:<port>`，在手机上那就是手机自己 | 改用浏览器自身的 origin |
| 句子还没写完，一按回车就发出去了 | 触屏设备上回车是换行，Cmd/Ctrl+回车才发送 |
| iOS 底部 home 条挡住输入框 | 处理 `safe-area-inset-bottom`，键盘弹出时去掉留白 |
| 点一下模型就按 medium 选中并关掉菜单 | 改成点开推理强度子菜单 |
| standalone 模式没有语音转写，但麦克风按钮还在 | 隐藏掉 |
| 新建项目从 `/` 开始 | 从你配置的工作目录开始 |
| 没有应用图标、300ms 点击延迟、浏览器外壳 | 加图标、点击即时响应、加 manifest 实现全屏 |

想看原样的界面就用 `agy-remote --no-mobile-patches`。

<div align="center">
<table><tr>
<td align="center"><img src="docs/assets/patch-models.png" width="190" alt="模型选择" /></td>
<td align="center"><img src="docs/assets/patch-effort.png" width="190" alt="推理强度" /></td>
<td align="center"><img src="docs/assets/settings.png" width="190" alt="设置" /></td>
</tr></table>
</div>

## 安全

能连上 Antigravity 的人就能读你的文件、执行命令，所以这跟给出 shell 权限差不多。

- 密码用 PBKDF2-SHA256 迭代 20 万次哈希，明文不落盘。
- 会话 token 是 256 位随机数，磁盘上只存哈希，所以 `sessions.json` 被拷走也没用。
  `agy-remote sessions revoke` 可以把所有设备踢下线。
- 登录限制为每个 IP 五分钟内五次失败，之后锁定时间翻倍，最长 30 分钟。另有一个全局
  限流应对分布式尝试。锁定期间就算密码对也一样被拦。
- Cookie 带 `HttpOnly`、`SameSite=Lax`，请求走 HTTPS 时还会加 `Secure`。
- 带二维码和关闭按钮的控制面板只在单独端口上监听 loopback，从不暴露到网络。

放在反向代理后面时要指定可信的来源，否则 forwarded 头会被忽略：

```bash
agy-remote serve --public-url https://agy.example.com --trusted-proxies 127.0.0.1/32
```

其余内容在 [SECURITY.md](SECURITY.md)。简单说：能用 Tailscale 就用，不能用就至少上 HTTPS。

## 命令

```
agy-remote                     把桌面版共享到你的网络
agy-remote serve               在服务器上以无界面模式运行
agy-remote doctor              全面检查并指出问题
agy-remote config [flags]      把选项写进 config.json
agy-remote passwd [password]   设置密码
agy-remote sessions [revoke]   查看或登出设备
```

常用参数：`--port`、`--public-url`、`--workspace-root`、`--trusted-proxies`、
`--session-days`、`--no-mobile-patches`、`--language-server`。每个都有对应的 `AGY_*`
环境变量，`agy-remote help` 会全部列出。

## 工作方式

```
   手机                        你的电脑或服务器
┌──────────┐              ┌────────────────────────────────┐
│  浏览器  │    密码      │ agy-remote                     │
│          │◄────────────►│   会话、限流                   │
│  Anti-   │    :8765     │   patch main.js / index.html   │
│ gravity  │              │              │ https           │
│   界面   │              │   language_server --standalone │
└──────────┘              └──────────────┼─────────────────┘
                                         ▼
                                  Google CloudCode
```

你的 prompt 和代码只是被代理转发的字节，除了同一台机器上的 language server 不会去别处。
Antigravity 跟 Google 之间的通信没有任何改动。

## 常见问题

**用 IDE 或 CLI 可以吗？**
不行，必须是名字就叫 “Antigravity” 的桌面版。只有它能跑
`language_server --standalone`，也就是提供 web 界面的那个模式。CLI 里虽然编进了 bundle，
但没有把它提供出来的参数。

**Antigravity 更新后还能用吗？**
代理本身没问题，个别补丁可能失效。远程访问真正必需的只有 `base-url-origin`，其余都是
体验优化。失效了欢迎提 issue。

**我的代码会被传到别处吗？**
不会。代理和 language server 都跑在你自己的机器上。

**为什么局域网里用 HTTP 没问题，公网上就不行？**
手机浏览器不接受自签证书，而 `192.168.x.x` 又拿不到正式证书。在可信网络里这个取舍可以
接受，公网上就不行，所以要用 `--public-url` 配合 Caddy 或者隧道。

## 开发

```bash
go test ./...
go run ./cmd/agy-remote
```

值得看的是 [`internal/patches`](internal/patches)。补丁就是一个带锚点字符串的结构体，
往 `All()` 里加一个就行，测试、`doctor`、控制面板和缓存键都从这个列表读。

仓库里没有 Antigravity 的 bundle，所以补丁分两层测试：`patches_test.go` 用注册表生成
一个合成文档来测引擎，`live_test.go` 从运行中的 language server 拉真实 bundle，断言每个
锚点恰好匹配一次（没在运行就跳过）。发版前打开桌面版跑一下：

```bash
go test ./internal/patches -run Live -v
```

## 许可

[Apache-2.0](LICENSE)。这不是 Google 的项目，也与 Google 无关。请看
[DISCLAIMER.md](DISCLAIMER.md)。

<div align="center">

# Antigravity Remote

**Antigravity デスクトップ版をスマホから使う。**

[![release](https://img.shields.io/github/v/release/AFSlayer/antigravity-remote?style=flat-square&color=4f7cff)](https://github.com/AFSlayer/antigravity-remote/releases/latest)
[![ci](https://img.shields.io/github/actions/workflow/status/AFSlayer/antigravity-remote/ci.yml?branch=main&style=flat-square)](https://github.com/AFSlayer/antigravity-remote/actions/workflows/ci.yml)
[![license](https://img.shields.io/badge/license-Apache--2.0-blue?style=flat-square)](LICENSE)

<img src="docs/assets/demo.gif" width="300" alt="スマホからエージェントにサーバーの状態を聞いている様子" />

<sub>モバイル Safari で動く Antigravity。サーバー上でシェルコマンドを実行中。</sub>

[English](README.md) · [한국어](README.ko.md) · [中文](README.zh-CN.md) · [Português](README.pt-BR.md) · [Español](README.es.md)

</div>

## これは何か

Antigravity デスクトップ版には `language_server` というバイナリが入っている。Google と
実際に通信しているのはこれで、`--standalone` を付けて起動すると Antigravity の UI 全体を
web アプリとして配信もする。ただし `127.0.0.1` しか listen しないので、デスクトップ版
以外からは誰も届かない。

`agy-remote` はその手前にパスワードを置いて、自分のネットワークに転送する。通り道で JS
バンドルの文字列もいくつか書き換える。デスクトップ用 IDE をそのままスマホのブラウザで
使うと、引っかかる箇所がいくつかあるからだ。

Antigravity をスマホで使うためのプロジェクトはすでに何本もあって、どれも UI を作っている。
独自のチャットパネルを書くか、CDP で画面をミラーリングするかだ。これは逆に、Antigravity
自身の UI をそのまま配信する。だからターミナルも動くし、ファイルツリーも動くし、
artifacts や browser agent も動く。Google が新機能を出せば勝手に付いてくる。自分で書いて
いないので、追いかける必要もない。

代わりに、minify 済みのバンドルにパッチを当てるのは壊れやすい。Antigravity が更新されると
パッチが当たらなくなることがある。`agy-remote` は起動ごとに全パッチを検査して、合わなく
なったものを知らせる。

## インストール

```bash
# macOS、Linux
curl -fsSL https://raw.githubusercontent.com/AFSlayer/antigravity-remote/main/scripts/install-desktop.sh | bash
```

```powershell
# Windows
irm https://raw.githubusercontent.com/AFSlayer/antigravity-remote/main/scripts/install-desktop.ps1 | iex
```

インストールしてそのまま起動する。ブラウザにコントロールパネルが開いて QR コードが出る。
スマホで読み取れば入れる。パスワードを打つ必要はない。QR には 10 分間有効なワンタイム
トークンが入っている。

<div align="center">
<img src="docs/assets/control-panel.png" width="320" alt="コントロールパネル" />
</div>

Antigravity を先に開いておく必要はない。起動していなければ `agy-remote` がアプリを立ち
上げ、設定でリモートコントロールを有効にし、language server を待ってから、どのポートが
UI を配信しているかを突き止める。

スマホでは共有 →「ホーム画面に追加」を押すといい。ブラウザのアドレスバーなしで
Antigravity のアイコンから全画面で開く。このページのスクリーンショットはその状態だ。

### セキュリティ警告について

コード署名はしていない。Apple も Microsoft も署名に年額を取るからだ。そのためブラウザで
アーカイブを直接ダウンロードすると OS が隔離する。

- macOS は開発者を確認できないと言う。ファイルを右クリックして**開く**、もう一度**開く**。
  または `xattr -d com.apple.quarantine agy-remote`。
- Windows は SmartScreen が出る。**詳細情報** → **実行**。

`curl` や `Invoke-WebRequest` で取得したファイルには隔離属性が付かない。上のインストール
コマンドを使えば、この手順自体に出会わない。

## サーバーに置く

Linux マシンに置けば、ノートを閉じても動き続ける Antigravity になる。安い VPS で十分だ。
自分のは Oracle の無料 ARM インスタンスで動いている。

```bash
curl -fsSL https://raw.githubusercontent.com/AFSlayer/antigravity-remote/main/scripts/install.sh | bash
```

ドメインとワークスペースのフォルダを聞いた後、Google の `storage.googleapis.com` から
公式の Antigravity ビルドを取得し、165MB の `language_server` だけを取り出す。このリポジ
トリが Google のバイナリを再配布することはない。続いて systemd unit を書き、Caddy で
HTTPS を自動設定し、パスワードを生成する。

一つだけ自動化できない部分がある。Antigravity のサインインは `localhost` 上の OAuth
コールバックを使うので、リモートのサーバーはそれを受け取れない。デスクトップ版を使って
いるマシンからトークンをコピーする必要がある。

```bash
scp ~/.gemini/jetski-standalone-oauth-token you@your-server:~/.gemini/
```

トークンが無いときは `agy-remote` がこのコマンドをそのまま表示する。

## 何をパッチしているか

全部で 12 個。それぞれの説明は
[`internal/patches/registry.go`](internal/patches/registry.go) にある。どれが適用されたかは
`agy-remote doctor` で確認できる。

| 問題 | パッチ |
| --- | --- |
| バンドルが `https://127.0.0.1:<port>` を呼ぶ。スマホではそれはスマホ自身 | ブラウザ自身の origin を使う |
| 文章の途中で Enter を押すと送信されてしまう | タッチ端末では Enter は改行、Cmd/Ctrl+Enter で送信 |
| iOS のホームバーが入力欄に被る | `safe-area-inset-bottom` を反映し、キーボード表示中は余白を消す |
| モデルをタップすると medium で選択されメニューが閉じる | タップで推論レベルのサブメニューを開く |
| standalone では文字起こしが無いのにマイクボタンがある | 隠す |
| 新規プロジェクトが `/` から始まる | 設定したワークスペースフォルダから始める |
| アプリアイコン無し、300ms のタップ遅延、ブラウザの枠 | アイコン、即時反応、manifest で全画面 |

素の UI が見たいときは `agy-remote --no-mobile-patches`。

<div align="center">
<table><tr>
<td align="center"><img src="docs/assets/patch-models.png" width="190" alt="モデル選択" /></td>
<td align="center"><img src="docs/assets/patch-effort.png" width="190" alt="推論レベル" /></td>
<td align="center"><img src="docs/assets/settings.png" width="190" alt="設定" /></td>
</tr></table>
</div>

## セキュリティ

Antigravity に届く人はファイルを読めてコマンドも実行できる。shell を渡すのと同じくらいの
話だと考えてほしい。

- パスワードは PBKDF2-SHA256 を 20 万回で hash する。平文はどこにも書かない。
- セッショントークンは 256 ビットの乱数で、ディスクには hash しか書かない。だから
  `sessions.json` を持ち出されても使えない。`agy-remote sessions revoke` で全端末を
  ログアウトできる。
- ログインは IP ごとに 5 分で 5 回失敗まで。その後は最大 30 分まで倍々になるロックが
  かかる。分散試行用にグローバルな制限もある。ロック中は正しいパスワードでも通らない。
- Cookie は `HttpOnly`、`SameSite=Lax`。HTTPS で来たリクエストなら `Secure` も付く。
- QR コードと停止ボタンのあるコントロールパネルは、別ポートで loopback だけを listen
  する。ネットワークには出さない。

リバースプロキシの後ろに置くなら、信頼する peer を指定する。指定しないと forwarded
ヘッダは無視される。

```bash
agy-remote serve --public-url https://agy.example.com --trusted-proxies 127.0.0.1/32
```

残りは [SECURITY.md](SECURITY.md) に書いた。要するに、可能なら Tailscale の中に置き、
無理なら HTTPS を使うという話だ。

## コマンド

```
agy-remote                     デスクトップ版を自分のネットワークに共有
agy-remote serve               サーバーでヘッドレス実行
agy-remote doctor              全体を点検して問題を出力
agy-remote config [flags]      オプションを config.json に保存
agy-remote passwd [password]   パスワードを設定
agy-remote sessions [revoke]   端末の一覧 / 全ログアウト
```

よく使うフラグは `--port`、`--public-url`、`--workspace-root`、`--trusted-proxies`、
`--session-days`、`--no-mobile-patches`、`--language-server`。すべて `AGY_*` の環境変数が
あり、`agy-remote help` に全部載っている。

## 仕組み

```
   スマホ                     自分のマシン / サーバー
┌──────────┐              ┌────────────────────────────────┐
│ ブラウザ │  パスワード  │ agy-remote                     │
│          │◄────────────►│   セッション、レート制限       │
│  Anti-   │    :8765     │   main.js / index.html を patch │
│ gravity  │              │              │ https           │
│   UI     │              │   language_server --standalone │
└──────────┘              └──────────────┼─────────────────┘
                                         ▼
                                  Google CloudCode
```

プロンプトとコードはプロキシされるバイト列として通るだけで、同じホストの language server
以外には行かない。Antigravity と Google の通信はそのままだ。

## よくある質問

**IDE や CLI でもできる？**
できない。名前がそのまま「Antigravity」のデスクトップ版が必要だ。web UI を配信する
`language_server --standalone` を動かせるのはそれだけ。CLI のバイナリにもバンドルは
入っているが、配信するフラグが無い。

**Antigravity が更新されても使える？**
プロキシは動く。個々のパッチは外れることがある。リモートアクセスに本当に必要なのは
`base-url-origin` だけで、他は快適さのためのものだ。外れたら issue を立ててほしい。

**コードはどこかに送られる？**
送られない。プロキシも language server も自分のマシンで動く。

**LAN では HTTP でいいのに、インターネットではだめなのは？**
スマホのブラウザは自己署名証書を受け付けないし、`192.168.x.x` に正式な証明書は取れない。
信頼できるネットワークなら許容できる取引だが、公開アドレスでは違う。`--public-url` と
Caddy かトンネルを使ってほしい。

## 開発

```bash
go test ./...
go run ./cmd/agy-remote
```

読む価値があるのは [`internal/patches`](internal/patches)。パッチはアンカー文字列を持つ
構造体で、`All()` に一つ足せば済む。テスト、`doctor`、コントロールパネル、キャッシュキーが
すべてそのリストを読む。

Antigravity のバンドルはこのリポジトリに入っていないので、パッチは 2 段構えでテストする。
`patches_test.go` はレジストリから合成ドキュメントを作ってエンジンを試し、`live_test.go` は
動いている language server から実物のバンドルを取ってきて、各アンカーがちょうど 1 回だけ
一致することを確認する（動いていなければ skip）。リリース前にデスクトップ版を開いて実行:

```bash
go test ./internal/patches -run Live -v
```

## ライセンス

[Apache-2.0](LICENSE)。Google のプロジェクトではなく、Google とは無関係だ。
[DISCLAIMER.md](DISCLAIMER.md) も見てほしい。

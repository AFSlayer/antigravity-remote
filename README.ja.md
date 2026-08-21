<div align="center">

# Antigravity Remote

**Antigravity デスクトップ版をスマホから使う。**

[![release](https://img.shields.io/github/v/release/AFSlayer/antigravity-remote?style=flat-square&color=4f7cff)](https://github.com/AFSlayer/antigravity-remote/releases/latest)
[![license](https://img.shields.io/badge/license-Apache--2.0-blue?style=flat-square)](LICENSE)

<img src="docs/assets/demo.gif" width="300" alt="スマホからエージェントにサーバーの状態を聞いている様子" />

[English](README.md) · [한국어](README.ko.md) · [中文](README.zh-CN.md) · [Português](README.pt-BR.md) · [Español](README.es.md)

</div>

## これは何

Antigravity デスクトップ版には `language_server` というバイナリが入っている。Google と
通信しているのはこれで、`--standalone` を付けると Antigravity の UI 全体を web アプリと
しても配信する。listen するのは `127.0.0.1` だけだ。

`agy-remote` はその手前にパスワードを置いて、手元のネットワークに転送する。通り道で JS
バンドルの文字列もいくつか書き換える。デスクトップ用 IDE をスマホのブラウザで使うと
引っかかる箇所があるからだ。

同じことをやっているプロジェクトは自前の UI を作るか、CDP で画面をミラーリングする。
これは Antigravity 自身の UI を配信する。だからターミナルもファイルツリーも artifacts も
browser agent も動くし、Google が新機能を出せば勝手に付いてくる。

代わりに、minify 済みのバンドルへのパッチは壊れやすい。Antigravity が更新されると当たら
なくなることがある。`agy-remote` は起動時に全部を検査して、外れたものを知らせる。

## インストール

```bash
# macOS、Linux
curl -fsSL https://raw.githubusercontent.com/AFSlayer/antigravity-remote/main/scripts/install-desktop.sh | bash
```

```powershell
# Windows
irm https://raw.githubusercontent.com/AFSlayer/antigravity-remote/main/scripts/install-desktop.ps1 | iex
```

入れたらそのまま起動する。コントロールパネルが開いて QR コードが出るので、スマホで読み
取れば入れる。パスワードは打たない。QR には 10 分有効のワンタイムトークンが入っている。

<div align="center">
<img src="docs/assets/control-panel.png" width="320" alt="コントロールパネル" />
</div>

Antigravity を先に開く必要はない。起動していなければ `agy-remote` がアプリを立ち上げ、
リモートコントロールを有効にし、language server を待って、どのポートが UI を配信して
いるかを調べる。

スマホでは共有 →「ホーム画面に追加」を使うといい。ホーム画面に Antigravity の
アイコンが並ぶ。

バイナリにコード署名は無いので、ブラウザでアーカイブを落とすと OS が隔離する。macOS は
右クリック → **開く** → もう一度 **開く**。Windows は **詳細情報** → **実行**。上の
インストールコマンドは `curl` を使うので隔離属性が付かない。

## サーバーに置く

Linux マシンに置けば、ノートを閉じても Antigravity は動き続ける。安い VPS や無料枠の
ARM インスタンスで足りる。

```bash
curl -fsSL https://raw.githubusercontent.com/AFSlayer/antigravity-remote/main/scripts/install.sh | bash
```

ドメインとワークスペースのフォルダを聞いて、Google の `storage.googleapis.com` から公式
ビルドを取得し、165MB の `language_server` だけを抜き出す。このリポジトリが Google の
バイナリを再配布することはない。あとは systemd unit を書き、Caddy で HTTPS を設定し、
パスワードを作る。

同じ web UI なので PC のブラウザからでも使える。見た目も操作もデスクトップ版と変わらない。
会話もワークスペースも動いているエージェントもサーバー側にあるから、電車でスマホで始めた
作業をそのままデスクトップで続けられる。インスタンスは一つなので同期という概念が無い。

一つだけ自動化できない。Antigravity のサインインは `localhost` の OAuth コールバックを
使うので、リモートのサーバーは受け取れない。デスクトップ版を使っているマシンからトークンを
コピーする。

```bash
scp ~/.gemini/jetski-standalone-oauth-token you@your-server:~/.gemini/
```

トークンが無いときは `agy-remote` がこのコマンドを表示する。

## パッチ一覧
 
25個のパッチ。それぞれ[`internal/patches/registry.go`](internal/patches/registry.go)に記載。
適用状態は`agy-remote doctor`で確認。

| 問題 | パッチ |
| --- | --- |
| バンドルが`https://127.0.0.1:<port>`を呼ぶ（スマホからは自分自身になる） | ブラウザのoriginを使用 |
| 入力途中でEnterを押すと送信される | タッチではEnterで改行、Cmd/Ctrl+Enterで送信 |
| モデルをタップするとmediumが選択されメニューが閉じる | タップでeffortサブメニューを開く |
| 最初の返答で「通知を有効にする」バナーが出る | タッチ端末ではスキップ |
| スタンドアロンでは文字起こしがないのにマイクボタンがある | 非表示にする |
| モバイルタイトルバーの動かないプロフィールアイコン | 非表示にする |
| モバイルで`+`をタップしても会話一覧にとどまる | 新しい会話画面に直接遷移しタイトルバーに戻るボタンを統合 |
| 新規プロジェクトが`/`で始まる | 設定したワークスペースフォルダで開始 |
| 大容量ファイル（.har、ログ等）でブラウザがフリーズまたは20MB制限エラー | ワークスペースへの非同期ストリーミングアップロードと進捗UI |
| 非標準拡張子（.har、.jsonl等）の添付拒否 | すべてのファイルタイプを許可しテキスト/データを解析 |
| アイコンなし、300ミリ秒のタップ遅延 | Antigravityアイコン、即座に反応 |
| iOSのホームバー重なり＆仮想キーボードで画面が押し出される | Safe Areaインセット＆ビューポート高さ同期 |
| リモートブラウザからGoogleログインできない | 設定のログインボタンをWeb認証フローに接続 |

素の UI が見たいときは `agy-remote --no-mobile-patches`。

## セキュリティ

入れた人はファイルを読めてコマンドも実行できる。ドキュメント共有より shell を渡すのに
近い。

- パスワードは PBKDF2-SHA256 を 20 万回。平文はどこにも置かない。
- セッショントークンは 256 ビットの乱数で、ディスクには hash だけ。`sessions.json` を
  持ち出されても使えない。
- IP ごとに 5 分で 5 回失敗まで、その後は最大 30 分まで倍々のロック。正しいパスワードも
  同じく弾かれる。
- QR コードと停止ボタンのあるコントロールパネルは別ポートで loopback だけを listen する。
  外には出さない。

リバースプロキシの後ろでは信頼する peer を指定する。指定しないと forwarded ヘッダは
無視される。

```bash
agy-remote serve --public-url https://agy.example.com --trusted-proxies 127.0.0.1/32
```

残りは [SECURITY.md](SECURITY.md)。可能なら Tailscale、無理なら HTTPS。

## コマンド

```
agy-remote                     デスクトップ版をローカルネットワークに共有
agy-remote serve               サーバーでヘッドレス実行
agy-remote update [flags]      Antigravity language_server を最新の公式リリースに更新
agy-remote doctor              全体を点検して問題を出力
agy-remote config [flags]      オプションを config.json に保存
agy-remote passwd [password]   パスワードを設定
agy-remote sessions [revoke]   端末の一覧 / 全ログアウト
```

フラグは `agy-remote help` に出る。すべて `AGY_*` の環境変数がある。

## その他

必要なのは名前がそのまま「Antigravity」のデスクトップ版で、IDE や CLI では動かない。
`language_server --standalone` を起動できるのがそれだけだからだ。コードが外に出ることは
ない。プロキシも language server も同じマシンで動く。開発手順、FAQ、詳しいセキュリティは
[英語の README](README.md) にある。

[Apache-2.0](LICENSE)。Google のプロジェクトではない。[DISCLAIMER.md](DISCLAIMER.md) 参照。

<div align="center">

# Antigravity Remote

**Antigravity 데스크톱 앱을 폰에서 쓰기.**

[![release](https://img.shields.io/github/v/release/AFSlayer/antigravity-remote?style=flat-square&color=4f7cff)](https://github.com/AFSlayer/antigravity-remote/releases/latest)
[![ci](https://img.shields.io/github/actions/workflow/status/AFSlayer/antigravity-remote/ci.yml?branch=main&style=flat-square)](https://github.com/AFSlayer/antigravity-remote/actions/workflows/ci.yml)
[![license](https://img.shields.io/badge/license-Apache--2.0-blue?style=flat-square)](LICENSE)

<img src="docs/assets/demo.gif" width="300" alt="폰에서 에이전트에게 서버 상태를 물어보는 모습" />

<sub>모바일 사파리에서 돌아가는 Antigravity. 서버에서 셸 명령을 실행하는 중.</sub>

[English](README.md) · [中文](README.zh-CN.md) · [日本語](README.ja.md) · [Português](README.pt-BR.md) · [Español](README.es.md)

</div>

## 뭐하는 건가

Antigravity 데스크톱 앱에는 `language_server`라는 바이너리가 들어 있다. Google과
실제로 통신하는 건 이 녀석이고, `--standalone`으로 실행하면 Antigravity UI 전체를
웹앱으로 서브하기도 한다. 문제는 `127.0.0.1`만 듣기 때문에 데스크톱 앱 외에는
아무도 접근할 수 없다는 것이다.

`agy-remote`는 그 앞에 비밀번호를 걸고 내 네트워크로 넘겨준다. 지나가는 JS 번들의
문자열 몇 개도 고쳐 쓴다. 데스크톱 IDE를 폰 브라우저에서 쓰면 걸리는 데가 좀 있어서다.

폰에서 Antigravity를 쓰게 해주는 프로젝트는 이미 여러 개 있는데, 다들 UI를 만든다.
자체 채팅 패널을 짜거나 CDP로 화면을 미러링한다. 이건 Antigravity 자기 UI를 그대로
서브한다. 그래서 터미널도 되고 파일 트리도 되고 artifacts랑 브라우저 에이전트도
된다. Google이 새 기능을 넣으면 그냥 따라온다. 내가 만든 게 아니니 따라 만들 일도
없다.

대신 압축된 번들에 패치를 거는 건 깨지기 쉽다. Antigravity가 업데이트되면 패치가
안 먹을 수 있다. `agy-remote`는 시작할 때마다 전체 패치를 검사해서 안 맞는 게 있으면
알려준다.

## 설치

```bash
# macOS, Linux
curl -fsSL https://raw.githubusercontent.com/AFSlayer/antigravity-remote/main/scripts/install-desktop.sh | bash
```

```powershell
# Windows
irm https://raw.githubusercontent.com/AFSlayer/antigravity-remote/main/scripts/install-desktop.ps1 | iex
```

설치하고 바로 실행된다. 브라우저에 제어판이 열리고 QR 코드가 뜬다. 그걸 폰으로
찍으면 끝이다. 비밀번호는 안 쳐도 된다. QR에 10분짜리 1회용 토큰이 들어 있다.

<div align="center">
<img src="docs/assets/control-panel.png" width="320" alt="제어판" />
</div>

Antigravity를 미리 켜둘 필요는 없다. 안 켜져 있으면 `agy-remote`가 앱을 실행하고,
설정에서 원격 제어를 켜고, language server를 기다린 다음, 여러 포트 중 어느 게 UI를
서브하는지 알아낸다.

폰에서는 공유 → *홈 화면에 추가*를 누르는 게 좋다. 브라우저 주소창 없이 Antigravity
아이콘으로 전체화면 실행된다. 이 문서의 스크린샷이 그 상태다.

### 보안 경고에 대해

코드 서명을 안 했다. Apple도 Microsoft도 서명에 연 사용료를 받는다. 그래서 브라우저로
압축 파일을 직접 받으면 OS가 격리한다.

- macOS는 확인할 수 없는 개발자라고 한다. 파일 우클릭 → **열기** → 다시 **열기**.
  아니면 `xattr -d com.apple.quarantine agy-remote`.
- Windows는 SmartScreen 경고가 뜬다. **추가 정보** → **실행**.

`curl`이나 `Invoke-WebRequest`로 받은 파일에는 격리 속성이 안 붙는다. 위 설치 명령을
쓰면 이 과정을 아예 안 겪는다.

## 서버에 올리기

리눅스 서버에 올리면 노트북을 닫아도 계속 돌아가는 Antigravity가 된다. 싼 VPS로도
충분하다. 내 건 Oracle 무료 ARM 인스턴스에서 돌고 있다.

```bash
curl -fsSL https://raw.githubusercontent.com/AFSlayer/antigravity-remote/main/scripts/install.sh | bash
```

도메인과 워크스페이스 폴더를 물어본 다음, Google의 `storage.googleapis.com`에서
공식 Antigravity 빌드를 받아 165MB `language_server`만 꺼낸다. 이 저장소가 Google
바이너리를 재배포하지는 않는다. 그리고 systemd 유닛을 쓰고, Caddy로 HTTPS를 자동
설정하고, 비밀번호를 만든다.

한 단계는 자동화가 안 된다. Antigravity는 `localhost` OAuth 콜백으로 로그인하는데,
원격 서버는 그 콜백을 받을 수가 없다. 데스크톱 앱을 이미 쓰는 컴퓨터에서 토큰을
복사해야 한다.

```bash
scp ~/.gemini/jetski-standalone-oauth-token you@your-server:~/.gemini/
```

토큰이 없으면 `agy-remote`가 이 명령을 그대로 찍어준다.

## 어떤 걸 패치하나

패치 12개, 각각 설명이
[`internal/patches/registry.go`](internal/patches/registry.go)에 달려 있다.
`agy-remote doctor`로 어떤 게 적용됐는지 볼 수 있다.

| 문제 | 패치 |
| --- | --- |
| 번들이 `https://127.0.0.1:<port>`로 호출한다. 폰에서는 그게 폰 자신이다 | 브라우저 origin을 쓴다 |
| 문장 쓰는 중에 엔터를 누르면 전송된다 | 터치 기기에서 엔터는 줄바꿈, Cmd/Ctrl+엔터가 전송 |
| iOS 홈 바가 입력창을 가린다 | `safe-area-inset-bottom` 반영, 키보드 올라오면 여백 제거 |
| 모델을 탭하면 medium으로 선택되고 팝업이 닫힌다 | 탭하면 추론 단계 서브메뉴가 열린다 |
| 스탠드얼론에는 전사 기능이 없는데 마이크 버튼이 있다 | 숨긴다 |
| 새 프로젝트가 `/`에서 시작한다 | 지정한 워크스페이스 폴더에서 시작 |
| 앱 아이콘 없음, 300ms 탭 지연, 브라우저 UI | 아이콘, 즉시 반응, 매니페스트로 전체화면 |

패치 없이 원본 그대로 보고 싶으면 `agy-remote --no-mobile-patches`.

<div align="center">
<table><tr>
<td align="center"><img src="docs/assets/patch-models.png" width="190" alt="모델 선택" /></td>
<td align="center"><img src="docs/assets/patch-effort.png" width="190" alt="추론 단계" /></td>
<td align="center"><img src="docs/assets/settings.png" width="190" alt="설정" /></td>
</tr></table>
</div>

## 보안

Antigravity에 접근할 수 있으면 파일을 읽고 명령을 실행할 수 있다. 셸 접근을 주는 것과
같다고 보면 된다.

- 비밀번호는 PBKDF2-SHA256 20만 회로 해시한다. 평문으로 저장하는 곳은 없다.
- 세션 토큰은 랜덤 256비트고, 디스크에는 해시만 쓴다. `sessions.json`을 통째로
  가져가도 쓸 수 없다. `agy-remote sessions revoke`로 전부 로그아웃시킬 수 있다.
- 로그인은 IP당 5분에 5회 실패까지다. 그 뒤로는 최대 30분까지 두 배씩 늘어나는
  락아웃이 걸린다. 분산 시도용 전역 제한도 있다. 락아웃 중에는 맞는 비밀번호도 막힌다.
- 쿠키는 `HttpOnly`, `SameSite=Lax`이고, HTTPS로 들어온 요청이면 `Secure`도 붙는다.
- QR 코드와 종료 버튼이 있는 제어판은 별도 포트에서 루프백만 듣는다. 네트워크로
  나가지 않는다.

리버스 프록시 뒤에 둘 거면 신뢰할 peer를 지정해야 한다. 안 하면 forwarded 헤더를
무시한다.

```bash
agy-remote serve --public-url https://agy.example.com --trusted-proxies 127.0.0.1/32
```

나머지는 [SECURITY.md](SECURITY.md)에 있다. 요약하면, 가능하면 Tailscale 안에 두고,
안 되면 HTTPS를 쓰라는 얘기다.

## 명령어

```
agy-remote                     데스크톱 앱을 내 네트워크에 공유
agy-remote serve               서버에서 헤드리스로 실행
agy-remote doctor              전체 점검하고 문제점 출력
agy-remote config [flags]      옵션을 config.json에 저장
agy-remote passwd [password]   비밀번호 설정
agy-remote sessions [revoke]   로그인된 기기 목록 / 전체 로그아웃
```

자주 쓰는 플래그는 `--port`, `--public-url`, `--workspace-root`,
`--trusted-proxies`, `--session-days`, `--no-mobile-patches`, `--language-server`.
전부 `AGY_*` 환경변수가 있고 `agy-remote help`에 다 나온다.

## 구조

```
    폰                          내 컴퓨터 또는 서버
┌──────────┐              ┌────────────────────────────────┐
│ 브라우저 │   비밀번호   │ agy-remote                     │
│          │◄────────────►│   세션, 레이트리밋             │
│  Anti-   │    :8765     │   main.js / index.html 패치    │
│ gravity  │              │              │ https           │
│   UI     │              │   language_server --standalone │
└──────────┘              └──────────────┼─────────────────┘
                                         ▼
                                  Google CloudCode
```

프롬프트와 코드는 프록시되는 바이트로만 지나가고, 같은 호스트의 language server 외
어디로도 안 간다. Antigravity가 Google과 주고받는 건 그대로다.

## 자주 묻는 것

**IDE나 CLI로도 되나?**
안 된다. 그냥 "Antigravity"라고 되어 있는 데스크톱 앱이어야 한다. 웹 UI를 서브하는
`language_server --standalone`을 실행할 수 있는 건 그것뿐이다. CLI 바이너리도 번들을
품고 있지만 그걸 서브하는 플래그가 없다.

**Antigravity가 업데이트되면 계속 쓸 수 있나?**
프록시는 계속 된다. 개별 패치는 깨질 수 있다. 원격 접속에 꼭 필요한 건
`base-url-origin` 하나고 나머지는 편의 기능이다. 깨지면 이슈를 남겨주면 된다.

**내 코드가 어디로 가나?**
안 간다. 프록시도 language server도 내 컴퓨터에서 돈다.

**둘이 같이 써도 되나?**
기기마다 세션은 따로지만 Antigravity 인스턴스와 Google 계정은 하나를 공유한다. 팀
서버가 아니라 내 기기용이다.

**LAN에서는 HTTP가 괜찮은데 인터넷에서는 왜 안 되나?**
폰 브라우저는 자체 서명 인증서를 거부하고, `192.168.x.x`로는 정식 인증서를 못 받는다.
믿는 망에서는 감수할 만하지만 공개 주소에서는 아니다. `--public-url`과 Caddy나 터널을
쓰면 된다.

## 개발

```bash
go test ./...
go run ./cmd/agy-remote
```

볼 만한 건 [`internal/patches`](internal/patches)다. 패치는 앵커 문자열을 가진
구조체고, `All()`에 하나 추가하면 끝이다. 테스트, `doctor`, 제어판, 캐시 키가 전부 그
목록을 읽는다.

Antigravity 번들은 이 저장소에 없다. 그래서 패치를 두 단계로 검증한다.
`patches_test.go`는 레지스트리로 합성 문서를 만들어 엔진을 테스트하고,
`live_test.go`는 실행 중인 language server에서 실제 번들을 받아 각 앵커가 정확히 1번
매칭되는지 확인한다(안 켜져 있으면 skip). 릴리즈 전에는 데스크톱 앱을 켜고 이걸
돌리면 된다.

```bash
go test ./internal/patches -run Live -v
```

## 라이선스

[Apache-2.0](LICENSE). Google 프로젝트가 아니고 Google과 아무 관계도 없다.
[DISCLAIMER.md](DISCLAIMER.md)를 봐주면 좋겠다.

<div align="center">

# Antigravity Remote

**Use o app desktop do Antigravity pelo celular.**

[![release](https://img.shields.io/github/v/release/AFSlayer/antigravity-remote?style=flat-square&color=4f7cff)](https://github.com/AFSlayer/antigravity-remote/releases/latest)
[![ci](https://img.shields.io/github/actions/workflow/status/AFSlayer/antigravity-remote/ci.yml?branch=main&style=flat-square)](https://github.com/AFSlayer/antigravity-remote/actions/workflows/ci.yml)
[![license](https://img.shields.io/badge/license-Apache--2.0-blue?style=flat-square)](LICENSE)

<img src="docs/assets/demo.gif" width="300" alt="Pedindo o estado do servidor a um agente pelo celular" />

<sub>Antigravity no Safari do celular, rodando um comando de shell num servidor.</sub>

[English](README.md) · [한국어](README.ko.md) · [中文](README.zh-CN.md) · [日本語](README.ja.md) · [Español](README.es.md)

</div>

## O que é isso

O app desktop do Antigravity vem com um binário chamado `language_server`. É ele que
realmente conversa com o Google, e com `--standalone` ele também serve toda a
interface do Antigravity como aplicação web. O detalhe é que ele escuta só em
`127.0.0.1`, então ninguém além do próprio desktop consegue chegar nele.

O `agy-remote` coloca uma senha na frente desse servidor e o encaminha para a sua
rede. Ele também reescreve algumas strings no bundle JS no caminho, porque uma IDE
de desktop no navegador do celular tem alguns pontos incômodos.

Já existe uma dúzia de projetos para usar o Antigravity no celular, e todos
constroem uma interface: um painel de chat próprio, ou espelhamento de tela via CDP.
Este serve a interface do próprio Antigravity. Então o terminal funciona, a árvore de
arquivos funciona, artifacts e o browser agent funcionam, e recursos novos do Google
aparecem por conta própria. Não escrevi nada disso e também não preciso acompanhar.

O custo é que aplicar patch num bundle minificado é frágil. Uma atualização do
Antigravity pode quebrar um patch. O `agy-remote` verifica todos eles a cada
inicialização e diz quais pararam de casar.

## Instalação

```bash
# macOS, Linux
curl -fsSL https://raw.githubusercontent.com/AFSlayer/antigravity-remote/main/scripts/install-desktop.sh | bash
```

```powershell
# Windows
irm https://raw.githubusercontent.com/AFSlayer/antigravity-remote/main/scripts/install-desktop.ps1 | iex
```

Isso instala o binário e já inicia. Um painel de controle abre no navegador com um QR
code. Escaneie e você entra. Não precisa digitar senha, porque o código carrega um
token de uso único válido por dez minutos.

<div align="center">
<img src="docs/assets/control-panel.png" width="320" alt="Painel de controle" />
</div>

O Antigravity não precisa estar aberto antes. Se não estiver rodando, o `agy-remote`
inicia o app, liga o controle remoto nas configurações, espera o language server e
descobre qual das portas dele serve a interface.

No celular, use Compartilhar e *Adicionar à Tela de Início*. Você fica com um app em
tela cheia, com o ícone do Antigravity e sem barra do navegador. As capturas desta
página são assim.

### Sobre os avisos de segurança

Os binários não são assinados, porque Apple e Microsoft cobram anualmente por isso.
Se você baixar o arquivo pelo navegador, o sistema coloca em quarentena:

- O macOS diz que não é possível verificar o desenvolvedor. Clique com o botão direito,
  **Abrir**, e **Abrir** de novo. Ou rode `xattr -d com.apple.quarantine agy-remote`.
- O Windows mostra o SmartScreen. **Mais informações** → **Executar assim mesmo**.

`curl` e `Invoke-WebRequest` não marcam o arquivo em quarentena, então os comandos de
instalação acima evitam tudo isso.

## Colocando num servidor

Numa máquina Linux, isso vira um Antigravity que continua trabalhando com o notebook
fechado. Uma VPS barata já resolve. A minha é uma instância ARM gratuita da Oracle.

```bash
curl -fsSL https://raw.githubusercontent.com/AFSlayer/antigravity-remote/main/scripts/install.sh | bash
```

Ele pergunta o domínio e a pasta de trabalho, baixa a build oficial do Antigravity do
`storage.googleapis.com` do Google e extrai apenas o `language_server` de 165 MB.
Nenhum binário do Google é redistribuído aqui. Depois escreve uma unit do systemd,
configura o Caddy para HTTPS automático e gera uma senha.

Um passo não dá para automatizar. O login do Antigravity usa um callback OAuth em
`localhost`, que um servidor remoto não consegue receber. Copie o token de uma máquina
onde você já usa o app desktop:

```bash
scp ~/.gemini/jetski-standalone-oauth-token voce@seu-servidor:~/.gemini/
```

O `agy-remote` imprime esse comando quando o token está faltando.

## O que é modificado

Doze patches, cada um com descrição em
[`internal/patches/registry.go`](internal/patches/registry.go). O
`agy-remote doctor` informa quais foram aplicados.

| Problema | Patch |
| --- | --- |
| O bundle chama `https://127.0.0.1:<port>`, que no celular é o próprio celular | Usar a origin do navegador |
| Enter envia a mensagem no meio da frase | Em telas de toque Enter é nova linha, Cmd/Ctrl+Enter envia |
| A barra inferior do iOS cobre o campo de texto | Respeitar `safe-area-inset-bottom` e remover o espaço com o teclado aberto |
| Tocar num modelo escolhe medium e fecha o menu | Tocar abre o submenu de esforço |
| Um botão de microfone que não funciona, já que standalone não transcreve | Esconder |
| Projetos novos começam em `/` | Começar na pasta de trabalho configurada |
| Sem ícone, atraso de 300 ms no toque, moldura do navegador | Ícone, toque imediato e manifest para tela cheia |

Para a interface sem alterações, use `agy-remote --no-mobile-patches`.

<div align="center">
<table><tr>
<td align="center"><img src="docs/assets/patch-models.png" width="190" alt="Seleção de modelo" /></td>
<td align="center"><img src="docs/assets/patch-effort.png" width="190" alt="Nível de esforço" /></td>
<td align="center"><img src="docs/assets/settings.png" width="190" alt="Configurações" /></td>
</tr></table>
</div>

## Segurança

Quem alcança o Antigravity pode ler seus arquivos e rodar comandos, então trate o
acesso como equivalente a um shell.

- Senhas usam PBKDF2-SHA256 com 200 mil iterações e nunca ficam em texto puro.
- Tokens de sessão têm 256 bits aleatórios e só os hashes vão para o disco, então um
  `sessions.json` copiado não serve para nada. `agy-remote sessions revoke` desconecta
  tudo.
- O login permite cinco falhas por IP a cada cinco minutos, depois entra um bloqueio
  que dobra até 30 minutos. Há também um limite global para tentativas distribuídas. O
  bloqueio vale até para a senha correta.
- Cookies são `HttpOnly`, `SameSite=Lax`, e `Secure` quando a requisição chegou por
  HTTPS.
- O painel de controle, com o QR code e o botão de desligar, escuta apenas em loopback
  numa porta separada. Nunca é exposto à rede.

Atrás de um proxy reverso, informe em quais peers confiar, senão os cabeçalhos
forwarded são ignorados:

```bash
agy-remote serve --public-url https://agy.example.com --trusted-proxies 127.0.0.1/32
```

O resto está em [SECURITY.md](SECURITY.md). Resumindo: coloque a máquina no Tailscale
se puder, e use HTTPS se não puder.

## Comandos

```
agy-remote                     compartilha o app desktop na sua rede
agy-remote serve               roda headless num servidor
agy-remote doctor              verifica tudo e diz o que está errado
agy-remote config [flags]      grava opções no config.json
agy-remote passwd [password]   define a senha
agy-remote sessions [revoke]   lista ou desconecta dispositivos
```

Flags úteis: `--port`, `--public-url`, `--workspace-root`, `--trusted-proxies`,
`--session-days`, `--no-mobile-patches`, `--language-server`. Cada uma tem uma
variável de ambiente `AGY_*`, e `agy-remote help` lista todas.

## Como funciona

```
  celular                     sua máquina ou servidor
┌──────────┐              ┌────────────────────────────────┐
│navegador │    senha     │ agy-remote                     │
│          │◄────────────►│   sessões, rate limiting       │
│  Anti-   │    :8765     │   patch main.js / index.html   │
│ gravity  │              │              │ https           │
│   UI     │              │   language_server --standalone │
└──────────┘              └──────────────┼─────────────────┘
                                         ▼
                                  Google CloudCode
```

Prompts e código passam como bytes encaminhados e não vão a lugar nenhum além do
language server no mesmo host. O tráfego do Antigravity com o Google não muda.

## Perguntas frequentes

**Dá para usar a IDE ou a CLI?**
Não, precisa ser o app desktop, aquele chamado só de "Antigravity". Só ele roda
`language_server --standalone`, o modo que serve a interface web. O binário da CLI tem
o bundle compilado dentro, mas nenhuma flag para servi-lo.

**Vai continuar funcionando depois de atualizações?**
O proxy sim. Patches individuais podem parar. O `base-url-origin` é o único de que o
acesso remoto realmente precisa, os outros são conveniência. Abra uma issue quando um
quebrar.

**Meu código vai para algum lugar?**
Não. O proxy e o language server rodam na sua máquina.

**Por que HTTP serve na rede local e não na internet?**
Navegadores de celular rejeitam certificados autoassinados, e não existe certificado
válido para `192.168.x.x`. Numa rede de confiança a troca é aceitável, num endereço
público não. Use `--public-url` atrás do Caddy ou de um túnel.

## Desenvolvimento

```bash
go test ./...
go run ./cmd/agy-remote
```

A parte que vale ler é [`internal/patches`](internal/patches). Um patch é uma struct
com uma string de âncora, e adicionar um em `All()` basta: testes, `doctor`, painel de
controle e chave de cache todos leem dessa lista.

O bundle do Antigravity não está neste repositório, então os patches são testados em
dois níveis. `patches_test.go` monta um documento sintético a partir do registro para
testar o motor. `live_test.go` busca o bundle real de um language server em execução e
garante que cada âncora casa exatamente uma vez, pulando quando nada está rodando.
Antes de uma release, abra o app desktop e rode:

```bash
go test ./internal/patches -run Live -v
```

## Licença

[Apache-2.0](LICENSE). Não é um projeto do Google nem tem ligação com o Google. Veja
[DISCLAIMER.md](DISCLAIMER.md).

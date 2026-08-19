<div align="center">

# Antigravity Remote

**Use o app desktop do Antigravity pelo celular.**

[![release](https://img.shields.io/github/v/release/AFSlayer/antigravity-remote?style=flat-square&color=4f7cff)](https://github.com/AFSlayer/antigravity-remote/releases/latest)
[![license](https://img.shields.io/badge/license-Apache--2.0-blue?style=flat-square)](LICENSE)

<img src="docs/assets/demo.gif" width="300" alt="Pedindo o estado do servidor a um agente pelo celular" />

[English](README.md) · [한국어](README.ko.md) · [中文](README.zh-CN.md) · [日本語](README.ja.md) · [Español](README.es.md)

</div>

## O que é

O app desktop do Antigravity vem com um binário chamado `language_server`. É ele que
fala com o Google, e com `--standalone` ele também serve toda a interface do
Antigravity como aplicação web. Só escuta em `127.0.0.1`.

O `agy-remote` coloca uma senha na frente dele e encaminha para a sua rede. De
passagem, reescreve algumas strings no bundle JS, porque uma IDE de desktop no
navegador do celular tem pontos incômodos.

Os outros projetos que fazem isso constroem a própria interface, ou espelham a tela
via CDP. Este serve a interface do próprio Antigravity, então terminal, árvore de
arquivos, artifacts e browser agent funcionam, e recursos novos do Google aparecem
sozinhos.

O custo é que aplicar patch em bundle minificado é frágil. Uma atualização pode
quebrar um patch. O `agy-remote` verifica todos na inicialização e diz quais pararam
de casar.

## Instalação

```bash
# macOS, Linux
curl -fsSL https://raw.githubusercontent.com/AFSlayer/antigravity-remote/main/scripts/install-desktop.sh | bash
```

```powershell
# Windows
irm https://raw.githubusercontent.com/AFSlayer/antigravity-remote/main/scripts/install-desktop.ps1 | iex
```

Ele instala e já inicia. Abre um painel de controle com um QR code. Escaneie e você
entra, sem digitar senha. O código leva um token de uso único válido por dez minutos.

<div align="center">
<img src="docs/assets/control-panel.png" width="320" alt="Painel de controle" />
</div>

O Antigravity não precisa estar aberto. Se não estiver, o `agy-remote` inicia o app,
liga o controle remoto, espera o language server e descobre qual porta serve a
interface.

No celular use Compartilhar → *Adicionar à Tela de Início*. Fica com o ícone do
Antigravity na tela de início.

Os binários não são assinados, então baixar o arquivo pelo navegador deixa ele em
quarentena. No macOS: clique com o botão direito, **Abrir**, e **Abrir** de novo. No
Windows: **Mais informações** → **Executar assim mesmo**. Os comandos acima usam
`curl`, que não marca a quarentena.

## Num servidor

Numa máquina Linux o Antigravity continua trabalhando com o notebook fechado. Uma VPS
barata ou uma instância ARM de nível gratuito resolve.

```bash
curl -fsSL https://raw.githubusercontent.com/AFSlayer/antigravity-remote/main/scripts/install.sh | bash
```

Pergunta o domínio e a pasta de trabalho, baixa a build oficial do
`storage.googleapis.com` do Google e extrai só o `language_server` de 165 MB. Nenhum
binário do Google é redistribuído aqui. Depois escreve a unit do systemd, configura o
Caddy para HTTPS e gera uma senha.

Como é a mesma interface web, o navegador do desktop também funciona, com a mesma
aparência e o mesmo comportamento do app. Conversas, workspaces e agentes em execução
ficam no servidor, então dá para começar algo no celular no caminho e continuar no
desktop ao chegar. Não há sincronização porque existe só uma instância.

Um passo não é automático. O login do Antigravity usa um callback OAuth em
`localhost`, que um servidor remoto não recebe. Copie o token de uma máquina onde
você já usa o app desktop:

```bash
scp ~/.gemini/jetski-standalone-oauth-token voce@seu-servidor:~/.gemini/
```

O `agy-remote` mostra esse comando quando o token está faltando.

## O que é modificado

Treze patches, descritos em
[`internal/patches/registry.go`](internal/patches/registry.go). O `agy-remote doctor`
diz quais foram aplicados.

| Problema | Patch |
| --- | --- |
| O bundle chama `https://127.0.0.1:<port>`, que no celular é o próprio celular | Usar a origin do navegador |
| Enter envia no meio da frase | No toque, Enter é nova linha e Cmd/Ctrl+Enter envia |
| Tocar num modelo escolhe medium e fecha o menu | Tocar abre o submenu de esforço |
| Banner "Enable Notifications" na primeira resposta | Não mostrar em telas de toque |
| Botão de microfone que não funciona, já que standalone não transcreve | Esconder |
| Projetos novos começam em `/` | Começar na pasta de trabalho configurada |
| Sem ícone, atraso de 300 ms | O ícone do Antigravity, toque imediato |

Para a interface intacta, `agy-remote --no-mobile-patches`.

## Segurança

Quem entra pode ler seus arquivos e rodar comandos, então isso é mais parecido com dar
acesso ao shell do que com compartilhar um documento.

- Senhas passam por PBKDF2-SHA256 com 200 mil iterações. Nada fica em texto puro.
- Tokens de sessão têm 256 bits aleatórios e só os hashes vão para o disco, então um
  `sessions.json` copiado não serve.
- Cinco falhas por IP a cada cinco minutos, depois um bloqueio que dobra até 30
  minutos. Vale também para a senha correta.
- O painel de controle, com o QR code e o botão de desligar, escuta só em loopback numa
  porta separada. Nunca vai para a rede.

Atrás de um proxy reverso, diga em quais peers confiar ou os cabeçalhos forwarded são
ignorados:

```bash
agy-remote serve --public-url https://agy.example.com --trusted-proxies 127.0.0.1/32
```

O resto está em [SECURITY.md](SECURITY.md). Use Tailscale se puder, HTTPS se não.

## Comandos

```
agy-remote                     compartilha o app desktop na sua rede
agy-remote serve               roda headless num servidor
agy-remote doctor              verifica tudo e diz o que está errado
agy-remote config [flags]      grava opções no config.json
agy-remote passwd [password]   define a senha
agy-remote sessions [revoke]   lista ou desconecta dispositivos
```

As flags estão em `agy-remote help`. Cada uma tem sua variável `AGY_*`.

## Resto

Precisa ser o app desktop, o chamado só de "Antigravity". A IDE e a CLI não servem,
porque só ele roda `language_server --standalone`. Seu código não sai da máquina: o
proxy e o language server rodam no mesmo host. Desenvolvimento, FAQ e a documentação
completa de segurança estão no [README em inglês](README.md).

[Apache-2.0](LICENSE). Não é um projeto do Google. Veja [DISCLAIMER.md](DISCLAIMER.md).

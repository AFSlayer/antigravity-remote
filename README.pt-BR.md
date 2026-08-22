<div align="center">

# Antigravity Server

Servidor auto-hospedado e ponte de interface web para o Google Antigravity.  
Execute o Antigravity 24/7 em uma instância Linux headless ou no seu desktop local, acessível via navegador web.

[![release](https://img.shields.io/github/v/release/AFSlayer/antigravity-server?style=flat-square&color=4f7cff)](https://github.com/AFSlayer/antigravity-server/releases/latest)
[![ci](https://img.shields.io/github/actions/workflow/status/AFSlayer/antigravity-server/ci.yml?branch=main&style=flat-square)](https://github.com/AFSlayer/antigravity-server/actions/workflows/ci.yml)
[![license](https://img.shields.io/badge/license-Apache--2.0-blue?style=flat-square)](LICENSE)

<img src="docs/assets/demo.gif" width="320" alt="Antigravity Server no navegador mobile" />

[English](README.md) · [한국어](README.ko.md) · [中文](README.zh-CN.md) · [日本語](README.ja.md) · [Español](README.es.md)

</div>

---

## Por que Antigravity Server? (vs Ponte Remota Oficial)

O Google oferece uma ponte remota oficial (`antigravity.google.com/r/...`), mas o tráfego passa por um relay na nuvem e exige que o aplicativo desktop permaneça aberto.

O `agy-server` roda em modo headless em um VPS Linux ou servidor local, oferecendo conexão de rede direta e patches de interface para dispositivos móveis:

| Recurso | Ponte Remota Oficial Google | Antigravity Server (`agy-server`) |
| :--- | :--- | :--- |
| **Ambiente de Hospedagem** | Exige app desktop com interface gráfica aberto | **VPS Linux Headless / VM na nuvem** (serviço systemd, auto-atualizador) |
| **Conexão e Latência** | Relay em nuvem através dos servidores Google | **Conexão Direta** (LAN, VPN ou proxy reverso com HTTPS) |
| **Gestão de Projetos Mobile** | Sem botão `(+)`; troca incômoda pelo input | **Botão `(+)` restaurado** no cabeçalho dos projetos |
| **Controle de Conversas** | Não permite excluir, fixar ou arquivar no celular | **Controle por toque**: Excluir, Renomear, Fixar e Arquivar |
| **Ações de Mensagem** | Ocultas sob o cursor (hover) | **Botões Desfazer (`↶`) e Copiar (`📋`) visíveis no mobile** |
| **Ajuste de Teclado iOS/PWA** | Espaço vazio no Safe Area; tela pula ao focar | **Ajuste 0px ao teclado**: colapso dinâmico do Safe Area |
| **Upload de Arquivos** | Limite de 1MB por RPC | **Upload por streaming fragmentado**: envie logs pesados e datasets |
| **Autenticação e Privacidade** | Vinculado à conta Google e relay na nuvem | Protegido por senha (PBKDF2), gestão de sessões e rate-limiting |

---

## Início Rápido

### Opção 1: Servidor Linux / VPS na Nuvem (Recomendado)

Execute o Antigravity em uma instância Linux headless (Oracle Cloud Free Tier, AWS, DigitalOcean ou servidor caseiro):

```bash
curl -fsSL https://raw.githubusercontent.com/AFSlayer/antigravity-server/main/scripts/install.sh | bash
```

O instalador:
1. Solicita seu domínio (ex: `agy.example.com`) e o diretório do workspace.
2. Baixa o `language_server` diretamente do bucket oficial do Google (`storage.googleapis.com`).
3. Configura o Caddy para HTTPS automático, cria o serviço systemd e define a senha de acesso.

#### Autenticação com o Google
Ao acessar o servidor pela primeira vez:
- **Login Direto pela Web**: Acesse a interface no navegador, vá em **Settings** e faça login com sua conta Google.
- **Ou Copiar Token Existente (Opcional)**: Caso já tenha feito login no desktop local:
  ```bash
  scp ~/.gemini/jetski-standalone-oauth-token user@seu-servidor:~/.gemini/
  ```

---

### Opção 2: Modo Desktop Companion (macOS, Windows, Linux Desktop)

Para compartilhar o Antigravity local na mesma rede Wi-Fi:

```bash
# macOS & Linux
curl -fsSL https://raw.githubusercontent.com/AFSlayer/antigravity-server/main/scripts/install-desktop.sh | bash
```

```powershell
# Windows (PowerShell)
irm https://raw.githubusercontent.com/AFSlayer/antigravity-server/main/scripts/install-desktop.ps1 | iex
```

O `agy-server` abre um painel de controle com QR code para conexão rápida pelo celular sem senha.

<div align="center">
<img src="docs/assets/control-panel.png" width="320" alt="Control Panel" />
</div>

---

## Configuração PWA Mobile (Adicionar à Tela de Início)

O Antigravity Server suporta o padrão Progressive Web App (PWA). Ao adicioná-lo à tela de início do smartphone, ele roda em **tela cheia sem barra de endereços**:

- **iOS (Safari)**: Toque em **Compartilhar (`⎋`)** → Selecione **Adicionar à Tela de Início**.
- **Android (Chrome)**: Toque no **Menu (`⋮`)** → Selecione **Instalar aplicativo** ou **Adicionar à tela inicial**.

> [!TIP]
> Executar como PWA garante que o patch de **ajuste 0px do teclado virtual** funcione com máxima suavidade.

---

## Principais Recursos

### ⚡ Patches de Interface para Mobile
- **Controles por Toque**: Botões Desfazer (`↶`) e Copiar (`📋`) permanentemente visíveis nos balões de mensagem.
- **Gerenciamento de Conversas**: Exclua conversas pela barra superior e fixe ou arquive pelo menu lateral.
- **Ajuste Preciso do Teclado**: Reduz o Safe Area para 0px assim que o teclado virtual aparece.

---

### 📁 Upload de Arquivos Grandes por Streaming
Transfira arquivos pesados, logs e datasets diretamente para o workspace sem o limite de 1MB:

<div align="center">
<img src="docs/assets/upload.gif" width="560" alt="Demonstração do uploader por streaming" />
</div>

---

### 🖥️ Interface Web para Desktop e Tablet
Interface perfeitamente responsiva para navegadores de computadores e tablets:

<div align="center">
<img src="docs/assets/desktop.png" width="700" alt="Antigravity Web UI no navegador desktop" />
</div>

---

### 🔄 Atualizações Automáticas sem Quedas
Em servidores Linux headless, o `agy-server` inclui serviço de atualização automática:
- Verifica diariamente novas versões oficiais do `language_server`.
- Substitui o binário de forma atômica e segura.
- Verificação manual: execute `agy-server update`.

---

## Configuração de Proxy Reverso (Caddy / Nginx)

Para habilitar streaming em tempo real (SSE), WebSockets e uploads grandes, desative o buffer do proxy:

### Caddy
```caddyfile
agy.example.com {
    encode zstd gzip

    reverse_proxy 127.0.0.1:8765 {
        flush_interval -1
    }
}
```

### Nginx
```nginx
server {
    listen 443 ssl http2;
    server_name agy.example.com;

    client_max_body_size 0;

    location / {
        proxy_pass http://127.0.0.1:8765;
        proxy_http_version 1.1;

        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";

        proxy_buffering off;
        proxy_cache off;
        proxy_read_timeout 86400s;

        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

> [!IMPORTANT]
> Configure `--trusted-proxies 127.0.0.1/32` (ou a variável `AGY_TRUSTED_PROXIES=127.0.0.1/32`) para que a proteção contra força bruta identifique o IP real do usuário.

---

## Comandos CLI

```
agy-server                      Inicia no modo desktop companion (rede local)
agy-server serve                Executa como daemon em servidor headless
agy-server update               Verifica e atualiza o language_server oficial
agy-server doctor               Diagnostica o estado do sistema e patches
agy-server passwd [password]    Define ou altera a senha de acesso web
agy-server sessions [revoke]    Lista sessões ativas ou desconecta todos os aparelhos
agy-server config [flags]       Gerencia configurações em config.json
```

---

## Licença

[Apache-2.0](LICENSE). Não afiliado nem endossado pelo Google. Consulte [DISCLAIMER.md](DISCLAIMER.md).

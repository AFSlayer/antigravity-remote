<div align="center">

# Antigravity Remote

**Usa la app de escritorio de Antigravity desde el móvil.**

[![release](https://img.shields.io/github/v/release/AFSlayer/antigravity-remote?style=flat-square&color=4f7cff)](https://github.com/AFSlayer/antigravity-remote/releases/latest)
[![license](https://img.shields.io/badge/license-Apache--2.0-blue?style=flat-square)](LICENSE)

<img src="docs/assets/demo.gif" width="300" alt="Pidiendo el estado del servidor a un agente desde el móvil" />

[English](README.md) · [한국어](README.ko.md) · [中文](README.zh-CN.md) · [日本語](README.ja.md) · [Português](README.pt-BR.md)

</div>

## Qué es

La app de escritorio de Antigravity incluye un binario llamado `language_server`. Es
el que habla con Google, y con `--standalone` además sirve toda la interfaz de
Antigravity como aplicación web. Solo escucha en `127.0.0.1`.

`agy-remote` le pone una contraseña delante y lo reenvía a tu red. De paso reescribe
algunas cadenas del bundle JS, porque un IDE de escritorio en el navegador del móvil
tiene asperezas.

Los demás proyectos que hacen esto construyen su propia interfaz, o espejan la
pantalla por CDP. Este sirve la de Antigravity, así que la terminal, el árbol de
archivos, los artifacts y el browser agent funcionan, y lo nuevo de Google aparece
solo.

El precio es que parchear un bundle minificado es frágil. Una actualización puede
romper un parche. `agy-remote` los revisa al arrancar y dice cuáles dejaron de
coincidir.

## Instalación

```bash
# macOS, Linux
curl -fsSL https://raw.githubusercontent.com/AFSlayer/antigravity-remote/main/scripts/install-desktop.sh | bash
```

```powershell
# Windows
irm https://raw.githubusercontent.com/AFSlayer/antigravity-remote/main/scripts/install-desktop.ps1 | iex
```

Se instala y arranca. Se abre un panel de control con un código QR. Lo escaneas y
entras, sin escribir contraseña. El código lleva un token de un solo uso válido diez
minutos.

<div align="center">
<img src="docs/assets/control-panel.png" width="320" alt="Panel de control" />
</div>

Antigravity no tiene que estar abierto. Si no lo está, `agy-remote` lo arranca, activa
el control remoto, espera al language server y averigua qué puerto sirve la interfaz.

En el móvil usa Compartir → *Añadir a pantalla de inicio*. Queda con el icono de
Antigravity en la pantalla de inicio.

Los binarios no están firmados, así que descargar el archivo desde el navegador lo
deja en cuarentena. En macOS: clic derecho, **Abrir**, y **Abrir** otra vez. En
Windows: **Más información** → **Ejecutar de todas formas**. Los comandos de arriba
usan `curl`, que no marca la cuarentena.

## En un servidor

En una máquina Linux, Antigravity sigue trabajando con el portátil cerrado. Basta un
VPS barato o una instancia ARM de nivel gratuito.

```bash
curl -fsSL https://raw.githubusercontent.com/AFSlayer/antigravity-remote/main/scripts/install.sh | bash
```

Pregunta el dominio y la carpeta de trabajo, descarga la build oficial desde el
`storage.googleapis.com` de Google y extrae solo el `language_server` de 165 MB. Aquí
no se redistribuye ningún binario de Google. Luego escribe la unit de systemd,
configura Caddy para HTTPS y genera una contraseña.

Como es la misma interfaz web, el navegador del escritorio también sirve, con el mismo
aspecto y comportamiento que la app. Las conversaciones, los workspaces y los agentes
en marcha están en el servidor, así que puedes empezar algo en el móvil de camino y
seguir en el escritorio al llegar. No hay sincronización porque solo hay una instancia.

Un paso no es automático. El inicio de sesión de Antigravity usa un callback OAuth en
`localhost`, que un servidor remoto no puede recibir. Copia el token desde una máquina
donde ya uses la app de escritorio:

```bash
scp ~/.gemini/jetski-standalone-oauth-token tu@tu-servidor:~/.gemini/
```

`agy-remote` muestra ese comando cuando falta el token.

## Qué parches se aplican
 
Veinticinco parches, cada uno descrito en
[`internal/patches/registry.go`](internal/patches/registry.go). `agy-remote doctor`
indica cuáles se aplicaron.

| Problema | Parche |
| --- | --- |
| El paquete llama a `https://127.0.0.1:<puerto>`, que desde un teléfono es el teléfono | Usar el origen del navegador |
| Enter envía a mitad de frase | En táctil, Enter es salto de línea y Cmd/Ctrl+Enter envía |
| Tocar un modelo selecciona medium y cierra el menú | El toque abre el submenú de esfuerzo |
| Banner "Enable Notifications" en la primera respuesta | Omitirlo en dispositivos táctiles |
| Un botón de micrófono que no puede funcionar, ya que standalone no tiene transcripción | Ocultarlo |
| Icono de perfil de usuario sin función en la barra superior móvil | Ocultarlo |
| Tocar `+` en móvil no abre un compositor dedicado | Abrir vista de nueva conversación directamente e integrar botón volver en la barra superior móvil |
| Nuevos proyectos empiezan en `/` | Empezar en la carpeta de espacio de trabajo configurada |
| Archivos grandes (.har, logs) congelan el navegador o fallan por límite de 20MB | Transmisión asíncrona directa al espacio de trabajo con interfaz de progreso nativa |
| Archivos no estándar (.har, .jsonl) rechazados al adjuntar | Permitir todos los tipos de archivo y analizar datos de texto |
| Sin icono, retraso de toque de 300 ms | El icono de Antigravity, toques instantáneos |
| Solapamiento de la barra de inicio de iOS y teclado virtual que desborda la vista | Márgenes de área segura y sincronización de altura de vista |
| La versión standalone no puede iniciar sesión desde un navegador remoto | Conectar el botón de inicio de sesión de Ajustes al flujo de autenticación web |

Para dejar la interfaz intacta, `agy-remote --no-mobile-patches`.

## Seguridad

Quien entra puede leer tus archivos y ejecutar comandos, así que esto se parece más a
dar acceso al shell que a compartir un documento.

- Las contraseñas pasan por PBKDF2-SHA256 con 200 mil iteraciones. Nada en claro.
- Los tokens de sesión son 256 bits aleatorios y al disco solo van sus hashes, así que
  un `sessions.json` copiado no sirve.
- Cinco fallos por IP cada cinco minutos, luego un bloqueo que se duplica hasta 30
  minutos. Aplica también a la contraseña correcta.
- El panel de control, con el QR y el botón de apagado, escucha solo en loopback y en
  un puerto aparte. Nunca sale a la red.

Detrás de un proxy inverso, indica en qué peers confías o se ignoran las cabeceras
forwarded:

```bash
agy-remote serve --public-url https://agy.example.com --trusted-proxies 127.0.0.1/32
```

El resto está en [SECURITY.md](SECURITY.md). Tailscale si puedes, HTTPS si no.

## Comandos

```
agy-remote                     comparte la app de escritorio en tu red
agy-remote serve               ejecuta sin interfaz en un servidor
agy-remote update [flags]      actualiza Antigravity language_server a la última versión oficial
agy-remote doctor              revisa todo y dice qué está mal
agy-remote config [flags]      escribe opciones en config.json
agy-remote passwd [password]   establece la contraseña
agy-remote sessions [revoke]   lista o desconecta dispositivos
```

Las flags están en `agy-remote help`. Cada una tiene su variable `AGY_*`.

## Lo demás

Hace falta la app de escritorio, la que se llama solo "Antigravity". El IDE y la CLI no
sirven, porque únicamente ella ejecuta `language_server --standalone`. Tu código no sale
de la máquina: el proxy y el language server corren en el mismo host. Desarrollo,
preguntas frecuentes y la documentación completa de seguridad están en el
[README en inglés](README.md).

[Apache-2.0](LICENSE). No es un proyecto de Google. Lee [DISCLAIMER.md](DISCLAIMER.md).

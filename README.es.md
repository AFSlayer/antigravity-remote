<div align="center">

# Antigravity Remote

**Usa la app de escritorio de Antigravity desde el móvil.**

[![release](https://img.shields.io/github/v/release/AFSlayer/antigravity-remote?style=flat-square&color=4f7cff)](https://github.com/AFSlayer/antigravity-remote/releases/latest)
[![ci](https://img.shields.io/github/actions/workflow/status/AFSlayer/antigravity-remote/ci.yml?branch=main&style=flat-square)](https://github.com/AFSlayer/antigravity-remote/actions/workflows/ci.yml)
[![license](https://img.shields.io/badge/license-Apache--2.0-blue?style=flat-square)](LICENSE)

<img src="docs/assets/demo.gif" width="300" alt="Pidiendo el estado del servidor a un agente desde el móvil" />

<sub>Antigravity en Safari móvil, ejecutando un comando de shell en un servidor.</sub>

[English](README.md) · [한국어](README.ko.md) · [中文](README.zh-CN.md) · [日本語](README.ja.md) · [Português](README.pt-BR.md)

</div>

## Qué es esto

La app de escritorio de Antigravity incluye un binario llamado `language_server`. Es
el que realmente habla con Google, y con `--standalone` además sirve toda la interfaz
de Antigravity como aplicación web. El problema es que solo escucha en `127.0.0.1`,
así que nadie más que la propia app de escritorio puede llegar a él.

`agy-remote` pone una contraseña delante de ese servidor y lo reenvía a tu red.
También reescribe algunas cadenas del bundle JS al pasar, porque un IDE de escritorio
en el navegador del móvil tiene algunas asperezas.

Ya hay una docena de proyectos para usar Antigravity desde el móvil, y todos
construyen una interfaz: su propio panel de chat, o un espejo de pantalla por CDP.
Este sirve la interfaz de Antigravity misma. Así funciona la terminal, funciona el
árbol de archivos, funcionan los artifacts y el browser agent, y las novedades de
Google aparecen solas. No escribí nada de eso y tampoco tengo que ir detrás.

El precio es que parchear un bundle minificado es frágil. Una actualización de
Antigravity puede romper un parche. `agy-remote` los comprueba todos en cada arranque
y avisa cuáles dejaron de coincidir.

## Instalación

```bash
# macOS, Linux
curl -fsSL https://raw.githubusercontent.com/AFSlayer/antigravity-remote/main/scripts/install-desktop.sh | bash
```

```powershell
# Windows
irm https://raw.githubusercontent.com/AFSlayer/antigravity-remote/main/scripts/install-desktop.ps1 | iex
```

Eso instala el binario y lo arranca. Se abre un panel de control en el navegador con
un código QR. Lo escaneas y ya estás dentro. No hay contraseña que teclear, porque el
código lleva un token de un solo uso válido diez minutos.

<div align="center">
<img src="docs/assets/control-panel.png" width="320" alt="Panel de control" />
</div>

Antigravity no tiene que estar abierto antes. Si no está corriendo, `agy-remote`
arranca la app, activa el control remoto en los ajustes, espera al language server y
averigua cuál de sus puertos sirve la interfaz.

En el móvil, usa Compartir y *Añadir a pantalla de inicio*. Obtienes una app a
pantalla completa, con el icono de Antigravity y sin barra del navegador. Las capturas
de esta página son así.

### Sobre los avisos de seguridad

Los binarios no están firmados, porque Apple y Microsoft cobran una cuota anual por
eso. Si descargas el archivo desde el navegador, el sistema lo pone en cuarentena:

- macOS dice que no puede verificar al desarrollador. Clic derecho, **Abrir**, y
  **Abrir** otra vez. O ejecuta `xattr -d com.apple.quarantine agy-remote`.
- Windows muestra SmartScreen. **Más información** → **Ejecutar de todas formas**.

`curl` e `Invoke-WebRequest` no marcan el archivo en cuarentena, así que los comandos
de instalación de arriba evitan todo esto.

## Ponerlo en un servidor

En una máquina Linux esto se convierte en un Antigravity que sigue trabajando con el
portátil cerrado. Un VPS barato sobra. El mío es una instancia ARM gratuita de Oracle.

```bash
curl -fsSL https://raw.githubusercontent.com/AFSlayer/antigravity-remote/main/scripts/install.sh | bash
```

Pregunta el dominio y la carpeta de trabajo, descarga la build oficial de Antigravity
desde el `storage.googleapis.com` de Google y extrae solo el `language_server` de
165 MB. Aquí no se redistribuye ningún binario de Google. Después escribe una unit de
systemd, configura Caddy para HTTPS automático y genera una contraseña.

Hay un paso que no se puede automatizar. Antigravity inicia sesión con un callback
OAuth en `localhost`, que un servidor remoto no puede recibir. Copia el token desde
una máquina donde ya uses la app de escritorio:

```bash
scp ~/.gemini/jetski-standalone-oauth-token tu@tu-servidor:~/.gemini/
```

`agy-remote` imprime ese comando cuando falta el token.

## Qué se parchea

Doce parches, cada uno con su descripción en
[`internal/patches/registry.go`](internal/patches/registry.go). `agy-remote doctor`
indica cuáles se aplicaron.

| Problema | Parche |
| --- | --- |
| El bundle llama a `https://127.0.0.1:<port>`, que en el móvil es el móvil mismo | Usar la origin del navegador |
| Enter envía el mensaje a media frase | En pantallas táctiles Enter es salto de línea, Cmd/Ctrl+Enter envía |
| La barra inferior de iOS tapa el campo de texto | Respetar `safe-area-inset-bottom` y quitar el hueco con el teclado abierto |
| Tocar un modelo elige medium y cierra el menú | Tocar abre el submenú de esfuerzo |
| Un botón de micrófono que no puede funcionar, porque standalone no transcribe | Ocultarlo |
| Los proyectos nuevos empiezan en `/` | Empezar en la carpeta de trabajo configurada |
| Sin icono, retardo de 300 ms al tocar, marco del navegador | Icono, respuesta inmediata y manifest para pantalla completa |

Para la interfaz sin tocar, usa `agy-remote --no-mobile-patches`.

<div align="center">
<table><tr>
<td align="center"><img src="docs/assets/patch-models.png" width="190" alt="Selección de modelo" /></td>
<td align="center"><img src="docs/assets/patch-effort.png" width="190" alt="Nivel de esfuerzo" /></td>
<td align="center"><img src="docs/assets/settings.png" width="190" alt="Ajustes" /></td>
</tr></table>
</div>

## Seguridad

Quien llega a Antigravity puede leer tus archivos y ejecutar comandos, así que trata
el acceso como equivalente a un shell.

- Las contraseñas se hashean con PBKDF2-SHA256 a 200 mil iteraciones y nunca se
  guardan en claro.
- Los tokens de sesión son 256 bits aleatorios y al disco solo van sus hashes, así que
  un `sessions.json` copiado no sirve. `agy-remote sessions revoke` cierra todo.
- El login admite cinco fallos por IP cada cinco minutos, y luego un bloqueo que se
  duplica hasta 30 minutos. Hay también un límite global para intentos distribuidos. El
  bloqueo aplica incluso a la contraseña correcta.
- Las cookies son `HttpOnly`, `SameSite=Lax`, y `Secure` cuando la petición llegó por
  HTTPS.
- El panel de control, con el QR y el botón de apagado, escucha solo en loopback y en
  un puerto aparte. Nunca se expone a la red.

Detrás de un proxy inverso, indica en qué peers confías o se ignorarán las cabeceras
forwarded:

```bash
agy-remote serve --public-url https://agy.example.com --trusted-proxies 127.0.0.1/32
```

El resto está en [SECURITY.md](SECURITY.md). En resumen: mete la máquina en Tailscale
si puedes, y usa HTTPS si no.

## Comandos

```
agy-remote                     comparte la app de escritorio en tu red
agy-remote serve               ejecuta sin interfaz en un servidor
agy-remote doctor              revisa todo y dice qué está mal
agy-remote config [flags]      escribe opciones en config.json
agy-remote passwd [password]   establece la contraseña
agy-remote sessions [revoke]   lista o desconecta dispositivos
```

Flags útiles: `--port`, `--public-url`, `--workspace-root`, `--trusted-proxies`,
`--session-days`, `--no-mobile-patches`, `--language-server`. Cada una tiene su
variable de entorno `AGY_*`, y `agy-remote help` las lista todas.

## Cómo funciona

```
   móvil                      tu máquina o servidor
┌──────────┐              ┌────────────────────────────────┐
│navegador │  contraseña  │ agy-remote                     │
│          │◄────────────►│   sesiones, rate limiting      │
│  Anti-   │    :8765     │   patch main.js / index.html   │
│ gravity  │              │              │ https           │
│   UI     │              │   language_server --standalone │
└──────────┘              └──────────────┼─────────────────┘
                                         ▼
                                  Google CloudCode
```

Los prompts y el código pasan como bytes reenviados y no van a ningún sitio salvo al
language server del mismo host. El tráfico de Antigravity con Google no cambia.

## Preguntas frecuentes

**¿Sirve el IDE o la CLI?**
No, hace falta la app de escritorio, la que se llama solo "Antigravity". Únicamente
ella ejecuta `language_server --standalone`, el modo que sirve la interfaz web. El
binario de la CLI lleva el bundle compilado dentro, pero no tiene ninguna flag para
servirlo.

**¿Seguirá funcionando tras las actualizaciones?**
El proxy sí. Algún parche puede dejar de aplicar. `base-url-origin` es el único que el
acceso remoto necesita de verdad, los demás son comodidad. Abre una issue cuando uno
se rompa.

**¿Mi código va a algún sitio?**
No. El proxy y el language server corren en tu máquina.

**¿Por qué HTTP vale en la red local y no en internet?**
Los navegadores móviles rechazan certificados autofirmados, y no hay certificado
válido para `192.168.x.x`. En una red de confianza el intercambio es aceptable, en una
dirección pública no. Usa `--public-url` detrás de Caddy o de un túnel.

## Desarrollo

```bash
go test ./...
go run ./cmd/agy-remote
```

La parte que merece leerse es [`internal/patches`](internal/patches). Un parche es una
struct con una cadena de anclaje, y añadir uno a `All()` es suficiente: los tests,
`doctor`, el panel de control y la clave de caché leen todos de esa lista.

El bundle de Antigravity no está en este repositorio, así que los parches se prueban a
dos niveles. `patches_test.go` construye un documento sintético desde el registro para
probar el motor. `live_test.go` descarga el bundle real de un language server en
ejecución y comprueba que cada anclaje coincide exactamente una vez, saltándose si no
hay nada corriendo. Antes de publicar, abre la app de escritorio y ejecuta:

```bash
go test ./internal/patches -run Live -v
```

## Licencia

[Apache-2.0](LICENSE). No es un proyecto de Google ni está afiliado a Google. Lee
[DISCLAIMER.md](DISCLAIMER.md).

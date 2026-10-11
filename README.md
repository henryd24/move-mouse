# move-mouse

Mueve el mouse de forma configurable para mantener la sesión activa. Escrito en Go, con una CLI y una app gráfica (Wails).

## Características

- Modos: `random` (saltos con trayectorias curvas tipo Bézier) y `jiggle` (micro movimientos discretos).
- Intervalos, velocidad y duración total configurables; `--region` para limitar el área.
- Click y scroll opcionales.
- Se detiene con cualquier tecla, con Ctrl+C/SIGTERM, tras un tiempo (`--duration`) o si tú mueves el mouse (`--stop-on-mouse`).
- Horario (`--from/--to`, `--days`) y `--keep-awake` (evita suspensión: caffeinate / systemd-inhibit / SetThreadExecutionState).
- Archivo de configuración TOML y log a archivo.
- App gráfica en `app/` con todas las opciones (modo, región, horario por días, click/scroll, keep-awake), log en vivo y guardado automático de la configuración.
  - Detención por movimiento del mouse y/o **cualquier tecla** (global, aunque la ventana no tenga foco).
  - **Tecla global** configurable (ej. `Ctrl+Shift+M`) para iniciar/detener desde cualquier aplicación.

## Uso rápido

```bash
move-mouse                                   # equivalente al comportamiento original
move-mouse --mode jiggle --min 20s --max 60s --stop-on-mouse
move-mouse --from 09:00 --to 18:00 --days mon,tue,wed,thu,fri --keep-awake --stop-key none
move-mouse --region 0,0,800,600 --duration 90m --click
move-mouse --print-config > ~/.config/move-mouse/config.toml   # genera configuración
move-mouse --help
```

Precedencia: valores por defecto < archivo TOML < flags. El archivo por defecto es `os.UserConfigDir()/move-mouse/config.toml` (o `--config ruta`).

```toml
mode = "jiggle"
min_interval = "20s"
max_interval = "60s"
stop_on_mouse = true
keep_awake = true

[schedule]
start = "09:00"
end = "18:00"
days = ["mon", "tue", "wed", "thu", "fri"]
```

## Compilar (Makefile)

Requisitos: Go 1.24+, `make` y, en Linux, las librerías de sistema (`make deps-linux`; para la GUI `make deps-gui-linux`).

```bash
make help          # lista todos los targets
make build         # CLI en dist/
make gui-build     # app gráfica en dist/
make test lint
make package gui-package   # .tar.gz/.zip + sha256
make deb rpm       # requiere nfpm
make install       # ~/.local/bin
```

## Ejecutar como servicio

Plantillas en [`packaging/`](packaging): systemd (usuario), launchd (macOS) y tarea programada (Windows). Usan `--stop-key none`.

## CI / Releases

- `.github/workflows/ci.yml`: lint, tests y build (CLI + GUI) en Linux, macOS y Windows.
- `.github/workflows/release.yml`: al subir un tag `v*` compila en cada SO, genera `.tar.gz`/`.zip`, `.deb`, `.rpm` y `SHA256SUMS`, y publica un GitHub Release.

```bash
git tag v0.1.0 && git push origin v0.1.0
```

## Notas

- Wayland: el control de mouse no funciona de forma fiable; usa una sesión X11.
- macOS: concede permisos de Accesibilidad al binario. Los binarios no están firmados/notarizados.
- La tecla global y la detención por "cualquier tecla" global solo existen en la app gráfica (usa un hook de teclado nativo que necesita sesión gráfica; en macOS pide permiso de Accesibilidad). La CLI no carga ese hook para poder correr sin pantalla: ahí usa "cualquier tecla" en la terminal, `--stop-on-mouse` o señales. El campo `hotkey` del TOML solo lo usa la GUI.

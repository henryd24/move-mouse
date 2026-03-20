# move-mouse

Proyecto con dos implementaciones para mover el mouse aleatoriamente hasta que presiones una tecla:

- Python: [py-version](py-version)
- Go: [go-version](go-version)

## Python: uso

Requisitos:

- Python 3.13+

Pasos desde la raíz del repositorio:

```bash
cd py-version
python -m venv .venv
source .venv/bin/activate
pip install -U pip
pip install -r requirements.txt
python main.py
```

Funcionamiento:

- El cursor se mueve a posiciones aleatorias cada 1 a 5 segundos.
- Al presionar cualquier tecla, el programa termina.

## Go: compilación y uso

Requisitos:

- Go 1.24+

Compilar y ejecutar desde código fuente:

```bash
cd go-version
go mod download
go run .
```

Compilar binario:

```bash
cd go-version
go build -o move-mouse .
./move-mouse
```

## Go: instalar binario en PATH (recomendado)

Se recomienda instalarlo en `~/.local/bin`:

```bash
cd go-version
go build -o move-mouse .
mkdir -p ~/.local/bin
mv move-mouse ~/.local/bin/
```

Asegura que `~/.local/bin` este en tu `PATH` (Linux):

```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

Si usas zsh, agrega la linea en `~/.zshrc` en lugar de `~/.bashrc`.

Luego podras ejecutar:

```bash
move-mouse
```

## Nota

Las librerias de control de mouse/teclado pueden requerir permisos del sistema segun tu entorno de escritorio o sistema operativo.

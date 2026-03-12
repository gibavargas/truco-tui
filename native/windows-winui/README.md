# Truco - Microsoft Windows WinUI 3 Native Client

Este é o cliente nativo do Windows para o jogo Truco, implementado com **C#** e **WinUI 3 (Windows App SDK)**. A lógica do jogo vem de um core em **Go**, carregado via FFI.

## Requisitos
- **Windows 10** (1809 ou superior) ou **Windows 11**.
- **.NET 8 SDK**.
- **Visual Studio 2022** com toolchain C++ quando for necessário gerar a DLL FFI.
- **Go** instalado no PATH.

## Como compilar

### Gerando a DLL FFI
A DLL consumida pelo cliente WinUI é gerada a partir de `cmd/truco-core-ffi`:

```bash
go build -buildmode=c-shared -o bin/truco-core-ffi.dll ./cmd/truco-core-ffi
```

### Publicando o cliente WinUI portátil (Windows x64)
Com a DLL pronta, publique o cliente e envie a saída final para a pasta de binários guirias:

```bash
dotnet publish native/windows-winui/TrucoWinUI.csproj -c Release -r win-x64 --self-contained true -o bin/gui/winui/truco-gui-winui-windows-amd64-portable
```

O diretório em `bin/gui/winui` segue a política descrita em `docs/BINARY_NAMING.md` e já está pronto para distribuição portátil.

## Limitação atual em Windows ARM64
No toolchain Go atual, `-buildmode=c-shared` não é suportado para Windows ARM64. Por isso, o cliente WinUI não pode ser compilado de forma nativa em uma máquina Windows ARM64 usando este bridge FFI.

Nessa plataforma, o script raiz `build-portable.bat` faz fallback para o executável TUI nativo:

```powershell
.\build-portable.bat
```

Saída em ARM64:
- `bin\tui\truco-tui-core-windows-arm64-portable.exe`

## Funcionalidades
- UI nativa WinUI 3 com Fluent Design.
- Jogo offline.
- Multiplayer online.
- Chat e convites integrados.

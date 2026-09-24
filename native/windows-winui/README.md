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

O fluxo recomendado para produção é o script da raiz:

```powershell
.\build-portable.bat
```

Ele gera a DLL FFI, publica o WinUI 3 em modo self-contained e copia `truco-core-ffi.dll` para o bundle final. Se a DLL for carregada por `TRUCO_CORE_LIB`, mantenha no mesmo diretório as dependências nativas do MinGW: `libgcc_s_seh-1.dll`, `libstdc++-6.dll` e `libwinpthread-1.dll`.

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
- Jogo offline de 2 ou 4 jogadores usando o contrato `SnapshotBundle`.
- Host/join online direto ou via relay, com seleção de transporte e papel desejado.
- Chat, voto de host, convites de reposição e fechamento de sessão via intents do runtime.
- Ações de partida com paridade de contrato: jogar carta, jogar virada, pedir/aumentar truco, aceitar, recusar, nova mão e pulso de CPU.
- Painel de diagnóstico com versões do core/protocolo/schema, backlog, seed e log de intents.
- Validação de versão do runtime FFI no startup para rejeitar DLLs incompatíveis.

## Usabilidade e acessibilidade
- Metadados de acessibilidade (`AutomationProperties`) nos comandos principais, assentos do lobby, convites e acoes de partida.
- Atalhos de teclado para operacoes frequentes:
  - `Ctrl+D`: mostrar ou ocultar detalhes da mesa
  - `Ctrl+L`: alternar idioma
  - `Ctrl+W`: encerrar sessao
  - `Ctrl+T`: pedir ou aumentar truco
  - `Ctrl+A`: aceitar truco
  - `Ctrl+R`: recusar truco
  - `Ctrl+N`: iniciar nova mao
  - `Ctrl+P`: pulso de CPU no painel de diagnostico
  - `Ctrl+Enter`: iniciar partida no lobby do host
  - `Enter` no campo de chat: enviar mensagem
- O layout move o painel lateral para baixo quando a janela fica mais estreita, reduz paddings e empilha os comandos do cabeçalho para preservar legibilidade.

## Validação em host não Windows
- Em macOS ou Linux, ainda é útil validar o cliente de forma parcial:
  - `dotnet build native/windows-winui/TrucoWinUI.csproj -p:EnableWindowsTargeting=true`
  - `go test ./...`
- O build final do WinUI **precisa** de um host Windows ou de CI Windows. Mesmo com `EnableWindowsTargeting=true`, o Windows App SDK invoca `XamlCompiler.exe`, que não executa em macOS.
- Trate a validação fora do Windows como checagem de código/restore/contrato. Compile e empacote a entrega final no Windows antes de publicar.

# Truco - macOS Native Client

Este é o cliente nativo de macOS para o jogo Truco, construído com **SwiftUI** e utilizando um back-end compartilhado em **Go** via FFI (Foreign Function Interface).

## Requisitos
- **macOS** 13.0 (Ventura) ou superior.
- **Xcode** 15 ou superior.
- **Go** 1.22 ou superior (para compilar a engine compartilhada `appcore`).

## Compilando o Core em Go

Antes de compilar o aplicativo macOS, você precisa compilar a engine compartilhada em um formato utilizável pelo Xcode (`XCFramework` ou arquivo `.a` estático com header de C). O script de build lidará com essa dependência gerando as bibliotecas no diretório correto.

1. No terminal, vá até a pasta raiz do projeto.
2. Compile a biblioteca compartilhada para macOS (em caso de ausência dos scripts automatizados):
   ```bash
   cd internal/appcore/ffi
   go build -buildmode=c-archive -o ../../../native/macos/Truco/trucocore.a trucocore.go
   ```
*(Opcional: você também pode usar o `gomobile` se for integrar o projeto como xcframework no iOS).*

## Instalando e Rodando no Xcode

1. Abra o arquivo `Truco.xcodeproj` no **Xcode**.
   ```bash
   open native/macos/Truco/Truco.xcodeproj
   ```
2. No Xcode, selecione um simulador de Mac, ou "My Mac" como o `Run Destination`.
3. Verifique se o arquivo `trucocore.a` e o header gerado (`trucocore.h`) estão vinculados ao projeto em *Frameworks, Libraries, and Embedded Content*.
4. Clique em **Run** (`Cmd + R`) para compilar e iniciar o aplicativo.

## Funcionalidades
- Interface 100% Nativa em SwiftUI (com efeito Glassmorphism / Materiais Fluidos).
- Modo Offline vs CPU.
- Multiplayer Online: Sistema de Salas com códigos de convite.
- Chat em tempo real embutido.
- Votação de Troca de Host e Substituição por CPU.
- Acompanhamento do placar, vira e manilha.

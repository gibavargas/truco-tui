Original prompt: Faça a passada final, dê commit e push e nas demais alterações também se você entender pertinente

- Rodada consolidada em um único commit:
  - compatibilidade v1/v2 e snapshot de rede em `appcore`/`netp2p`
  - remoção de `native/windows-wpf/`
  - alinhamento de browser/macOS/GTK/WinUI no lobby online
  - Browser Edition com home online/offline explícita, auto-sync, melhoria visual e correção de `pollEvents`
- Validação concluída:
  - `go test ./...`
  - Playwright offline: home -> game -> ação válida sem refresh manual
  - Playwright online: host/guest em contexts separados, lobby com resumo de rede, início da partida e chat sem refresh manual
- Observação:
  - WinUI não foi recompilado nesta máquina macOS; a cobertura dessa rodada para desktop Windows continua no nível de mudança de código.

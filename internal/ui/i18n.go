package ui

import (
	"strings"
	"sync"
)

type locale string

const (
	localePTBR locale = "pt-BR"
	localeENUS locale = "en-US"
)

var activeLocale = localePTBR
var localeMu sync.RWMutex

var translations = map[locale]map[string]string{
	localePTBR: {
		"app_title":                            "TRUCO PAULISTA - TUI EM GO",
		"menu_offline":                         "Offline (com CPU)",
		"menu_online":                          "Multiplayer P2P (chave de sessão)",
		"menu_exit":                            "Sair",
		"menu_invalid_option":                  "Opção inválida.",
		"prompt_players_count":                 "Quantidade de jogadores (2/4)",
		"prompt_your_name":                     "Seu nome",
		"default_you":                          "Você",
		"prompt_player_cpu":                    "Jogador %d é CPU? (s/n)",
		"prompt_player_name":                   "Nome do jogador %d",
		"default_player_name":                  "Jogador-%d",
		"starting_bubble_tea":                  "Iniciando Bubble Tea TUI...",
		"bubble_tea_run_error":                 "erro no Bubble Tea: %w",
		"press_enter_continue":                 "Pressione Enter para continuar...",
		"error_prefix":                         "Erro: ",
		"error_invalid_player_count":           "quantidade inválida",
		"lobby_title":                          "TRUCO PAULISTA",
		"lobby_help":                           "↑/↓ navega  •  Enter confirma  •  q sai",
		"lobby_unknown_command":                "comando inválido",
		"lobby_option_prefix":                  "Opção %d",
		"lobby_choose_offline":                 "offline",
		"lobby_choose_online":                  "online",
		"lobby_choose_language":                "idioma",
		"lobby_choose_exit":                    "sair",
		"menu_language":                        "Idioma / Language",
		"language_prompt":                      "Idioma (pt-BR/en-US)",
		"language_changed":                     "Idioma alterado para %s.",
		"language_invalid":                     "Idioma inválido.",
		"help_commands":                        "Comandos:",
		"online_title":                         "Multiplayer P2P",
		"online_create_host":                   "Criar host",
		"online_join":                          "Seguir partida (join)",
		"online_back":                          "Voltar",
		"online_choice":                        "Escolha",
		"host_name_prompt":                     "Seu nome (host)",
		"host_default_name":                    "Host",
		"host_players_prompt":                  "Partida para 2 ou 4 jogadores",
		"host_lobby_title":                     "=================== LOBBY HOST ===================",
		"host_share_key":                       "Compartilhe esta chave com os convidados:",
		"line_sep":                             "--------------------------------------------------",
		"slot_empty":                           "(vazio)",
		"slot_format":                          "Slot %d: %s",
		"lobby_full_format":                    "Lobby cheio: %v",
		"host_commands":                        "Comandos: chat <msg> | start | refresh | back",
		"host_prompt":                          "host>",
		"join_name_prompt":                     "Seu nome",
		"join_default_name":                    "Jogador",
		"join_key_prompt":                      "Cole a chave do host",
		"join_role_prompt":                     "Papel (partner/opponent/auto)",
		"join_lobby_title":                     "=================== LOBBY CLIENT =================",
		"join_slot_format":                     "Seu slot: %d",
		"join_commands":                        "Comandos: chat <msg> | refresh | back",
		"join_prompt":                          "join>",
		"timeout_initial_state":                "timeout aguardando estado inicial da partida",
		"lobby_shortcuts":                      "[1-3] escolhe  •  [0]/[q] sai",
		"chat_welcome":                         "Bem-vindo ao Truco TUI (Bubble Tea Edition)!",
		"chat_online_started":                  "Partida online iniciada.",
		"online_invalid_action_prefix":         "Ação inválida: ",
		"online_unknown_action_format":         "ação desconhecida: %s",
		"online_host_closed_match":             "Host encerrou a partida.",
		"online_failover_promoted":             "[system] Handoff concluído: você assumiu como novo host.",
		"online_failover_rejoined":             "[system] Reconectado automaticamente ao novo host.",
		"online_cmd_use_host":                  "use /host <slot>",
		"online_cmd_use_invite":                "use /invite <slot>",
		"online_cmd_invalid_slot":              "slot inválido",
		"online_provisional_cpu_on":            "[system] %s virou CPU provisório.",
		"online_provisional_cpu_off":           "[system] %s retomou controle humano.",
		"ui_loading":                           "Carregando...",
		"header_title":                         "♠ TRUCO PAULISTA - MESA ♣",
		"score_stake":                          "Aposta",
		"score_stake_ladder":                   "Truco",
		"score_turn":                           "Vez",
		"score_flip_prefix":                    "Vira ",
		"score_trump_prefix":                   "Manilha ",
		"overlay_trick_end_title":              "Fim da vaza",
		"overlay_stake_in_dispute":             "Aposta em disputa",
		"overlay_collecting_trick":             "Recolhendo vaza...",
		"overlay_collecting_trick_by_format":   "Recolhendo vaza para %s...",
		"tabs_label":                           "Abas: ",
		"tab_mesa":                             "mesa",
		"tab_chat":                             "chat",
		"tab_log":                              "log",
		"panel_chat_desc":                      "CHAT (offline): mensagens locais da sessao.",
		"panel_chat_hint":                      "Digite na aba CHAT e pressione Enter para enviar.",
		"panel_chat_message_prefix":            "Mensagem: ",
		"panel_log_title":                      "LOG DA PARTIDA:",
		"panel_error_log_title":                "Erros recentes:",
		"panel_recent_score_prefix":            "Pontuacao recente: ",
		"panel_table_label":                    "MESA",
		"panel_tricks_label":                   "vazas",
		"panel_round_label":                    "rodada",
		"panel_history_prefix":                 "Historico: ",
		"panel_trump_label":                    "Manilha",
		"panel_flip_label":                     "Vira",
		"panel_stake_label":                    "Aposta",
		"panel_raise_pending_format":           "Pedido pendente: %s (%d) por %s",
		"ui_turn_short":                        "vez",
		"turn_badge_prefix":                    "Vez: ",
		"score_flip_label":                     "Vira",
		"table_empty":                          "(mesa vazia)",
		"table_shuffling_dealing":              "Embaralhando e distribuindo...",
		"table_no_cards":                       "(sem cartas)",
		"ui_you_label":                         "voce",
		"status_match_over_format":             "🏆  PARTIDA ENCERRADA - Time %d venceu!  Pressione 'q' para sair.",
		"status_truco_response":                "⚠  TRUCO! Responda: [a]ceitar  [r]ecusar",
		"status_truco_response_format":         "⚠  %s (%d)! Responda: [a]ceitar  [r]ecusar  [t] subir",
		"status_truco_response_owner_format":   "⚠  %s (%d) por %s: %s responde.",
		"status_truco_wait_format":             "⚠  %s (%d) pendente por %s",
		"status_truco_wait_owner_format":       "⚠  %s (%d) por %s: aguardando %s.",
		"truco_call_truco":                     "truco",
		"truco_call_six":                       "seis",
		"truco_call_nine":                      "nove",
		"truco_call_twelve":                    "doze",
		"status_turn_format":                   "Vez de: %s",
		"help_play_cards":                      "[1][2][3] jogar carta",
		"help_play_cards_short":                "jogar",
		"help_truco_short":                     "truco/subir",
		"help_accept":                          "[a] aceitar",
		"help_refuse":                          "[r] recusar",
		"help_answer_short":                    "responder",
		"help_vote_host_short":                 "votar host",
		"help_invite_short":                    "substituir",
		"help_invite_request_short":            "solicitar convite",
		"help_send_short":                      "enviar",
		"help_tab_short":                       "abas",
		"help_quit":                            "[q] sair",
		"help_quit_short":                      "sair",
		"ui_role_you":                          "você",
		"ui_role_partner":                      "parceiro",
		"ui_role_opponent":                     "adversário",
		"ui_role_mao":                          "mão",
		"ui_role_pe":                           "pé",
		"ui_role_turn":                         "vez",
		"ui_role_host":                         "host",
		"ui_role_cpu":                          "cpu",
		"ui_role_cpu_prov":                     "cpu provisório",
		"role_lane_you_prefix":                 "Você:",
		"role_lane_host_prefix":                "Host:",
		"role_lane_raise_pending_format":       "Truco por %s: %s (%d)",
		"panel_chat_commands_offline":          "Comandos: chat local",
		"panel_chat_commands_online_host":      "Comandos: chat | /host <slot> | /invite <slot>",
		"panel_chat_commands_online_client":    "Comandos: chat | /host <slot> | /invite <slot>",
		"table_trick_history_empty":            "V1: ... | V2: ... | V3: ...",
		"table_trick_prefix":                   "V",
		"table_tie_word":                       "empate",
		"effect_trick_tie_format":              "Vaza %d empatou.",
		"effect_trick_point_own_team_format":   "Vaza %d: ponto da sua equipe.",
		"effect_trick_point_enemy_team_format": "Vaza %d: ponto da equipe adversaria.",
		"log_hand_ended_prefix":                "Mão encerrada:",
		"log_match_ended_prefix":               "Partida encerrada:",
	},
	localeENUS: {
		"app_title":                            "TRUCO PAULISTA - GO TUI",
		"menu_offline":                         "Offline (with CPU)",
		"menu_online":                          "P2P Multiplayer (session key)",
		"menu_exit":                            "Exit",
		"menu_invalid_option":                  "Invalid option.",
		"prompt_players_count":                 "Number of players (2/4)",
		"prompt_your_name":                     "Your name",
		"default_you":                          "You",
		"prompt_player_cpu":                    "Is player %d a CPU? (y/n)",
		"prompt_player_name":                   "Player %d name",
		"default_player_name":                  "Player-%d",
		"starting_bubble_tea":                  "Starting Bubble Tea TUI...",
		"bubble_tea_run_error":                 "Bubble Tea error: %w",
		"press_enter_continue":                 "Press Enter to continue...",
		"error_prefix":                         "Error: ",
		"error_invalid_player_count":           "invalid player count",
		"lobby_title":                          "TRUCO PAULISTA",
		"lobby_help":                           "↑/↓ navigate  •  Enter confirm  •  q exit",
		"lobby_unknown_command":                "invalid command",
		"lobby_option_prefix":                  "Option %d",
		"lobby_choose_offline":                 "offline",
		"lobby_choose_online":                  "online",
		"lobby_choose_language":                "language",
		"lobby_choose_exit":                    "exit",
		"menu_language":                        "Language / Idioma",
		"language_prompt":                      "Language (pt-BR/en-US)",
		"language_changed":                     "Language changed to %s.",
		"language_invalid":                     "Invalid language.",
		"help_commands":                        "Commands:",
		"online_title":                         "P2P Multiplayer",
		"online_create_host":                   "Create host",
		"online_join":                          "Join match",
		"online_back":                          "Back",
		"online_choice":                        "Choose",
		"host_name_prompt":                     "Host name",
		"host_default_name":                    "Host",
		"host_players_prompt":                  "Match for 2 or 4 players",
		"host_lobby_title":                     "=================== HOST LOBBY ===================",
		"host_share_key":                       "Share this key with guests:",
		"line_sep":                             "--------------------------------------------------",
		"slot_empty":                           "(empty)",
		"slot_format":                          "Slot %d: %s",
		"lobby_full_format":                    "Lobby full: %v",
		"host_commands":                        "Commands: chat <msg> | start | refresh | back",
		"host_prompt":                          "host>",
		"join_name_prompt":                     "Your name",
		"join_default_name":                    "Player",
		"join_key_prompt":                      "Paste host key",
		"join_role_prompt":                     "Role (partner/opponent/auto)",
		"join_lobby_title":                     "=================== CLIENT LOBBY =================",
		"join_slot_format":                     "Your slot: %d",
		"join_commands":                        "Commands: chat <msg> | refresh | back",
		"join_prompt":                          "join>",
		"timeout_initial_state":                "timeout waiting for initial game state",
		"lobby_shortcuts":                      "[1-3] choose  •  [0]/[q] exit",
		"chat_welcome":                         "Welcome to Truco TUI (Bubble Tea Edition)!",
		"chat_online_started":                  "Online match started.",
		"online_invalid_action_prefix":         "Invalid action: ",
		"online_unknown_action_format":         "unknown action: %s",
		"online_host_closed_match":             "Host closed the match.",
		"online_failover_promoted":             "[system] Handoff complete: you are now the new host.",
		"online_failover_rejoined":             "[system] Reconnected automatically to the new host.",
		"online_cmd_use_host":                  "use /host <slot>",
		"online_cmd_use_invite":                "use /invite <slot>",
		"online_cmd_invalid_slot":              "invalid slot",
		"online_provisional_cpu_on":            "[system] %s became provisional CPU.",
		"online_provisional_cpu_off":           "[system] %s resumed human control.",
		"ui_loading":                           "Loading...",
		"header_title":                         "♠ TRUCO PAULISTA - TABLE ♣",
		"score_stake":                          "Stake",
		"score_stake_ladder":                   "Truco",
		"score_turn":                           "Turn",
		"score_flip_prefix":                    "Flip ",
		"score_trump_prefix":                   "Trump ",
		"overlay_trick_end_title":              "Trick ended",
		"overlay_stake_in_dispute":             "Stake in dispute",
		"overlay_collecting_trick":             "Collecting trick...",
		"overlay_collecting_trick_by_format":   "Collecting trick to %s...",
		"tabs_label":                           "Tabs: ",
		"tab_mesa":                             "table",
		"tab_chat":                             "chat",
		"tab_log":                              "log",
		"panel_chat_desc":                      "CHAT (offline): local session messages.",
		"panel_chat_hint":                      "Type in CHAT and press Enter to send.",
		"panel_chat_message_prefix":            "Message: ",
		"panel_log_title":                      "MATCH LOG:",
		"panel_error_log_title":                "Recent errors:",
		"panel_recent_score_prefix":            "Recent score: ",
		"panel_table_label":                    "TABLE",
		"panel_tricks_label":                   "tricks",
		"panel_round_label":                    "round",
		"panel_history_prefix":                 "History: ",
		"panel_trump_label":                    "Trump",
		"panel_flip_label":                     "Flip",
		"panel_stake_label":                    "Stake",
		"panel_raise_pending_format":           "Pending call: %s (%d) by %s",
		"ui_turn_short":                        "turn",
		"turn_badge_prefix":                    "Turn: ",
		"score_flip_label":                     "Flip",
		"table_empty":                          "(empty table)",
		"table_shuffling_dealing":              "Shuffling and dealing...",
		"table_no_cards":                       "(no cards)",
		"ui_you_label":                         "you",
		"status_match_over_format":             "🏆  MATCH OVER - Team %d won!  Press 'q' to quit.",
		"status_truco_response":                "⚠  TRUCO! Respond: [a]ccept  [r]efuse",
		"status_truco_response_format":         "⚠  %s (%d)! Respond: [a]ccept  [r]efuse  [t] raise",
		"status_truco_response_owner_format":   "⚠  %s (%d) by %s: %s responds.",
		"status_truco_wait_format":             "⚠  %s (%d) pending by %s",
		"status_truco_wait_owner_format":       "⚠  %s (%d) by %s: waiting for %s.",
		"truco_call_truco":                     "truco",
		"truco_call_six":                       "six",
		"truco_call_nine":                      "nine",
		"truco_call_twelve":                    "twelve",
		"status_turn_format":                   "Turn: %s",
		"help_play_cards":                      "[1][2][3] play card",
		"help_play_cards_short":                "play",
		"help_truco_short":                     "truco/raise",
		"help_accept":                          "[a] accept",
		"help_refuse":                          "[r] refuse",
		"help_answer_short":                    "respond",
		"help_vote_host_short":                 "vote host",
		"help_invite_short":                    "replace",
		"help_invite_request_short":            "request invite",
		"help_send_short":                      "send",
		"help_tab_short":                       "tabs",
		"help_quit":                            "[q] quit",
		"help_quit_short":                      "quit",
		"ui_role_you":                          "you",
		"ui_role_partner":                      "partner",
		"ui_role_opponent":                     "opponent",
		"ui_role_mao":                          "mao",
		"ui_role_pe":                           "pe",
		"ui_role_turn":                         "turn",
		"ui_role_host":                         "host",
		"ui_role_cpu":                          "cpu",
		"ui_role_cpu_prov":                     "provisional cpu",
		"role_lane_you_prefix":                 "You:",
		"role_lane_host_prefix":                "Host:",
		"role_lane_raise_pending_format":       "Truco by %s: %s (%d)",
		"panel_chat_commands_offline":          "Commands: local chat",
		"panel_chat_commands_online_host":      "Commands: chat | /host <slot> | /invite <slot>",
		"panel_chat_commands_online_client":    "Commands: chat | /host <slot> | /invite <slot>",
		"table_trick_history_empty":            "T1: ... | T2: ... | T3: ...",
		"table_trick_prefix":                   "T",
		"table_tie_word":                       "tie",
		"effect_trick_tie_format":              "Trick %d tied.",
		"effect_trick_point_own_team_format":   "Trick %d: point for your team.",
		"effect_trick_point_enemy_team_format": "Trick %d: point for the opponent team.",
		"log_hand_ended_prefix":                "Hand ended:",
		"log_match_ended_prefix":               "Match ended:",
	},
}

func tr(key string) string {
	l := currentLocale()
	if m, ok := translations[l]; ok {
		if v, ok := m[key]; ok {
			return v
		}
	}
	return key
}

func normalizeLocaleCode(code string) locale {
	switch strings.ToLower(strings.TrimSpace(code)) {
	case "pt", "pt-br", "pt_br":
		return localePTBR
	case "en", "en-us", "en_us":
		return localeENUS
	default:
		return ""
	}
}

func setLocale(code string) bool {
	l := normalizeLocaleCode(code)
	if l == "" {
		return false
	}
	localeMu.Lock()
	activeLocale = l
	localeMu.Unlock()
	return true
}

func localeCode() string {
	return string(currentLocale())
}

func allTranslationsForKey(key string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(translations))
	for _, byKey := range translations {
		v, ok := byKey[key]
		if !ok || v == "" {
			continue
		}
		if _, exists := seen[v]; exists {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func currentLocale() locale {
	localeMu.RLock()
	defer localeMu.RUnlock()
	return activeLocale
}

func Translate(key string) string {
	return tr(key)
}

func SetLocale(code string) bool {
	return setLocale(code)
}

func LocaleCode() string {
	return localeCode()
}

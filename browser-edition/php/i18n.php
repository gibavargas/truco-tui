<?php
/**
 * i18n — Internationalisation strings for the PHP frontend.
 * Ported from the JavaScript i18n object.
 */

$I18N = [
    'pt-BR' => [
        'title_main' => 'Truco Browser Edition',
        'title_sub' => 'Mesa online e offline com runtime Go',
        'locale_label' => 'Idioma',
        'setup_title' => 'Nova partida',
        'setup_kicker' => 'Mesa pronta',
        'setup_intro_title' => 'Escolha como quer abrir a próxima rodada',
        'setup_intro_body' => 'Jogue contra CPUs, hospede uma sala online ou entre por convite sem trocar de cliente.',
        'setup_live_sync' => 'Atualização viva',
        'setup_name' => 'Seu nome',
        'setup_name_hint' => 'Nome exibido na mesa e no lobby.',
        'setup_players' => 'Jogadores',
        'setup_start' => 'Iniciar partida',
        'setup_offline_title' => 'Offline imediato',
        'setup_offline_body' => 'Abra uma mesa local contra CPUs e entre direto no feltro.',
        'setup_host_title' => 'Criar sala',
        'setup_host_body' => 'Vire host, gere o convite e acompanhe a compatibilidade da mesa.',
        'setup_host_start' => 'Criar sala online',
        'setup_host_hint' => 'O convite aparece no lobby assim que a sala sobe.',
        'setup_join_title' => 'Entrar em sala',
        'setup_join_body' => 'Cole um convite, escolha o papel e entre direto no lobby.',
        'setup_join_start' => 'Entrar no lobby',
        'setup_join_hint' => 'No modo automático, o runtime escolhe o melhor assento disponível.',
        'setup_join_key_placeholder' => 'Cole a chave de convite',
        'setup_help' => 'Clique nos botões para jogar. Sem atalhos de teclado.',
        'team1' => 'Time 1',
        'team2' => 'Time 2',
        'stake' => 'Aposta',
        'vira' => 'Vira',
        'manilha' => 'Manilha',
        'status_title' => 'Status',
        'tricks_title' => 'Vazas',
        'help_title' => 'Controles',
        'log_title' => 'Log da partida',
        'btn_truco' => 'Truco',
        'btn_raise' => 'Subir',
        'btn_accept' => 'Aceitar',
        'btn_refuse' => 'Recusar',
        'btn_new_hand' => 'Nova mão',
        'hand_title' => 'Sua mão',
        'btn_play_again' => 'Jogar novamente',
        'status_ready' => 'Pronto.',
        'status_your_turn' => 'Sua vez. Escolha uma carta.',
        'status_wait_cpu' => 'Aguardando %s...',
        'status_match_end' => 'Partida encerrada.',
        'status_pending_you' => '%s (%s) pendente: aceitar, recusar ou subir.',
        'status_pending_other' => 'Aguardando resposta de %s para %s (%s).',
        'decision_pending_eyebrow' => 'Resposta imediata',
        'decision_pending_title' => '%s está em jogo',
        'decision_pending_body' => 'A mão está pausada até você aceitar, recusar ou subir a aposta.',
        'decision_turn_eyebrow' => 'Sua vez',
        'decision_turn_title' => 'Jogue uma carta ou pressione a mesa',
        'decision_turn_body' => 'A rodada está pronta para a sua ação.',
        'action_hint_auto_refresh' => 'Atualização automática ativa. Use Atualizar só se algo parecer parado.',
        'overlay_match_win' => 'Você venceu!',
        'overlay_match_loss' => 'Você perdeu.',
        'overlay_match_detail' => 'Placar final: Time 1 %d x %d Time 2',
        'trick_short' => 'V',
        'trick_tie' => 'empate',
        'error_prefix' => 'Erro: ',
        'turn_of' => 'Vez de %s',
        'score_line' => 'T1 %d x %d T2',
        'cpu_tag' => 'CPU',
        'card_of' => '%s de %s',
        'suit_Ouros' => 'Ouros',
        'suit_Espadas' => 'Espadas',
        'suit_Copas' => 'Copas',
        'suit_Paus' => 'Paus',
        'setup_mode' => 'Modo',
        'setup_mode_offline' => 'Offline',
        'setup_mode_online' => 'Online (Alpha)',
        'lobby_title' => 'Lobby online',
        'lobby_start' => 'Iniciar partida',
        'lobby_refresh' => 'Atualizar',
        'lobby_leave' => 'Sair do lobby',
        'lobby_invite' => 'Convite',
        'lobby_slots_title' => 'Slots',
        'lobby_events_title' => 'Eventos',
        'network_transport_label' => 'Transporte',
        'network_compatibility_label' => 'Compatibilidade',
        'network_supported_versions_label' => 'Suporte do build',
        'network_transport_tcp_tls' => 'TCP + TLS',
        'network_transport_relay_quic_v2' => 'Relay QUIC v2',
        'network_compatibility_host_uniform' => 'todos em %s',
        'network_compatibility_host_mixed' => 'sessão mista %s',
        'network_compatibility_client' => 'negociado %s',
        'lobby_slot_empty' => 'vazio',
        'lobby_events_empty' => 'Sem eventos.',
        'lobby_chat_send' => 'Enviar',
        'chat_placeholder' => 'Digite uma mensagem...',
        'action_vote_host' => 'Votar host',
        'action_replacement_invite' => 'Convite de substituição',
        'game_events_title' => 'Eventos online',
        'setup_online_action' => 'Ação online',
        'setup_online_role' => 'Papel',
        'setup_online_key' => 'Convite',
        'online_action_host' => 'Criar host',
        'online_action_join' => 'Entrar com convite',
        'online_role_auto' => 'Auto',
        'online_role_partner' => 'Parceiro',
        'online_role_opponent' => 'Adversário',
        'back_to_setup' => 'Voltar ao início',
        'refresh' => 'Atualizar',
    ],
    'en-US' => [
        'title_main' => 'Truco Browser Edition',
        'title_sub' => 'Online and offline table backed by Go runtime',
        'locale_label' => 'Language',
        'setup_title' => 'New match',
        'setup_kicker' => 'Table ready',
        'setup_intro_title' => 'Choose how you want to open the next round',
        'setup_intro_body' => 'Play against CPUs, host an online room, or join through an invite without changing clients.',
        'setup_live_sync' => 'Live sync',
        'setup_name' => 'Your name',
        'setup_name_hint' => 'Name shown at the table and in the lobby.',
        'setup_players' => 'Players',
        'setup_start' => 'Start match',
        'setup_offline_title' => 'Instant offline',
        'setup_offline_body' => 'Open a local table against CPUs and get straight to the felt.',
        'setup_host_title' => 'Create room',
        'setup_host_body' => 'Become host, generate the invite, and track session compatibility.',
        'setup_host_start' => 'Create online room',
        'setup_host_hint' => 'The invite appears in the lobby as soon as the room is live.',
        'setup_join_title' => 'Join room',
        'setup_join_body' => 'Paste an invite, choose a role, and go straight into the lobby.',
        'setup_join_start' => 'Join lobby',
        'setup_join_hint' => 'In auto mode, the runtime picks the best open seat.',
        'setup_join_key_placeholder' => 'Paste the invite key',
        'setup_help' => 'Click the buttons to play. No keyboard shortcuts.',
        'team1' => 'Team 1',
        'team2' => 'Team 2',
        'stake' => 'Stake',
        'vira' => 'Flip',
        'manilha' => 'Trump',
        'status_title' => 'Status',
        'tricks_title' => 'Tricks',
        'help_title' => 'Controls',
        'log_title' => 'Match log',
        'btn_truco' => 'Truco',
        'btn_raise' => 'Raise',
        'btn_accept' => 'Accept',
        'btn_refuse' => 'Refuse',
        'btn_new_hand' => 'New hand',
        'hand_title' => 'Your hand',
        'btn_play_again' => 'Play again',
        'status_ready' => 'Ready.',
        'status_your_turn' => 'Your turn. Pick a card.',
        'status_wait_cpu' => 'Waiting for %s...',
        'status_match_end' => 'Match ended.',
        'status_pending_you' => '%s (%s) pending: accept, refuse or raise.',
        'status_pending_other' => 'Waiting for %s to answer %s (%s).',
        'decision_pending_eyebrow' => 'Answer now',
        'decision_pending_title' => '%s is on the line',
        'decision_pending_body' => 'The hand is paused until you accept, refuse, or raise the bet.',
        'decision_turn_eyebrow' => 'Your turn',
        'decision_turn_title' => 'Play a card or press the table',
        'decision_turn_body' => 'The round is ready for your action.',
        'action_hint_auto_refresh' => 'Auto-refresh is on. Use Refresh only if the table looks stuck.',
        'overlay_match_win' => 'You won!',
        'overlay_match_loss' => 'You lost.',
        'overlay_match_detail' => 'Final score: Team 1 %d x %d Team 2',
        'trick_short' => 'T',
        'trick_tie' => 'tie',
        'error_prefix' => 'Error: ',
        'turn_of' => 'Turn: %s',
        'score_line' => 'T1 %d x %d T2',
        'cpu_tag' => 'CPU',
        'card_of' => '%s of %s',
        'suit_Ouros' => 'Diamonds',
        'suit_Espadas' => 'Spades',
        'suit_Copas' => 'Hearts',
        'suit_Paus' => 'Clubs',
        'setup_mode' => 'Mode',
        'setup_mode_offline' => 'Offline',
        'setup_mode_online' => 'Online (Alpha)',
        'lobby_title' => 'Online lobby',
        'lobby_start' => 'Start match',
        'lobby_refresh' => 'Refresh',
        'lobby_leave' => 'Leave lobby',
        'lobby_invite' => 'Invite',
        'lobby_slots_title' => 'Slots',
        'lobby_events_title' => 'Events',
        'network_transport_label' => 'Transport',
        'network_compatibility_label' => 'Compatibility',
        'network_supported_versions_label' => 'Build support',
        'network_transport_tcp_tls' => 'TCP + TLS',
        'network_transport_relay_quic_v2' => 'Relay QUIC v2',
        'network_compatibility_host_uniform' => 'everyone on %s',
        'network_compatibility_host_mixed' => 'mixed session %s',
        'network_compatibility_client' => 'negotiated %s',
        'lobby_slot_empty' => 'empty',
        'lobby_events_empty' => 'No events.',
        'lobby_chat_send' => 'Send',
        'chat_placeholder' => 'Type a message...',
        'action_vote_host' => 'Vote host',
        'action_replacement_invite' => 'Replacement invite',
        'game_events_title' => 'Online events',
        'setup_online_action' => 'Online action',
        'setup_online_role' => 'Role',
        'setup_online_key' => 'Invite',
        'online_action_host' => 'Create host',
        'online_action_join' => 'Join with invite',
        'online_role_auto' => 'Auto',
        'online_role_partner' => 'Partner',
        'online_role_opponent' => 'Opponent',
        'back_to_setup' => 'Back to start',
        'refresh' => 'Refresh',
    ],
];

/**
 * Translate a key, with optional sprintf-style arguments.
 */
function tr(string $key, ...$args): string
{
    global $I18N;
    $locale = $_SESSION['locale'] ?? 'pt-BR';
    $dict = $I18N[$locale] ?? $I18N['pt-BR'];
    $tpl = $dict[$key] ?? $key;
    if (count($args) > 0) {
        return sprintf($tpl, ...$args);
    }
    return $tpl;
}

/**
 * Suit symbol map.
 */
function suitSymbol(string $suit): string
{
    $map = [
        'Ouros' => '♦',
        'Espadas' => '♠',
        'Copas' => '♥',
        'Paus' => '♣',
    ];
    return $map[$suit] ?? '?';
}

/**
 * Suit CSS color class.
 */
function suitColorClass(string $suit): string
{
    return ($suit === 'Copas' || $suit === 'Ouros') ? 'red' : 'black';
}

/**
 * Raise label for a given stake.
 */
function raiseLabel(int $stake, string $locale = 'pt-BR'): string
{
    if ($locale === 'en-US') {
        return match ($stake) {
            3 => 'Truco',
            6 => 'Six',
            9 => 'Nine',
            12 => 'Twelve',
            default => (string) $stake,
        };
    }
    return match ($stake) {
        3 => 'Truco',
        6 => 'Seis',
        9 => 'Nove',
        12 => 'Doze',
        default => (string) $stake,
    };
}

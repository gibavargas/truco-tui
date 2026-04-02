<?php
require_once __DIR__ . '/partials/card.php';

$bundle = $_SESSION['runtime_bundle'] ?? [];
$events = $_SESSION['runtime_events'] ?? [];
$players = $snap['Players'] ?? [];
$logs = $snap['Logs'] ?? [];
$myID = $snap['CurrentPlayerIdx'] ?? 0;
$turnPlayer = $snap['TurnPlayer'] ?? -1;
$pendingFor = $snap['PendingRaiseFor'] ?? -1;
$matchFinished = $snap['MatchFinished'] ?? false;
$winnerTeam = $snap['WinnerTeam'] ?? -1;
$hand = $snap['CurrentHand'] ?? [];
$mode = $bundle['mode'] ?? 'offline_match';
$isOnline = strpos($mode, 'host_') === 0 || strpos($mode, 'client_') === 0;
$lobby = $bundle['lobby'] ?? [];
$ui = $bundle['ui'] ?? [];
$connection = $bundle['connection'] ?? [];
$diagnostics = $bundle['diagnostics'] ?? [];
$network = $connection['network'] ?? [];
$slotStates = $ui['lobby_slots'] ?? [];
$actions = $ui['actions'] ?? [];
$locale = $_SESSION['locale'] ?? 'pt-BR';
$runtimeStateValid = (bool) ($_SESSION['runtime_state_valid'] ?? true);
$connectionStatus = (string) ($connection['status'] ?? $mode);
$roleLabel = trim((string) ($lobby['role'] ?? ''));
$transportLabel = (string) ($network['transport'] ?? '-');
$protocolLabel = '-';
if (!empty($network['negotiated_protocol_version'])) {
    $protocolLabel = 'v' . (int) $network['negotiated_protocol_version'];
} elseif (!empty($network['seat_protocol_versions']) && is_array($network['seat_protocol_versions'])) {
    $versions = array_values(array_unique(array_map('intval', $network['seat_protocol_versions'])));
    $versions = array_values(array_filter($versions, static fn($version) => $version > 0));
    sort($versions);
    if (count($versions) > 1) {
        $protocolLabel = tr('connection_protocol_mixed') . ' (' . implode(', ', array_map(static fn($version) => 'v' . $version, $versions)) . ')';
    } elseif (count($versions) === 1) {
        $protocolLabel = 'v' . $versions[0];
    }
}
$eventFeedLines = [];
foreach (array_reverse(array_slice($events, -12)) as $ev) {
    $eventFeedLines[] = '[' . substr((string) ($ev['timestamp'] ?? ''), 11, 8) . '] ' . formatEventLine($ev);
}

$playersByID = [];
$seatByID = [];
$numPlayers = $snap['NumPlayers'] ?? count($players);
foreach ($players as $p) {
    $id = (int) ($p['ID'] ?? 0);
    $playersByID[$id] = $p;
    $seatByID[$id] = ($numPlayers === 2)
        ? ($id === $myID ? 'bottom' : 'top')
        : (['bottom', 'right', 'top', 'left'][($id - $myID + 4) % 4] ?? 'top');
}

$myPlayer = $playersByID[$myID] ?? null;
$myCards = $myPlayer['Hand'] ?? [];
$myTeam = (int) ($myPlayer['Team'] ?? 0);
$stake = (int) ($hand['Stake'] ?? 1);
$pendingTo = (int) ($snap['PendingRaiseTo'] ?? 0);
$turnPlayerObj = $playersByID[$turnPlayer] ?? null;
$turnPlayerName = (string) ($turnPlayerObj['Name'] ?? '?');
$raiseRequester = (int) ($hand['RaiseRequester'] ?? -1);
$raiseRequesterName = (string) (($playersByID[$raiseRequester]['Name'] ?? '') ?: $turnPlayerName);
$roundCards = $hand['RoundCards'] ?? [];
$lastTrickCards = $snap['LastTrickCards'] ?? [];
$trickPiles = $snap['TrickPiles'] ?? [];
$trickWins = $hand['TrickWins'] ?? [];
$team1Tricks = (int) ($trickWins[0] ?? $trickWins['0'] ?? 0);
$team2Tricks = (int) ($trickWins[1] ?? $trickWins['1'] ?? 0);
$roundNumber = max(1, min(3, (int) ($hand['Round'] ?? 1)));
$lastTrickRound = (int) ($snap['LastTrickRound'] ?? 0);
$lastTrickTie = (bool) ($snap['LastTrickTie'] ?? false);
$lastTrickWinnerID = (int) ($snap['LastTrickWinner'] ?? -1);
$lastTrickWinnerName = (string) (($playersByID[$lastTrickWinnerID]['Name'] ?? '') ?: '?');
$lastTrickSeq = (int) ($snap['LastTrickSeq'] ?? 0);
$latestLog = trim((string) (count($logs) > 0 ? $logs[count($logs) - 1] : tr('status_ready')));
$recentLogs = array_reverse(array_slice($logs, -3));
$recentEvents = [];
foreach (array_reverse(array_slice($events, -2)) as $ev) {
    $recentEvents[] = formatEventLine($ev);
}
$matchPoints = $snap['MatchPoints'] ?? [];
$team1Score = (int) ($matchPoints[0] ?? $matchPoints['0'] ?? 0);
$team2Score = (int) ($matchPoints[1] ?? $matchPoints['1'] ?? 0);
$lastPlayedPlayerID = -1;
if (!empty($roundCards)) {
    $lastPlayed = $roundCards[count($roundCards) - 1];
    $lastPlayedPlayerID = (int) ($lastPlayed['PlayerID'] ?? -1);
}

$canPlayCard = (bool) ($actions['can_play_card'] ?? false);
$canTruco = (bool) ($actions['can_ask_or_raise'] ?? false);
$canAccept = (bool) ($actions['can_accept'] ?? false);
$canRefuse = (bool) ($actions['can_refuse'] ?? false);

if (!$runtimeStateValid) {
    $canPlayCard = false;
    $canTruco = false;
    $canAccept = false;
    $canRefuse = false;
}

$stakeSteps = [1, 3, 6, 9, 12];
$nextStake = $stake;
$nextRaisePreview = $pendingTo ?: $stake;
foreach ($stakeSteps as $idx => $step) {
    if ($step === $stake && isset($stakeSteps[$idx + 1])) {
        $nextStake = $stakeSteps[$idx + 1];
    }
    if ($pendingTo > 0 && $step === $pendingTo && isset($stakeSteps[$idx + 1])) {
        $nextRaisePreview = $stakeSteps[$idx + 1];
    }
}

if ($matchFinished) {
    $statusText = tr('status_match_end');
} elseif ($pendingFor !== -1 && $pendingFor === $myTeam) {
    $statusText = tr('status_pending_you', raiseLabel($pendingTo ?: 3, $locale), $pendingTo ?: 3);
} elseif ($pendingFor !== -1) {
    $statusText = tr('status_pending_other', $turnPlayerName, raiseLabel($pendingTo ?: 3, $locale), $pendingTo ?: 3);
} elseif ($turnPlayer === $myID) {
    $statusText = tr('status_your_turn');
} else {
    $statusText = tr('status_wait_cpu', $turnPlayerName);
}

if ($pendingFor !== -1 && $pendingFor === $myTeam) {
    $heroText = strtoupper(raiseLabel($pendingTo ?: 3, $locale));
    $heroSubtext = $raiseRequester === $myID
        ? tr('game_response_called_by_you', raiseLabel($nextRaisePreview, $locale))
        : tr('game_response_called_by', $raiseRequesterName, raiseLabel($nextRaisePreview, $locale));
} elseif ($pendingFor !== -1) {
    $heroText = strtoupper(raiseLabel($pendingTo ?: 3, $locale));
    $heroSubtext = $raiseRequester === $myID
        ? tr('game_response_called_by_you', raiseLabel($nextRaisePreview, $locale))
        : tr('game_response_called_by', $raiseRequesterName, raiseLabel($nextRaisePreview, $locale));
} elseif ($turnPlayer === $myID) {
    $heroText = tr('game_hand_ready');
    $heroSubtext = tr('game_center_turn_you');
} else {
    $heroText = tr('game_hand_wait');
    $heroSubtext = tr('game_center_turn_other', $turnPlayerName);
}

if (!$runtimeStateValid) {
    $statusText = tr('game_runtime_stale_copy');
    $heroText = tr('game_runtime_stale_title');
    $heroSubtext = tr('game_runtime_stale_copy');
}

$lastTrickText = '';
if ($lastTrickRound > 0) {
    $lastTrickText = $lastTrickTie
        ? tr('game_last_trick_tie', $lastTrickRound)
        : tr('game_last_trick_win', $lastTrickWinnerName, $lastTrickRound);
}

$viraLabel = '-';
if (!empty($hand['Vira'])) {
    $viraLabel = (string) ($hand['Vira']['Rank'] ?? '?') . suitSymbol((string) ($hand['Vira']['Suit'] ?? ''));
}
$manilhaLabel = htmlspecialchars((string) ($hand['Manilha'] ?? '-'));

$showTurnActions = $runtimeStateValid && !$matchFinished && $turnPlayer === $myID && !$canAccept;
$showResponseActions = $runtimeStateValid && !$matchFinished && $canAccept;
$showLastTrickMonte = !empty($lastTrickCards) && $lastTrickWinnerID >= 0;
$boardState = 'waiting-turn';
if ($matchFinished) {
    $boardState = 'round-end';
} elseif ($showResponseActions) {
    $boardState = 'truco-pending';
} elseif ($showTurnActions) {
    $boardState = 'player-turn';
} elseif ($lastTrickRound > 0 && empty($roundCards)) {
    $boardState = 'trick-resolved';
}
if (!$runtimeStateValid) {
    $boardState .= ' state-runtime-stale';
}

?>
<section class="panel game-panel game-panel-board game-shell state-<?= htmlspecialchars($boardState) ?>"
    data-mode="<?= htmlspecialchars($mode) ?>"
    data-match-finished="<?= $matchFinished ? '1' : '0' ?>"
    data-last-trick-seq="<?= htmlspecialchars($lastTrickSeq) ?>"
    data-last-trick-team="<?= htmlspecialchars((int) ($snap['LastTrickTeam'] ?? -1)) ?>"
    data-last-trick-winner="<?= htmlspecialchars($lastTrickWinnerID) ?>"
    data-last-trick-tie="<?= $lastTrickTie ? '1' : '0' ?>"
    data-last-trick-round="<?= htmlspecialchars($lastTrickRound) ?>"
    data-last-trick-round-label="<?= htmlspecialchars(tr('game_trick_round_label', max(1, $lastTrickRound))) ?>"
    data-local-player-id="<?= htmlspecialchars($myID) ?>"
    data-num-players="<?= htmlspecialchars($numPlayers) ?>"
    data-trick-toast-win="<?= htmlspecialchars(tr('game_trick_toast_win')) ?>"
    data-trick-toast-loss="<?= htmlspecialchars(tr('game_trick_toast_loss')) ?>"
    data-trick-toast-tie="<?= htmlspecialchars(tr('game_trick_toast_tie')) ?>"
    data-trick-toast-caption="<?= htmlspecialchars(tr('game_trick_toast_caption')) ?>"
>
    <div class="game-topline">
        <div class="game-head-copy">
            <span class="section-kicker"><?= $isOnline ? tr('game_kicker_online') : tr('game_kicker_offline') ?></span>
            <h2><?= $isOnline ? tr('game_title_online') : tr('game_title_offline') ?></h2>
            <p class="game-head-note"><?= htmlspecialchars($locale === 'en-US' ? 'Hand, flip, truco, and score in one frame.' : 'Mão, vira, truco e placar no mesmo quadro.') ?></p>
        </div>

        <div class="game-topline-actions">
            <div class="ui-mode-toggle" role="group" aria-label="<?= htmlspecialchars(tr('ui_mode_label')) ?>">
                <button type="button" class="btn btn-neutral btn-mini ui-mode-btn" data-ui-mode="wireframe"><?= tr('ui_mode_wireframe') ?></button>
                <button type="button" class="btn btn-neutral btn-mini ui-mode-btn" data-ui-mode="polished"><?= tr('ui_mode_polished') ?></button>
            </div>
            <?php if ($isOnline): ?>
                <form method="post" action="index.php" data-ajax="true">
                    <input type="hidden" name="action" value="leaveLobby">
                    <button type="submit" class="btn btn-neutral btn-mini"><?= tr('lobby_leave') ?></button>
                </form>
            <?php endif; ?>
        </div>
    </div>

    <div class="board-stage players-<?= $numPlayers ?> board-zone board-zone-table">
        <div class="board-hud board-zone board-zone-hud">
            <section class="hud-team <?= $myTeam === 0 ? 'friendly' : 'enemy' ?>">
                <div class="hud-team-top">
                    <span class="hud-label"><?= tr('team1') ?></span>
                    <span class="team-link <?= $myTeam === 0 ? 'ally' : 'enemy' ?>"><?= $myTeam === 0 ? tr('game_you_label') . ' + ' . tr('game_partner_label') : tr('game_opponent_label') ?></span>
                </div>
                <strong class="hud-score"><?= $team1Score ?></strong>
                <div class="hud-tricks">
                    <?php for ($i = 0; $i < 3; $i++): ?>
                        <span class="hud-trick <?= $i < $team1Tricks ? 'won' : '' ?>"></span>
                    <?php endfor; ?>
                </div>
            </section>

            <section class="hud-center">
                <div class="hud-round-row">
                    <span class="hud-chip"><?= tr('game_round_label', $roundNumber) ?></span>
                    <span class="hud-chip stake-chip"><?= tr('game_stake_value', $stake) ?></span>
                    <?php if ($pendingFor !== -1): ?>
                        <span class="hud-chip hot"><?= strtoupper(raiseLabel($pendingTo ?: 3, $locale)) ?></span>
                    <?php endif; ?>
                </div>

                <div class="hud-stake-track" aria-label="<?= htmlspecialchars(tr('game_stake_ladder_label')) ?>">
                    <?php foreach ($stakeSteps as $step): ?>
                        <span class="track-step <?= $step === $stake ? 'current' : ($step < $stake ? 'past' : ($step === $pendingTo ? 'pending' : '')) ?>">
                            <span class="track-dot"></span>
                            <small><?= $step ?></small>
                        </span>
                    <?php endforeach; ?>
                </div>

                <div class="hud-status-block">
                    <strong class="hud-status-title"><?= htmlspecialchars($heroText) ?></strong>
                    <span class="hud-status"><?= htmlspecialchars($statusText) ?></span>
                </div>
            </section>

            <section class="hud-team <?= $myTeam === 1 ? 'friendly' : 'enemy' ?>">
                <div class="hud-team-top">
                    <span class="hud-label"><?= tr('team2') ?></span>
                    <span class="team-link <?= $myTeam === 1 ? 'ally' : 'enemy' ?>"><?= $myTeam === 1 ? tr('game_you_label') . ' + ' . tr('game_partner_label') : tr('game_opponent_label') ?></span>
                </div>
                <strong class="hud-score"><?= $team2Score ?></strong>
                <div class="hud-tricks">
                    <?php for ($i = 0; $i < 3; $i++): ?>
                        <span class="hud-trick <?= $i < $team2Tricks ? 'won' : '' ?>"></span>
                    <?php endfor; ?>
                </div>
            </section>
        </div>

        <aside class="table-legend">
            <div class="legend-grid legend-grid-table">
                <div class="legend-card deck-card">
                    <span class="info-kicker"><?= tr('game_deck_label') ?></span>
                    <div class="meta-card meta-card-back"><?= renderCardBack(true) ?></div>
                    <strong><?= tr('game_round_label', $roundNumber) ?></strong>
                </div>
                <div class="legend-card vira-card">
                    <span class="info-kicker"><?= tr('vira') ?></span>
                    <div class="meta-card"><?= !empty($hand['Vira']) ? renderCard($hand['Vira'], true) : '' ?></div>
                    <strong><?= htmlspecialchars($viraLabel) ?></strong>
                    <span class="meta-sub"><?= tr('manilha') ?> <?= $manilhaLabel ?></span>
                </div>
            </div>
        </aside>

        <div class="table-callout <?= $showResponseActions ? 'hot' : '' ?>" data-focus-target="board-callout" tabindex="-1" role="status" aria-live="polite" aria-atomic="true">
            <span class="section-kicker"><?= tr('status_title') ?></span>
            <strong><?= htmlspecialchars($heroText) ?></strong>
            <span><?= htmlspecialchars($heroSubtext) ?></span>
        </div>

        <?php foreach ($players as $p): ?>
            <?php
            $playerID = (int) ($p['ID'] ?? -1);
            $pos = $seatByID[$playerID] ?? 'top';
            $team = (int) ($p['Team'] ?? 0);
            $teamNum = $team + 1;
            $cardCount = count($p['Hand'] ?? []);
            $isSelf = $playerID === $myID;
            $isTurn = $playerID === $turnPlayer;
            $isPartner = !$isSelf && $team === $myTeam;
            $seatRole = $isSelf ? tr('game_you_label') : ($isPartner ? tr('game_partner_label') : tr('game_opponent_label'));
            $playerTrickPiles = array_values(array_filter($trickPiles, fn($pile) => (int) ($pile['Winner'] ?? -1) === $playerID));
            ?>
            <div class="board-seat board-seat-<?= $pos ?> <?= $isTurn ? 'is-turn' : '' ?> <?= $isSelf ? 'is-self' : ($isPartner ? 'is-partner' : 'is-opponent') ?>"
                data-player-id="<?= htmlspecialchars($playerID) ?>"
                data-team="<?= htmlspecialchars($team) ?>">
                <div class="board-seat-badge team-<?= $teamNum ?>"><?= htmlspecialchars(strtoupper(substr((string) ($p['Name'] ?? '?'), 0, 1))) ?></div>
                <div class="board-seat-info">
                    <div class="seat-headline">
                        <strong><?= htmlspecialchars((string) ($p['Name'] ?? '?')) ?></strong>
                        <?php if ($isTurn): ?>
                            <span class="seat-turn-pill"><?= tr('game_turn_badge') ?></span>
                        <?php endif; ?>
                    </div>
                    <span><?= $seatRole ?><?= !empty($p['CPU']) ? ' · ' . tr('cpu_tag') : '' ?></span>
                    <?php if (!$isSelf): ?>
                        <div class="board-seat-cards" aria-hidden="true">
                            <?php for ($i = 0; $i < $cardCount; $i++): ?>
                                <span class="tiny-back"></span>
                            <?php endfor; ?>
                        </div>
                    <?php endif; ?>
                    <?php if (!empty($playerTrickPiles)): ?>
                        <div class="board-seat-monte" aria-label="<?= htmlspecialchars(tr('game_monte_label')) ?>">
                            <span class="board-seat-monte-label"><?= tr('game_monte_label') ?></span>
                            <div class="board-seat-monte-stack board-seat-monte-stack-multi" aria-hidden="true">
                                <?php foreach ($playerTrickPiles as $pileIndex => $pile): ?>
                                    <?php $pileCards = $pile['Cards'] ?? []; ?>
                                    <div class="board-seat-monte-pile">
                                        <span class="board-seat-monte-round"><?= htmlspecialchars(tr('game_trick_round_label', (int) ($pile['Round'] ?? ($pileIndex + 1)))) ?></span>
                                        <div class="board-seat-monte-pile-stack">
                                            <?php foreach (array_slice($pileCards, 0, 4) as $idx => $_card): ?>
                                                <span class="tiny-back board-seat-monte-back board-seat-monte-back-<?= $idx ?>"></span>
                                            <?php endforeach; ?>
                                        </div>
                                    </div>
                                <?php endforeach; ?>
                            </div>
                        </div>
                    <?php endif; ?>
                </div>
            </div>
        <?php endforeach; ?>

        <div class="board-center">
            <div class="board-tricks">
                <?php for ($i = 0; $i < 3; $i++): ?>
                    <?php
                    $trickResult = $hand['TrickResults'][$i] ?? null;
                    $trickClass = '';
                    if ($trickResult === -1) {
                        $trickClass = 'tie';
                    } elseif ($trickResult === 0 || $trickResult === 1) {
                        $trickClass = ((int) $trickResult === $myTeam) ? 'ally' : 'enemy';
                    }
                    ?>
                    <span class="board-trick <?= $trickClass ?>"><?= tr('trick_short') . ($i + 1) ?></span>
                <?php endfor; ?>
            </div>

            <div class="board-cards <?= empty($roundCards) ? 'empty' : '' ?>">
                <?php if (empty($roundCards)): ?>
                    <div class="board-center-note"><?= htmlspecialchars($statusText) ?></div>
                <?php endif; ?>

                <?php foreach ($roundCards as $index => $pc): ?>
                    <?php
                    $ownerID = (int) ($pc['PlayerID'] ?? -1);
                    $ownerName = (string) (($playersByID[$ownerID]['Name'] ?? '') ?: ('P' . ($ownerID + 1)));
                    $cardPos = $seatByID[$ownerID] ?? 'top';
                    $ownerTeam = (int) (($playersByID[$ownerID]['Team'] ?? 0));
                    $ownerClass = $ownerTeam === $myTeam ? 'ally' : 'enemy';
                    $isFaceDown = !empty($pc['FaceDown']);
                    if ($ownerID === $myID) {
                        $ownerClass = 'self';
                    }
                    ?>
                    <div class="board-played board-played-<?= $cardPos ?> <?= $index === count($roundCards) - 1 ? 'is-last' : '' ?> <?= $lastPlayedPlayerID === $ownerID ? 'is-current' : '' ?> <?= $ownerClass ?>">
                        <span class="board-played-owner"><?= htmlspecialchars($ownerName) ?></span>
                        <?= $isFaceDown ? renderCardBack(false) : renderCard($pc['Card']) ?>
                    </div>
                <?php endforeach; ?>
            </div>

            <?php if ($lastTrickText !== ''): ?>
                <div class="board-resolution <?= $lastTrickTie ? 'tie' : '' ?>">
                    <span class="section-kicker"><?= tr('log_title') ?></span>
                    <strong><?= htmlspecialchars($lastTrickText) ?></strong>
                </div>
            <?php endif; ?>
        </div>

        <div class="trick-overlay" data-trick-overlay role="status" aria-live="polite" aria-atomic="true" aria-hidden="true">
            <div class="trick-overlay-deck" data-trick-overlay-deck>
                <div class="trick-overlay-card"><?= renderCardBack(true) ?></div>
                <div class="trick-overlay-card"><?= renderCardBack(true) ?></div>
                <div class="trick-overlay-card"><?= renderCardBack(true) ?></div>
            </div>
            <div class="trick-overlay-toast" data-trick-overlay-toast>
                <span class="trick-overlay-kicker" data-trick-overlay-kicker></span>
                <strong data-trick-overlay-title></strong>
                <span data-trick-overlay-caption><?= htmlspecialchars(tr('game_trick_toast_caption')) ?></span>
            </div>
        </div>
    </div>

    <div class="player-dock board-zone board-zone-hand">
        <section class="player-hand" data-focus-target="player-hand" tabindex="-1">
            <div class="player-dock-head">
                <div>
                    <span class="section-kicker"><?= tr('hand_title') ?></span>
                    <h3><?= tr('game_hand_heading') ?></h3>
                </div>
                <div class="player-hand-meta">
                    <span class="player-dock-status <?= $canPlayCard ? 'ready' : '' ?>">
                        <?= $canPlayCard ? tr('game_hand_ready') : tr('game_hand_wait') ?>
                    </span>
                    <span class="hand-count"><?= htmlspecialchars(tr('game_hand_count', count($myCards))) ?></span>
                </div>
            </div>

            <div class="my-hand board-hand" role="list">
                <?php foreach ($myCards as $idx => $card): ?>
                    <?php $cardActionLabel = tr('card_action_play', cardLabel($card)); ?>
                    <?php if ($canPlayCard): ?>
                        <form method="post" action="index.php" class="card-form" data-ajax="true">
                            <input type="hidden" name="action" value="play">
                            <input type="hidden" name="cardIndex" value="<?= htmlspecialchars($idx) ?>">
                            <button type="submit" class="card-btn" role="listitem" aria-label="<?= htmlspecialchars($cardActionLabel) ?>" title="<?= htmlspecialchars($cardActionLabel) ?>"><?= renderCard($card, false, (string) ($idx + 1)) ?></button>
                            <?php if (($hand['Round'] ?? 1) >= 2): ?>
                                <?php $faceDownLabel = tr('card_action_play_face_down', cardLabel($card)); ?>
                                <button type="submit" name="faceDown" value="1" class="btn btn-neutral btn-face-down" aria-label="<?= htmlspecialchars($faceDownLabel) ?>" title="<?= htmlspecialchars($faceDownLabel) ?>">Virada</button>
                            <?php endif; ?>
                        </form>
                    <?php else: ?>
                        <div class="card-btn disabled-card" role="listitem" aria-label="<?= htmlspecialchars(cardLabel($card)) ?>" title="<?= htmlspecialchars(cardLabel($card)) ?>"><?= renderCard($card, false, (string) ($idx + 1)) ?></div>
                    <?php endif; ?>
                <?php endforeach; ?>
            </div>
        </section>

        <section class="player-actions" data-focus-target="player-actions" tabindex="-1">
            <?php if ($showResponseActions): ?>
                <div class="player-actions-row response-row">
                    <div class="player-actions-copy">
                        <span class="section-kicker"><?= tr('game_action_title_response') ?></span>
                        <strong><?= htmlspecialchars($statusText) ?></strong>
                        <span><?= htmlspecialchars($heroSubtext) ?></span>
                    </div>
                    <form method="post" action="index.php" data-ajax="true" class="action-form action-form-truco">
                        <input type="hidden" name="action" value="truco">
                        <button type="submit" class="action-btn action-btn-truco <?= $canTruco ? 'armed' : '' ?>" <?= $canTruco ? '' : 'disabled' ?>>
                            <strong><?= strtoupper(tr('btn_raise')) ?></strong>
                            <span class="action-sub"><?= htmlspecialchars(tr('game_action_raise_sub', raiseLabel($nextRaisePreview, $locale))) ?></span>
                        </button>
                    </form>
                    <form method="post" action="index.php" data-ajax="true" class="action-form">
                        <input type="hidden" name="action" value="accept">
                        <button type="submit" class="action-btn action-btn-accept" <?= $canAccept ? '' : 'disabled' ?>>
                            <strong><?= strtoupper(tr('btn_accept')) ?></strong>
                            <span class="action-sub"><?= htmlspecialchars(tr('game_action_accept_sub', $pendingTo ?: 3)) ?></span>
                        </button>
                    </form>
                    <form method="post" action="index.php" data-ajax="true" class="action-form">
                        <input type="hidden" name="action" value="refuse">
                        <button type="submit" class="action-btn action-btn-refuse" <?= $canRefuse ? '' : 'disabled' ?>>
                            <strong><?= strtoupper(tr('btn_refuse')) ?></strong>
                            <span class="action-sub"><?= tr('game_action_refuse_sub') ?></span>
                        </button>
                    </form>
                </div>
            <?php elseif ($showTurnActions): ?>
                <div class="player-actions-row turn-row">
                    <div class="player-actions-copy">
                        <span class="section-kicker"><?= tr('game_action_title_turn') ?></span>
                        <strong><?= htmlspecialchars($statusText) ?></strong>
                        <span><?= tr('game_action_turn_copy') ?></span>
                    </div>
                    <form method="post" action="index.php" data-ajax="true" class="action-form action-form-truco">
                        <input type="hidden" name="action" value="truco">
                        <button type="submit" class="action-btn action-btn-truco <?= $canTruco ? 'armed' : '' ?>" <?= $canTruco ? '' : 'disabled' ?>>
                            <strong><?= strtoupper(tr('btn_truco')) ?></strong>
                            <span class="action-sub"><?= htmlspecialchars(tr('game_action_call_sub', raiseLabel($nextStake, $locale))) ?></span>
                        </button>
                    </form>
                </div>
            <?php else: ?>
                <div class="player-actions-row waiting-row">
                    <div class="player-actions-copy">
                        <span class="section-kicker"><?= tr('game_action_title_wait') ?></span>
                        <strong><?= htmlspecialchars($statusText) ?></strong>
                        <span><?= tr('game_action_wait_copy') ?></span>
                    </div>
                </div>
            <?php endif; ?>
        </section>
    </div>

    <div class="board-footer board-zone board-zone-footer" data-focus-target="board-footer" tabindex="-1">
        <div class="board-logline">
            <span class="section-kicker"><?= tr('log_title') ?></span>
            <strong><?= htmlspecialchars($latestLog) ?></strong>
            <?php if (!empty($recentLogs)): ?>
                <div class="board-log-tape">
                    <?php foreach ($recentLogs as $line): ?>
                        <span class="hud-chip muted"><?= htmlspecialchars((string) $line) ?></span>
                    <?php endforeach; ?>
                </div>
            <?php endif; ?>
        </div>

        <?php if (!empty($recentEvents)): ?>
            <div class="board-events">
                <?php foreach ($recentEvents as $line): ?>
                    <span class="hud-chip muted"><?= htmlspecialchars($line) ?></span>
                <?php endforeach; ?>
            </div>
        <?php endif; ?>
    </div>

    <?php if ($isOnline): ?>
        <details class="table-sidecar online-emphasis" open>
            <summary><?= tr('game_table_online') ?></summary>

            <div class="table-ops sidecar-grid">
                <section class="side-block side-block-diagnostics">
                    <h3><?= tr('game_online_panel_title') ?></h3>
                    <div class="connection-grid">
                        <div><span><?= tr('connection_status') ?></span><strong><?= htmlspecialchars($connectionStatus) ?></strong></div>
                        <div><span><?= tr('connection_mode') ?></span><strong><?= !empty($connection['is_online']) ? tr('connection_online') : tr('connection_offline') ?></strong></div>
                        <div><span><?= tr('connection_transport') ?></span><strong><?= htmlspecialchars($transportLabel) ?></strong></div>
                        <div><span><?= tr('connection_protocol') ?></span><strong><?= htmlspecialchars($protocolLabel) ?></strong></div>
                        <?php if ($roleLabel !== ''): ?>
                            <div><span><?= tr('connection_role') ?></span><strong><?= htmlspecialchars($roleLabel) ?></strong></div>
                        <?php endif; ?>
                        <div><span><?= tr('connection_backlog') ?></span><strong><?= (int) ($diagnostics['event_backlog'] ?? 0) ?></strong></div>
                        <?php if (!empty($connection['last_error']['message'])): ?>
                            <div class="connection-error"><span><?= tr('connection_error') ?></span><strong><?= htmlspecialchars((string) $connection['last_error']['message']) ?></strong></div>
                        <?php endif; ?>
                    </div>
                </section>

                <section class="side-block side-block-pulse side-log">
                    <h3><?= tr('game_online_pulse_title') ?></h3>
                    <pre id="game-events" class="event-feed" data-focus-target="game-events" tabindex="-1" role="log" aria-live="polite" aria-atomic="false"><?php
                        if (empty($eventFeedLines)) {
                            echo htmlspecialchars(tr('lobby_events_empty'));
                        } else {
                            echo htmlspecialchars(implode("\n", $eventFeedLines));
                        }
                    ?></pre>
                </section>

                <section class="side-block side-block-controls">
                    <h3><?= tr('game_online_controls_title') ?></h3>
                    <div class="action-row compact-row">
                        <?php foreach (($lobby['slots'] ?? []) as $idx => $slotName): ?>
                            <?php $slotState = $slotStates[$idx] ?? []; ?>
                            <?php if (!empty($slotState['can_vote_host'])): ?>
                                <form method="post" action="index.php" data-ajax="true">
                                    <input type="hidden" name="action" value="voteHost">
                                    <input type="hidden" name="slot" value="<?= htmlspecialchars($idx) ?>">
                                    <button type="submit" class="btn btn-neutral"><?= tr('action_vote_host') ?> <?= $idx + 1 ?></button>
                                </form>
                            <?php endif; ?>
                            <?php if (!empty($slotState['can_request_replacement'])): ?>
                                <form method="post" action="index.php" data-ajax="true">
                                    <input type="hidden" name="action" value="requestReplacementInvite">
                                    <input type="hidden" name="slot" value="<?= htmlspecialchars($idx) ?>">
                                    <button type="submit" class="btn btn-truco"><?= tr('action_replacement_invite') ?> <?= $idx + 1 ?></button>
                                </form>
                            <?php endif; ?>
                        <?php endforeach; ?>
                    </div>
                    <form method="post" action="index.php" class="lobby-chat-row" data-ajax="true">
                        <input type="hidden" name="action" value="sendChat">
                        <input name="message" class="field" type="text" autocomplete="off" placeholder="<?= htmlspecialchars(tr('chat_placeholder')) ?>" required>
                        <button type="submit" class="btn btn-neutral"><?= tr('lobby_chat_send') ?></button>
                    </form>
                </section>
            </div>
        </details>
    <?php endif; ?>

    <?php if ($matchFinished): ?>
        <div class="overlay" style="display:flex">
            <div class="overlay-card match-card">
                <h2><?= $winnerTeam === $myTeam ? tr('overlay_match_win') : tr('overlay_match_loss') ?></h2>
                <p><?= tr('overlay_match_detail', $team1Score, $team2Score) ?></p>
                <form method="post" action="index.php" data-ajax="true">
                    <input type="hidden" name="action" value="reset">
                    <button type="submit" class="btn btn-primary"><?= tr('btn_play_again') ?></button>
                </form>
            </div>
        </div>
    <?php endif; ?>
</section>

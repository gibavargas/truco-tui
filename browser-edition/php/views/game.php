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
$slotStates = $ui['lobby_slots'] ?? [];
$actions = $ui['actions'] ?? [];
$locale = $_SESSION['locale'] ?? 'pt-BR';

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
$trickWins = $hand['TrickWins'] ?? [];
$team1Tricks = (int) ($trickWins[0] ?? $trickWins['0'] ?? 0);
$team2Tricks = (int) ($trickWins[1] ?? $trickWins['1'] ?? 0);
$roundNumber = max(1, min(3, (int) ($hand['Round'] ?? 1)));
$lastTrickRound = (int) ($snap['LastTrickRound'] ?? 0);
$lastTrickTie = (bool) ($snap['LastTrickTie'] ?? false);
$lastTrickWinnerID = (int) ($snap['LastTrickWinner'] ?? -1);
$lastTrickWinnerName = (string) (($playersByID[$lastTrickWinnerID]['Name'] ?? '') ?: '?');
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

$canPlayCard = (bool) ($actions['can_play_card'] ?? (!$matchFinished && $turnPlayer === $myID && $pendingFor === -1));
$canTruco = (bool) ($actions['can_ask_or_raise'] ?? (!$matchFinished && (($turnPlayer === $myID && $pendingFor === -1) || ($pendingFor !== -1 && $pendingFor === $myTeam))));
$canAccept = (bool) ($actions['can_accept'] ?? (!$matchFinished && $pendingFor !== -1 && $pendingFor === $myTeam));
$canRefuse = (bool) ($actions['can_refuse'] ?? $canAccept);

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
    $heroSubtext = tr('game_center_pending_you', raiseLabel($pendingTo ?: 3, $locale));
} elseif ($pendingFor !== -1) {
    $heroText = strtoupper(raiseLabel($pendingTo ?: 3, $locale));
    $heroSubtext = tr('game_center_pending_other', $raiseRequesterName, raiseLabel($pendingTo ?: 3, $locale));
} elseif ($turnPlayer === $myID) {
    $heroText = tr('game_hand_ready');
    $heroSubtext = tr('game_center_turn_you');
} else {
    $heroText = tr('game_hand_wait');
    $heroSubtext = tr('game_center_turn_other', $turnPlayerName);
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

$showTurnActions = !$matchFinished && $turnPlayer === $myID && !$canAccept;
$showResponseActions = !$matchFinished && $canAccept;
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

$hierarchyText = $locale === 'en-US'
    ? '1. Your hand. 2. Current trick. 3. Turn and truco pressure. 4. Team score.'
    : '1. Sua mão. 2. Vaza atual. 3. Turno e pressão do truco. 4. Placar da dupla.';
?>
<section class="panel game-panel game-panel-board game-shell state-<?= htmlspecialchars($boardState) ?>">
    <div class="game-topline">
        <div class="game-head-copy">
            <span class="section-kicker"><?= $isOnline ? tr('game_kicker_online') : tr('game_kicker_offline') ?></span>
            <h2><?= $isOnline ? tr('game_title_online') : tr('game_title_offline') ?></h2>
            <p class="game-head-note"><?= htmlspecialchars($hierarchyText) ?></p>
        </div>

        <div class="game-topline-actions">
            <div class="ui-mode-toggle" role="group" aria-label="UI mode">
                <button type="button" class="btn btn-neutral btn-mini ui-mode-btn" data-ui-mode="wireframe">Esquema</button>
                <button type="button" class="btn btn-neutral btn-mini ui-mode-btn" data-ui-mode="polished">Mesa</button>
            </div>
            <?php if ($isOnline): ?>
                <form method="post" action="index.php" data-ajax="true">
                    <input type="hidden" name="action" value="leaveLobby">
                    <button type="submit" class="btn btn-neutral btn-mini"><?= tr('lobby_leave') ?></button>
                </form>
            <?php endif; ?>
        </div>
    </div>

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
                <span class="hud-chip stake-chip">Vale <?= $stake ?></span>
                <?php if ($pendingFor !== -1): ?>
                    <span class="hud-chip hot"><?= strtoupper(raiseLabel($pendingTo ?: 3, $locale)) ?></span>
                <?php endif; ?>
            </div>

            <div class="hud-stake-track" aria-label="Truco escalation">
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

    <div class="board-stage players-<?= $numPlayers ?> board-zone board-zone-table">
        <aside class="table-legend">
            <div class="legend-chip legend-chip-primary"><?= htmlspecialchars($hierarchyText) ?></div>
            <div class="legend-grid">
                <div class="legend-card">
                    <span class="info-kicker"><?= tr('vira') ?></span>
                    <div class="meta-card"><?= !empty($hand['Vira']) ? renderCard($hand['Vira'], true) : '' ?></div>
                    <strong><?= htmlspecialchars($viraLabel) ?></strong>
                </div>
                <div class="legend-card">
                    <span class="info-kicker"><?= tr('manilha') ?></span>
                    <strong class="meta-rank"><?= $manilhaLabel ?></strong>
                </div>
            </div>
        </aside>

        <div class="table-callout <?= $showResponseActions ? 'hot' : '' ?>">
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
            ?>
            <div class="board-seat board-seat-<?= $pos ?> <?= $isTurn ? 'is-turn' : '' ?> <?= $isSelf ? 'is-self' : ($isPartner ? 'is-partner' : 'is-opponent') ?>">
                <div class="board-seat-badge team-<?= $teamNum ?>"><?= htmlspecialchars(strtoupper(substr((string) ($p['Name'] ?? '?'), 0, 1))) ?></div>
                <div class="board-seat-info">
                    <div class="seat-headline">
                        <strong><?= htmlspecialchars((string) ($p['Name'] ?? '?')) ?></strong>
                        <?php if ($isTurn): ?>
                            <span class="seat-turn-pill">VEZ</span>
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
                    if ($ownerID === $myID) {
                        $ownerClass = 'self';
                    }
                    ?>
                    <div class="board-played board-played-<?= $cardPos ?> <?= $index === count($roundCards) - 1 ? 'is-last' : '' ?> <?= $lastPlayedPlayerID === $ownerID ? 'is-current' : '' ?> <?= $ownerClass ?>">
                        <span class="board-played-owner"><?= htmlspecialchars($ownerName) ?></span>
                        <?= renderCard($pc['Card']) ?>
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
    </div>

    <div class="player-dock board-zone board-zone-hand">
        <section class="player-hand">
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
                    <?php if ($canPlayCard): ?>
                        <form method="post" action="index.php" class="card-form" data-ajax="true">
                            <input type="hidden" name="action" value="play">
                            <input type="hidden" name="cardIndex" value="<?= $idx ?>">
                            <button type="submit" class="card-btn" role="listitem"><?= renderCard($card, false, (string) ($idx + 1)) ?></button>
                        </form>
                    <?php else: ?>
                        <div class="card-btn disabled-card" role="listitem"><?= renderCard($card, false, (string) ($idx + 1)) ?></div>
                    <?php endif; ?>
                <?php endforeach; ?>
            </div>
        </section>

        <section class="player-actions">
            <?php if ($showResponseActions): ?>
                <div class="player-actions-row response-row">
                    <div class="player-actions-copy">
                        <span class="section-kicker"><?= tr('game_action_title_response') ?></span>
                        <strong><?= htmlspecialchars($statusText) ?></strong>
                    </div>
                    <form method="post" action="index.php" data-ajax="true" class="action-form action-form-truco">
                        <input type="hidden" name="action" value="truco">
                        <button type="submit" class="action-btn action-btn-truco" <?= $canTruco ? '' : 'disabled' ?>>
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
                        <button type="submit" class="action-btn action-btn-truco" <?= $canTruco ? '' : 'disabled' ?>>
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

    <div class="board-footer board-zone board-zone-footer">
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
        <details class="table-sidecar">
            <summary>Online mesa</summary>

            <div class="table-ops sidecar-grid">
                <section class="side-block">
                    <h3><?= tr('lobby_events_title') ?></h3>
                    <div class="action-row compact-row">
                        <?php foreach (($lobby['slots'] ?? []) as $idx => $slotName): ?>
                            <?php $slotState = $slotStates[$idx] ?? []; ?>
                            <?php if (($slotState['can_vote_host'] ?? ($idx !== ($lobby['assigned_seat'] ?? -1) && trim((string) $slotName) !== ''))): ?>
                                <form method="post" action="index.php" data-ajax="true">
                                    <input type="hidden" name="action" value="voteHost">
                                    <input type="hidden" name="slot" value="<?= $idx ?>">
                                    <button type="submit" class="btn btn-neutral"><?= tr('action_vote_host') ?> <?= $idx + 1 ?></button>
                                </form>
                            <?php endif; ?>
                            <?php if (($slotState['can_request_replacement'] ?? false)): ?>
                                <form method="post" action="index.php" data-ajax="true">
                                    <input type="hidden" name="action" value="requestReplacementInvite">
                                    <input type="hidden" name="slot" value="<?= $idx ?>">
                                    <button type="submit" class="btn btn-truco"><?= tr('action_replacement_invite') ?> <?= $idx + 1 ?></button>
                                </form>
                            <?php endif; ?>
                        <?php endforeach; ?>
                    </div>
                </section>

                <section class="side-block chat-block">
                    <h3><?= tr('lobby_chat_send') ?></h3>
                    <form method="post" action="index.php" class="lobby-chat-row" data-ajax="true">
                        <input type="hidden" name="action" value="sendChat">
                        <input name="message" class="field" type="text" autocomplete="off" placeholder="<?= tr('chat_placeholder') ?>">
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

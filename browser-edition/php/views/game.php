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
$connection = $bundle['connection'] ?? [];
$diagnostics = $bundle['diagnostics'] ?? [];
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
$myTeam = $myPlayer['Team'] ?? 0;
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
$roundStartID = (int) ($hand['RoundStart'] ?? -1);
$roundStartName = (string) (($playersByID[$roundStartID]['Name'] ?? '') ?: '?');
$lastTrickRound = (int) ($snap['LastTrickRound'] ?? 0);
$lastTrickTie = (bool) ($snap['LastTrickTie'] ?? false);
$lastTrickWinnerID = (int) ($snap['LastTrickWinner'] ?? -1);
$lastTrickWinnerName = (string) (($playersByID[$lastTrickWinnerID]['Name'] ?? '') ?: '?');
$latestLog = trim((string) (count($logs) > 0 ? $logs[count($logs) - 1] : tr('status_ready')));
$lastActorName = $turnPlayerName;
foreach ($players as $candidate) {
    $candidateName = (string) ($candidate['Name'] ?? '');
    if ($candidateName !== '' && str_starts_with($latestLog, $candidateName . ' ')) {
        $lastActorName = $candidateName;
        break;
    }
}
if (str_starts_with($latestLog, 'Nova mão')) {
    $lastActorName = $locale === 'en-US' ? 'Table' : 'Mesa';
}
$recentLogs = array_reverse(array_slice($logs, -4));
$recentEvents = [];
foreach (array_reverse(array_slice($events, -3)) as $ev) {
    $recentEvents[] = formatEventLine($ev);
}
$matchPoints = $snap['MatchPoints'] ?? [];
$team1Score = (int) ($matchPoints[0] ?? $matchPoints['0'] ?? 0);
$team2Score = (int) ($matchPoints[1] ?? $matchPoints['1'] ?? 0);
$team1Tone = $myTeam === 0 ? 'friendly' : 'rival';
$team2Tone = $myTeam === 1 ? 'friendly' : 'rival';
$team1SideLabel = $myTeam === 0 ? tr('game_team_you') : tr('game_team_rival');
$team2SideLabel = $myTeam === 1 ? tr('game_team_you') : tr('game_team_rival');
$team1PillLabel = $myTeam === 0 ? tr('game_you_label') : tr('game_opponent_label');
$team2PillLabel = $myTeam === 1 ? tr('game_you_label') : tr('game_opponent_label');
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
    $centerTitle = strtoupper(raiseLabel($pendingTo ?: 3, $locale));
    $centerNote = tr('game_center_pending_you', raiseLabel($pendingTo ?: 3, $locale));
} elseif ($pendingFor !== -1) {
    $centerTitle = strtoupper(raiseLabel($pendingTo ?: 3, $locale));
    $centerNote = tr('game_center_pending_other', $raiseRequesterName, raiseLabel($pendingTo ?: 3, $locale));
} elseif ($turnPlayer === $myID) {
    $centerTitle = 'VALE ' . $stake;
    $centerNote = tr('game_center_turn_you');
} else {
    $centerTitle = 'VALE ' . $stake;
    $centerNote = tr('game_center_turn_other', $turnPlayerName);
}

if ($canAccept) {
    $actionTitle = tr('game_action_title_response');
    $actionCopy = tr('game_action_response_copy');
} elseif ($turnPlayer === $myID && !$matchFinished) {
    $actionTitle = tr('game_action_title_turn');
    $actionCopy = tr('game_action_turn_copy');
} else {
    $actionTitle = tr('game_action_title_wait');
    $actionCopy = tr('game_action_wait_copy');
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
?>
<section class="panel game-panel">
    <div class="section-head game-head">
        <div>
            <span class="section-kicker"><?= $isOnline ? tr('game_kicker_online') : tr('game_kicker_offline') ?></span>
            <h2><?= $isOnline ? tr('game_title_online') : tr('game_title_offline') ?></h2>
        </div>
        <strong class="mode-pill"><?= htmlspecialchars($mode) ?></strong>
    </div>

    <div class="game-scoreboard">
        <article class="team-score team-score-<?= $team1Tone ?>">
            <span class="score-side-label"><?= $team1SideLabel ?></span>
            <div class="score-team-row">
                <span class="score-label"><?= tr('team1') ?></span>
                <span class="team-pill team-pill-<?= $team1Tone ?>"><?= $team1PillLabel ?></span>
            </div>
            <strong class="score-value"><?= $team1Score ?></strong>
            <div class="score-tricks"><?= tr('tricks_title') ?> <span><?= $team1Tricks ?></span></div>
        </article>

        <article class="status-marquee <?= $pendingFor !== -1 ? 'hot' : '' ?>">
            <div class="status-marquee-top">
                <span class="marquee-tag"><?= tr('game_round_label', $roundNumber) ?></span>
                <span class="marquee-tag"><?= tr('game_turn_chip', $turnPlayerName) ?></span>
                <span class="marquee-tag"><?= tr('game_opening_chip', $roundStartName) ?></span>
                <?php if ($lastTrickText !== ''): ?>
                    <span class="marquee-tag muted"><?= htmlspecialchars($lastTrickText) ?></span>
                <?php endif; ?>
            </div>

            <div class="stake-card <?= $pendingFor !== -1 ? 'hot' : '' ?>">
                <div class="stake-main">
                    <span><?= tr('stake') ?></span>
                    <strong><?= $stake ?></strong>
                </div>
                <div class="stake-ladder">
                    <?php foreach ($stakeSteps as $step): ?>
                        <span class="stake-step <?= $step === $stake ? 'current' : ($step < $stake ? 'done' : ($pendingTo === $step ? 'pending' : 'future')) ?>">
                            <span class="stake-dot"></span>
                            <span class="stake-label"><?= $step ?></span>
                        </span>
                    <?php endforeach; ?>
                </div>
            </div>

            <div class="status-marquee-copy">
                <div class="turn-line <?= $turnPlayer === $myID ? 'my-turn' : '' ?>">
                    <?= tr('turn_of', $turnPlayerName) ?>
                </div>
                <div class="match-ribbon">
                    <span class="ribbon-chip"><?= 'Vale ' . $stake ?></span>
                    <?php if ($pendingFor !== -1): ?>
                        <span class="ribbon-chip hot"><?= strtoupper(raiseLabel($pendingTo ?: 3, $locale)) ?></span>
                    <?php endif; ?>
                    <span class="ribbon-note"><?= htmlspecialchars($statusText) ?></span>
                    <?php if ($isOnline): ?>
                        <span class="ribbon-chip muted"><?= !empty($connection['is_online']) ? tr('connection_online') : tr('connection_offline') ?></span>
                    <?php endif; ?>
                </div>
            </div>
        </article>

        <article class="team-score team-score-<?= $team2Tone ?>">
            <span class="score-side-label"><?= $team2SideLabel ?></span>
            <div class="score-team-row">
                <span class="score-label"><?= tr('team2') ?></span>
                <span class="team-pill team-pill-<?= $team2Tone ?>"><?= $team2PillLabel ?></span>
            </div>
            <strong class="score-value"><?= $team2Score ?></strong>
            <div class="score-tricks"><?= tr('tricks_title') ?> <span><?= $team2Tricks ?></span></div>
        </article>
    </div>

    <div class="table-pulse">
        <div class="pulse-leading">
            <span class="pulse-kicker"><?= tr('status_title') ?></span>
            <strong><?= htmlspecialchars($latestLog) ?></strong>
        </div>
        <div class="pulse-tags">
            <span class="pulse-tag"><?= tr('game_last_speaker_chip', $lastActorName) ?></span>
            <?php if ($lastTrickText !== ''): ?>
                <span class="pulse-tag muted"><?= htmlspecialchars($lastTrickText) ?></span>
            <?php endif; ?>
        </div>
    </div>

    <div class="table-panel players-<?= $numPlayers ?>">
        <div class="table-rail table-rail-left">
            <div class="info-stack">
                <div class="info-card info-card-vira">
                    <span class="info-kicker"><?= tr('vira') ?></span>
                    <div class="meta-card"><?= !empty($hand['Vira']) ? renderCard($hand['Vira'], true) : '' ?></div>
                    <strong><?= htmlspecialchars($viraLabel) ?></strong>
                </div>
                <div class="info-card info-card-manilha">
                    <span class="info-kicker"><?= tr('manilha') ?></span>
                    <div class="manilha-pill <?= !empty($hand['Manilha']) && $hand['Manilha'] !== '-' ? 'hot' : '' ?>">
                        <?= $manilhaLabel ?>
                    </div>
                    <small><?= tr('game_table_note') ?></small>
                </div>
            </div>
        </div>

        <?php foreach ($players as $p): ?>
            <?php
            $playerID = (int) ($p['ID'] ?? -1);
            $pos = $seatByID[$playerID] ?? 'top';
            $teamNum = (int) (($p['Team'] ?? 0) + 1);
            $team = (int) ($p['Team'] ?? 0);
            $cardCount = count($p['Hand'] ?? []);
            $isSelf = $playerID === $myID;
            $isTurn = $playerID === $turnPlayer;
            $isFriendly = $team === $myTeam;
            $seatRole = $isSelf ? tr('game_you_label') : ($isFriendly ? tr('game_partner_label') : tr('game_opponent_label'));
            $seatClasses = [
                'seat',
                'seat-' . $pos,
                $isFriendly ? 'seat-friendly' : 'seat-rival',
                $isSelf ? 'seat-self' : '',
                $isTurn ? 'is-active' : '',
                $playerID === $lastPlayedPlayerID ? 'just-played' : '',
                $playerID === $lastTrickWinnerID ? 'won-last-trick' : '',
                ($pendingFor !== -1 && $team === $pendingFor) ? 'awaiting-answer' : '',
            ];
            ?>
            <div class="<?= trim(implode(' ', array_filter($seatClasses))) ?>">
                <div class="seat-head">
                    <div class="seat-avatar team-<?= $teamNum ?>"><?= htmlspecialchars(strtoupper(substr((string) ($p['Name'] ?? '?'), 0, 1))) ?></div>
                    <div class="seat-pill team-<?= $teamNum ?>">
                        <div class="seat-topline">
                            <span class="seat-name"><?= htmlspecialchars((string) ($p['Name'] ?? '?')) ?></span>
                            <?php if ($isTurn): ?>
                                <span class="seat-tag turn"><?= tr('game_turn_badge') ?></span>
                            <?php elseif ($isSelf): ?>
                                <span class="seat-tag self"><?= tr('game_you_label') ?></span>
                            <?php endif; ?>
                        </div>
                        <span class="seat-role"><?= $seatRole ?> · T<?= $teamNum ?><?= !empty($p['CPU']) ? ' · ' . tr('cpu_tag') : '' ?></span>
                        <div class="seat-tags">
                            <?php if ($roundStartID === $playerID): ?>
                                <span class="seat-tag subtle"><?= tr('game_opening_badge') ?></span>
                            <?php endif; ?>
                            <?php if ($lastTrickWinnerID === $playerID): ?>
                                <span class="seat-tag subtle"><?= tr('game_last_trick_badge') ?></span>
                            <?php endif; ?>
                            <?php if ($pendingFor !== -1 && $team === $pendingFor): ?>
                                <span class="seat-tag pressure"><?= tr('game_pending_badge') ?></span>
                            <?php endif; ?>
                        </div>
                    </div>
                </div>

                <?php if (!$isSelf): ?>
                    <div class="seat-meta seat-card-stack" aria-hidden="true">
                        <?php for ($i = 0; $i < $cardCount; $i++): ?>
                            <span class="tiny-back"></span>
                        <?php endfor; ?>
                    </div>
                <?php else: ?>
                    <div class="seat-meta seat-footnote"><?= tr('game_hand_count', count($myCards)) ?></div>
                <?php endif; ?>
            </div>
        <?php endforeach; ?>

        <div class="table-center">
            <div class="center-callout <?= $pendingFor !== -1 ? 'hot' : '' ?>">
                <span class="center-kicker"><?= tr('game_round_title') ?></span>
                <strong><?= htmlspecialchars($centerTitle) ?></strong>
                <span class="center-note"><?= htmlspecialchars($centerNote) ?></span>
            </div>

            <div class="trick-badges" aria-label="<?= tr('tricks_title') ?>">
                <?php for ($i = 0; $i < 3; $i++): ?>
                    <?php
                    $trickResult = $hand['TrickResults'][$i] ?? null;
                    $badgeClass = 'pending';
                    $badgeText = '...';
                    if ($trickResult === -1) {
                        $badgeClass = 'tie';
                        $badgeText = tr('trick_tie');
                    } elseif ($trickResult === 0 || $trickResult === 1) {
                        $badgeClass = ((int) $trickResult === $myTeam) ? 'win' : 'loss';
                        $badgeText = 'T' . (((int) $trickResult) + 1);
                    }
                    ?>
                    <span class="trick-badge <?= $badgeClass ?>">
                        <small><?= tr('trick_short') . ($i + 1) ?></small>
                        <strong><?= htmlspecialchars($badgeText) ?></strong>
                    </span>
                <?php endfor; ?>
            </div>

            <div class="played-layer <?= empty($roundCards) ? 'is-empty' : '' ?>">
                <?php if (empty($roundCards)): ?>
                    <div class="table-silence"><?= htmlspecialchars($statusText) ?></div>
                <?php endif; ?>

                <?php foreach ($roundCards as $index => $pc): ?>
                    <?php
                    $ownerID = (int) ($pc['PlayerID'] ?? -1);
                    $ownerName = (string) (($playersByID[$ownerID]['Name'] ?? '') ?: ('P' . ($ownerID + 1)));
                    $cardPos = $seatByID[$ownerID] ?? 'top';
                    $cardClasses = [
                        'played-card',
                        'played-' . $cardPos,
                        $index === count($roundCards) - 1 ? 'last-played' : '',
                    ];
                    ?>
                    <div class="<?= trim(implode(' ', array_filter($cardClasses))) ?>">
                        <div class="owner"><?= htmlspecialchars($ownerName) ?></div>
                        <?= renderCard($pc['Card'], true) ?>
                    </div>
                <?php endforeach; ?>
            </div>
        </div>
    </div>

    <div class="play-zone">
        <section class="hand-salon">
            <div class="hand-summary">
                <div>
                    <span class="section-kicker"><?= tr('hand_title') ?></span>
                    <h3><?= tr('game_hand_heading') ?></h3>
                </div>
                <div class="hand-chips">
                    <span class="hand-chip hand-chip-primary"><?= $canPlayCard ? tr('game_hand_ready') : tr('game_hand_wait') ?></span>
                    <span class="hand-chip"><?= tr('game_hand_count', count($myCards)) ?></span>
                    <span class="hand-chip"><?= tr('vira') . ' ' . htmlspecialchars($viraLabel) ?></span>
                    <span class="hand-chip"><?= tr('manilha') . ' ' . $manilhaLabel ?></span>
                </div>
            </div>

            <div class="my-hand premium-hand" role="list">
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

        <div class="under-table">
            <section class="action-dock <?= $canAccept ? 'response-mode' : (($turnPlayer === $myID && !$matchFinished) ? 'turn-mode' : 'wait-mode') ?>">
            <div class="action-copy">
                <span class="section-kicker"><?= $canAccept ? tr('game_action_answer_tag') : tr('status_title') ?></span>
                <h3><?= htmlspecialchars($actionTitle) ?></h3>
                <p><?= htmlspecialchars($actionCopy) ?></p>
            </div>

            <div class="action-grid">
                <form method="post" action="index.php" data-ajax="true" class="action-form action-form-truco">
                    <input type="hidden" name="action" value="truco">
                    <button type="submit" class="action-btn action-btn-truco <?= $canTruco ? 'armed' : '' ?>" <?= $canTruco ? '' : 'disabled' ?>>
                        <span class="action-eyebrow"><?= $canAccept ? tr('game_action_raise_tag') : tr('game_action_call_tag') ?></span>
                        <strong><?= strtoupper($canAccept ? tr('btn_raise') : tr('btn_truco')) ?></strong>
                        <span class="action-sub"><?= htmlspecialchars($canAccept ? tr('game_action_raise_sub', raiseLabel($nextRaisePreview, $locale)) : tr('game_action_call_sub', raiseLabel($nextStake, $locale))) ?></span>
                    </button>
                </form>

                <form method="post" action="index.php" data-ajax="true" class="action-form">
                    <input type="hidden" name="action" value="accept">
                    <button type="submit" class="action-btn action-btn-accept" <?= $canAccept ? '' : 'disabled' ?>>
                        <span class="action-eyebrow"><?= tr('game_action_answer_tag') ?></span>
                        <strong><?= strtoupper(tr('btn_accept')) ?></strong>
                        <span class="action-sub"><?= htmlspecialchars(tr('game_action_accept_sub', $pendingTo ?: 3)) ?></span>
                    </button>
                </form>

                <form method="post" action="index.php" data-ajax="true" class="action-form">
                    <input type="hidden" name="action" value="refuse">
                    <button type="submit" class="action-btn action-btn-refuse" <?= $canRefuse ? '' : 'disabled' ?>>
                        <span class="action-eyebrow"><?= tr('game_action_answer_tag') ?></span>
                        <strong><?= strtoupper(tr('btn_refuse')) ?></strong>
                        <span class="action-sub"><?= tr('game_action_refuse_sub') ?></span>
                    </button>
                </form>

                <form method="post" action="index.php" data-ajax="true" class="action-form action-form-utility">
                    <input type="hidden" name="action" value="refreshGame">
                    <button type="submit" class="action-btn action-btn-utility">
                        <span class="action-eyebrow"><?= tr('game_action_refresh_tag') ?></span>
                        <strong><?= tr('refresh') ?></strong>
                        <span class="action-sub"><?= tr('game_action_refresh_sub') ?></span>
                    </button>
                </form>

                <?php if ($isOnline): ?>
                    <form method="post" action="index.php" data-ajax="true" class="action-form action-form-utility">
                        <input type="hidden" name="action" value="leaveLobby">
                        <button type="submit" class="action-btn action-btn-leave">
                            <span class="action-eyebrow"><?= tr('game_action_leave_tag') ?></span>
                            <strong><?= tr('lobby_leave') ?></strong>
                            <span class="action-sub"><?= tr('game_action_leave_sub') ?></span>
                        </button>
                    </form>
                <?php endif; ?>
            </div>
            </section>

            <aside class="table-notes">
                <div class="notes-head">
                    <div>
                        <span class="section-kicker"><?= tr('game_notes_title') ?></span>
                        <h3><?= tr('log_title') ?></h3>
                    </div>
                    <?php if ($isOnline): ?>
                        <span class="notes-connection <?= !empty($connection['is_online']) ? 'online' : '' ?>">
                            <?= !empty($connection['is_online']) ? tr('connection_online') : tr('connection_offline') ?>
                        </span>
                    <?php endif; ?>
                </div>

                <div class="table-feed">
                    <?php if (empty($recentLogs)): ?>
                        <div class="feed-line muted"><?= tr('game_notes_empty') ?></div>
                    <?php else: ?>
                        <?php foreach ($recentLogs as $line): ?>
                            <div class="feed-line"><?= htmlspecialchars((string) $line) ?></div>
                        <?php endforeach; ?>
                    <?php endif; ?>
                </div>

                <?php if ($isOnline): ?>
                    <div class="online-notes">
                        <span class="info-kicker"><?= tr('game_online_notes_title') ?></span>
                        <?php if (empty($recentEvents)): ?>
                            <div class="feed-line muted"><?= tr('lobby_events_empty') ?></div>
                        <?php else: ?>
                            <?php foreach ($recentEvents as $line): ?>
                                <div class="feed-line compact"><?= htmlspecialchars($line) ?></div>
                            <?php endforeach; ?>
                        <?php endif; ?>
                        <div class="notes-meta">
                            <span><?= tr('connection_backlog') ?></span>
                            <strong><?= (int) ($diagnostics['event_backlog'] ?? 0) ?></strong>
                        </div>
                        <?php if (!empty($connection['last_error']['message'])): ?>
                            <div class="notes-error"><?= htmlspecialchars((string) $connection['last_error']['message']) ?></div>
                        <?php endif; ?>
                    </div>
                <?php endif; ?>
            </aside>
        </div>
    </div>

    <?php if ($isOnline): ?>
        <div class="table-ops">
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
                    <button type="submit" class="btn btn-neutral">💬 <?= tr('lobby_chat_send') ?></button>
                </form>
            </section>
        </div>
    <?php endif; ?>

    <?php if ($matchFinished): ?>
        <div class="overlay" style="display:flex">
            <div class="overlay-card match-card">
                <h2><?= $winnerTeam === $myTeam ? tr('overlay_match_win') : tr('overlay_match_loss') ?></h2>
                <p><?= tr('overlay_match_detail', $team1Score, $team2Score) ?></p>
                <form method="post" action="index.php" data-ajax="true">
                    <input type="hidden" name="action" value="reset">
                    <button type="submit" class="btn btn-primary">🔄 <?= tr('btn_play_again') ?></button>
                </form>
            </div>
        </div>
    <?php endif; ?>
</section>

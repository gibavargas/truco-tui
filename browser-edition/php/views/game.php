<?php
require_once __DIR__ . '/partials/card.php';

$bundle = $_SESSION['runtime_bundle'] ?? [];
$events = $_SESSION['runtime_events'] ?? [];
$players = $snap['Players'] ?? [];
$myID = $snap['CurrentPlayerIdx'] ?? 0;
$turnPlayer = $snap['TurnPlayer'] ?? -1;
$pendingFor = $snap['PendingRaiseFor'] ?? -1;
$matchFinished = $snap['MatchFinished'] ?? false;
$winnerTeam = $snap['WinnerTeam'] ?? -1;
$hand = $snap['CurrentHand'] ?? [];
$mode = $bundle['mode'] ?? 'offline_match';
$isOnline = strpos($mode, 'host_') === 0 || strpos($mode, 'client_') === 0;
$lobby = $bundle['lobby'] ?? [];
$myPlayer = null;
foreach ($players as $p) {
    if (($p['ID'] ?? -1) === $myID) {
        $myPlayer = $p;
        break;
    }
}
$myCards = $myPlayer['Hand'] ?? [];
$myTeam = $myPlayer['Team'] ?? 0;
$stake = $hand['Stake'] ?? 1;
$pendingTo = $snap['PendingRaiseTo'] ?? 0;
$turnPlayerObj = null;
foreach ($players as $p) {
    if (($p['ID'] ?? -1) === $turnPlayer) {
        $turnPlayerObj = $p;
        break;
    }
}
$turnName = htmlspecialchars($turnPlayerObj['Name'] ?? '?');
$canPlayCard = !$matchFinished && $turnPlayer === $myID && $pendingFor === -1;
$canTruco = !$matchFinished && (($turnPlayer === $myID && $pendingFor === -1) || ($pendingFor !== -1 && $pendingFor === $myTeam));
$canAccept = !$matchFinished && $pendingFor !== -1 && $pendingFor === $myTeam;
$locale = $_SESSION['locale'] ?? 'pt-BR';
?>
<section class="panel game-panel">
    <div class="hud">
        <div class="score-card team-a">
            <span class="score-label"><?= tr('team1') ?></span>
            <strong class="score-value"><?= $snap['MatchPoints'][0] ?? $snap['MatchPoints']['0'] ?? 0 ?></strong>
        </div>
        <div class="stake-card">
            <div class="stake-main"><span><?= tr('stake') ?></span> <strong><?= $stake ?></strong></div>
            <div class="stake-ladder">
                <?php foreach ([1, 3, 6, 9, 12] as $step): ?>
                    <span class="stake-step <?= $step === $stake ? 'current' : ($step < $stake ? 'done' : ($pendingTo === $step ? 'pending' : 'future')) ?>">
                        <span class="stake-dot"></span><span class="stake-label"><?= $step ?></span>
                    </span>
                <?php endforeach; ?>
            </div>
        </div>
        <div class="score-card team-b">
            <span class="score-label"><?= tr('team2') ?></span>
            <strong class="score-value"><?= $snap['MatchPoints'][1] ?? $snap['MatchPoints']['1'] ?? 0 ?></strong>
        </div>
    </div>

    <div class="turn-line <?= $turnPlayer === $myID ? 'my-turn' : '' ?>">
        <?= tr('turn_of', $turnName) ?>
    </div>

    <div class="layout">
        <div class="table-panel">
            <?php
            $numPlayers = $snap['NumPlayers'] ?? 2;
            foreach ($players as $p):
                $pos = ($numPlayers === 2)
                    ? (($p['ID'] ?? 0) === $myID ? 'bottom' : 'top')
                    : (['bottom', 'right', 'top', 'left'][($p['ID'] - $myID + 4) % 4] ?? 'top');
                $isTurn = (($p['ID'] ?? -1) === $turnPlayer);
                $teamNum = ($p['Team'] ?? 0) + 1;
                $cardCount = count($p['Hand'] ?? []);
            ?>
                <div class="seat seat-<?= $pos ?>">
                    <div class="seat-head">
                        <div class="seat-avatar team-<?= $teamNum ?>"><?= htmlspecialchars(strtoupper(substr($p['Name'] ?? '?', 0, 1))) ?></div>
                        <div class="seat-pill team-<?= $teamNum ?> <?= $isTurn ? 'active' : '' ?>">
                            <span class="seat-name"><?= htmlspecialchars($p['Name']) ?></span>
                            <span class="seat-team">T<?= $teamNum ?><?= !empty($p['CPU']) ? ' · ' . tr('cpu_tag') : '' ?></span>
                        </div>
                    </div>
                    <div class="seat-meta">
                        <?php if (($p['ID'] ?? -1) !== $myID): ?>
                            <?php for ($i = 0; $i < $cardCount; $i++): ?><span class="tiny-back"></span><?php endfor; ?>
                        <?php endif; ?>
                    </div>
                </div>
            <?php endforeach; ?>

            <div class="table-center">
                <div class="center-meta">
                    <div class="meta-chip">
                        <span><?= tr('vira') ?></span>
                        <div class="meta-card"><?= !empty($hand['Vira']) ? renderCard($hand['Vira'], true) : '' ?></div>
                    </div>
                    <div class="meta-chip">
                        <span><?= tr('manilha') ?></span>
                        <div class="manilha-pill <?= !empty($hand['Manilha']) && $hand['Manilha'] !== '-' ? 'hot' : '' ?>">
                            <?= htmlspecialchars($hand['Manilha'] ?? '-') ?>
                        </div>
                    </div>
                </div>
                <div class="played-layer">
                    <?php foreach (($hand['RoundCards'] ?? []) as $pc): ?>
                        <div class="played-card">
                            <div class="owner"><?= htmlspecialchars($players[$pc['PlayerID']]['Name'] ?? ('P' . (($pc['PlayerID'] ?? 0) + 1))) ?></div>
                            <?= renderCard($pc['Card'], true) ?>
                        </div>
                    <?php endforeach; ?>
                </div>
            </div>
        </div>

        <aside class="side-panel">
            <div class="side-block">
                <h3><?= tr('status_title') ?></h3>
                <div class="status-line">
                    <?php
                    if ($matchFinished) {
                        echo htmlspecialchars(tr('status_match_end'));
                    } elseif ($pendingFor !== -1 && $pendingFor === $myTeam) {
                        echo htmlspecialchars(tr('status_pending_you', raiseLabel($pendingTo ?: 3, $locale), $pendingTo ?: 3));
                    } elseif ($pendingFor !== -1) {
                        echo htmlspecialchars(tr('status_pending_other', $turnName, raiseLabel($pendingTo ?: 3, $locale), $pendingTo ?: 3));
                    } elseif ($turnPlayer === $myID) {
                        echo htmlspecialchars(tr('status_your_turn'));
                    } else {
                        echo htmlspecialchars(tr('status_wait_cpu', $turnName));
                    }
                    ?>
                </div>
            </div>
            <div class="side-block side-log">
                <h3><?= tr('log_title') ?></h3>
                <pre><?= htmlspecialchars(implode("\n", array_slice($snap['Logs'] ?? [], -18))) ?></pre>
            </div>
            <?php if ($isOnline): ?>
                <div class="side-block side-log">
                    <h3><?= tr('game_events_title') ?></h3>
                    <pre><?php
                        foreach (array_slice($events, -18) as $ev) {
                            $line = $ev['kind'] ?? 'event';
                            if (!empty($ev['payload']['text'])) {
                                $line .= ' · ' . $ev['payload']['text'];
                            } elseif (!empty($ev['payload']['invite_key'])) {
                                $line .= ' · ' . $ev['payload']['invite_key'];
                            }
                            echo htmlspecialchars($line) . "\n";
                        }
                    ?></pre>
                </div>
            <?php endif; ?>
        </aside>
    </div>

    <div class="action-row">
        <form method="post" action="index.php" style="display:inline" data-ajax="true">
            <input type="hidden" name="action" value="truco">
            <button type="submit" class="btn btn-truco <?= $canTruco ? 'armed' : '' ?>" <?= $canTruco ? '' : 'disabled' ?>>⚡ <?= $canAccept ? tr('btn_raise') : tr('btn_truco') ?></button>
        </form>
        <form method="post" action="index.php" style="display:inline" data-ajax="true">
            <input type="hidden" name="action" value="accept">
            <button type="submit" class="btn btn-accept" <?= $canAccept ? '' : 'disabled' ?>>✓ <?= tr('btn_accept') ?></button>
        </form>
        <form method="post" action="index.php" style="display:inline" data-ajax="true">
            <input type="hidden" name="action" value="refuse">
            <button type="submit" class="btn btn-refuse" <?= $canAccept ? '' : 'disabled' ?>>✕ <?= tr('btn_refuse') ?></button>
        </form>
        <form method="post" action="index.php" style="display:inline" data-ajax="true">
            <input type="hidden" name="action" value="refreshGame">
            <button type="submit" class="btn btn-neutral">↺ <?= tr('refresh') ?></button>
        </form>
        <?php if ($isOnline): ?>
            <form method="post" action="index.php" style="display:inline" data-ajax="true">
                <input type="hidden" name="action" value="leaveLobby">
                <button type="submit" class="btn btn-refuse">⎋ <?= tr('lobby_leave') ?></button>
            </form>
        <?php endif; ?>
    </div>

    <div class="hand-block">
        <div class="hand-title"><?= tr('hand_title') ?></div>
        <div class="my-hand" role="list">
            <?php foreach ($myCards as $idx => $card): ?>
                <?php if ($canPlayCard): ?>
                    <form method="post" action="index.php" style="display:inline" class="card-form" data-ajax="true">
                        <input type="hidden" name="action" value="play">
                        <input type="hidden" name="cardIndex" value="<?= $idx ?>">
                        <button type="submit" class="card-btn" role="listitem"><?= renderCard($card, false, (string) ($idx + 1)) ?></button>
                    </form>
                <?php else: ?>
                    <div class="card-btn" role="listitem" style="opacity:0.56"><?= renderCard($card, false, (string) ($idx + 1)) ?></div>
                <?php endif; ?>
            <?php endforeach; ?>
        </div>
    </div>

    <?php if ($isOnline): ?>
        <div class="side-block" style="margin-top:12px;">
            <h3><?= tr('lobby_events_title') ?></h3>
            <div class="action-row" style="margin-bottom:8px;">
                <?php foreach (($lobby['slots'] ?? []) as $idx => $slotName): ?>
                    <?php if ($idx !== ($lobby['assigned_seat'] ?? -1) && trim((string) $slotName) !== ''): ?>
                        <form method="post" action="index.php" style="display:inline" data-ajax="true">
                            <input type="hidden" name="action" value="voteHost">
                            <input type="hidden" name="slot" value="<?= $idx ?>">
                            <button type="submit" class="btn btn-neutral"><?= tr('action_vote_host') ?> <?= $idx + 1 ?></button>
                        </form>
                    <?php endif; ?>
                    <?php if (strpos($mode, 'host_') === 0 && trim((string) $slotName) === ''): ?>
                        <form method="post" action="index.php" style="display:inline" data-ajax="true">
                            <input type="hidden" name="action" value="requestReplacementInvite">
                            <input type="hidden" name="slot" value="<?= $idx ?>">
                            <button type="submit" class="btn btn-truco"><?= tr('action_replacement_invite') ?> <?= $idx + 1 ?></button>
                        </form>
                    <?php endif; ?>
                <?php endforeach; ?>
            </div>
            <form method="post" action="index.php" class="lobby-chat-row" data-ajax="true">
                <input type="hidden" name="action" value="sendChat">
                <input name="message" class="field" type="text" autocomplete="off" placeholder="<?= tr('chat_placeholder') ?>">
                <button type="submit" class="btn btn-neutral">💬 <?= tr('lobby_chat_send') ?></button>
            </form>
        </div>
    <?php endif; ?>

    <?php if ($matchFinished): ?>
        <div class="overlay" style="display:flex">
            <div class="overlay-card match-card">
                <h2><?= $winnerTeam === $myTeam ? tr('overlay_match_win') : tr('overlay_match_loss') ?></h2>
                <p><?= tr('overlay_match_detail', $snap['MatchPoints'][0] ?? 0, $snap['MatchPoints'][1] ?? 0) ?></p>
                <form method="post" action="index.php" data-ajax="true">
                    <input type="hidden" name="action" value="reset">
                    <button type="submit" class="btn btn-primary">🔄 <?= tr('btn_play_again') ?></button>
                </form>
            </div>
        </div>
    <?php endif; ?>
</section>

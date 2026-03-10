<?php
/**
 * Game view — Server-rendered game board.
 * Expects $snap (decoded snapshot array) and $statusMsg / $errorMsg.
 */

// Pre-compute i18n globals used by partials
$GLOBALS['_tr_team1'] = tr('team1');
$GLOBALS['_tr_team2'] = tr('team2');
$GLOBALS['_tr_stake'] = tr('stake');
$GLOBALS['_tr_turn_prefix'] = tr('turn_of', '');
// Fix: the turn_of format has %s, so we rebuild it properly below.

require_once __DIR__ . '/partials/card.php';
require_once __DIR__ . '/partials/hud.php';

$players = $snap['Players'] ?? [];
$myID = 0; // local seat is always 0
$myTeam = 0;
$turnPlayer = $snap['TurnPlayer'] ?? -1;
$pendingFor = $snap['PendingRaiseFor'] ?? -1;
$matchFinished = $snap['MatchFinished'] ?? false;
$winnerTeam = $snap['WinnerTeam'] ?? -1;
$hand = $snap['CurrentHand'] ?? [];
$myPlayer = null;
foreach ($players as $p) {
    if ($p['ID'] === $myID) {
        $myPlayer = $p;
        break;
    }
}
$myCards = $myPlayer['Hand'] ?? [];
$vira = $hand['Vira'] ?? null;
$manilha = $hand['Manilha'] ?? '-';
$roundCards = $hand['RoundCards'] ?? [];
$logs = $snap['Logs'] ?? [];
$trickWins = $hand['TrickWins'] ?? [0 => 0, 1 => 0];
$tw0 = $trickWins[0] ?? $trickWins['0'] ?? 0;
$tw1 = $trickWins[1] ?? $trickWins['1'] ?? 0;
$stake = $hand['Stake'] ?? 1;
$pendingTo = $snap['PendingRaiseTo'] ?? 0;
$locale = $_SESSION['locale'] ?? 'pt-BR';

// Status
$turnPlayerObj = null;
foreach ($players as $p) {
    if ($p['ID'] === $turnPlayer) {
        $turnPlayerObj = $p;
        break;
    }
}
$turnName = htmlspecialchars($turnPlayerObj['Name'] ?? '?');

// Overwrite the global for hud partial
$GLOBALS['_tr_turn_prefix'] = '';

if ($matchFinished) {
    $statusMsg = tr('status_match_end');
} elseif ($pendingFor !== -1 && $pendingFor === $myTeam) {
    $statusMsg = tr('status_pending_you', raiseLabel($pendingTo ?: 3, $locale), $pendingTo ?: 3);
} elseif ($pendingFor !== -1) {
    $statusMsg = tr('status_pending_other', $turnName, raiseLabel($pendingTo ?: 3, $locale), $pendingTo ?: 3);
} elseif ($turnPlayer === $myID) {
    $statusMsg = tr('status_your_turn');
} else {
    $statusMsg = tr('status_wait_cpu', $turnName);
}

$canPlayCard = !$matchFinished && $turnPlayer === $myID && $pendingFor === -1;
$canTruco = !$matchFinished && (($turnPlayer === $myID && $pendingFor === -1) || ($pendingFor !== -1 && $pendingFor === $myTeam));
$canAccept = !$matchFinished && $pendingFor !== -1 && $pendingFor === $myTeam;
$canRefuse = $canAccept;
?>

<section class="panel game-panel">
    <!-- HUD -->
    <div class="hud">
        <div class="score-card team-a">
            <span class="score-label">
                <?= tr('team1') ?>
            </span>
            <strong class="score-value">
                <?= $snap['MatchPoints'][0] ?? $snap['MatchPoints']['0'] ?? 0 ?>
            </strong>
        </div>
        <div class="stake-card">
            <div class="stake-main"><span>
                    <?= tr('stake') ?>
                </span> <strong>
                    <?= $stake ?>
                </strong></div>
            <div class="stake-ladder">
                <?php foreach ([1, 3, 6, 9, 12] as $step):
                    $cls = 'future';
                    if ($step === $stake)
                        $cls = 'current';
                    elseif ($step < $stake)
                        $cls = 'done';
                    if ($pendingTo > 0 && $step === $pendingTo)
                        $cls = 'pending';
                    ?>
                    <span class="stake-step <?= $cls ?>"><span class="stake-dot"></span><span class="stake-label">
                            <?= $step ?>
                        </span></span>
                <?php endforeach; ?>
            </div>
        </div>
        <div class="score-card team-b">
            <span class="score-label">
                <?= tr('team2') ?>
            </span>
            <strong class="score-value">
                <?= $snap['MatchPoints'][1] ?? $snap['MatchPoints']['1'] ?? 0 ?>
            </strong>
        </div>
    </div>

    <div class="turn-line <?= $turnPlayer === $myID ? 'my-turn' : '' ?>">
        <?= tr('turn_of', $turnName) ?>
    </div>

    <!-- Layout: table + sidebar -->
    <div class="layout">
        <!-- Table panel -->
        <div class="table-panel">
            <!-- Seats -->
            <?php
            $numPlayers = $snap['NumPlayers'] ?? 2;
            foreach ($players as $p):
                $pos = ($numPlayers === 2)
                    ? ($p['ID'] === 0 ? 'bottom' : 'top')
                    : (['bottom', 'right', 'top', 'left'][$p['ID']] ?? 'top');
                $isTurn = ($p['ID'] === $turnPlayer);
                $teamNum = ($p['Team'] ?? 0) + 1;
                $cpuTag = ($p['CPU'] ?? false) ? ' · ' . tr('cpu_tag') : '';
                $avatar = ($p['ID'] === 0) ? '★' : (($p['CPU'] ?? false) ? '🤖' : strtoupper(substr($p['Name'] ?? '?', 0, 1)));
                $cardCount = count($p['Hand'] ?? []);
                $tinyBacks = '';
                if ($p['ID'] !== 0) {
                    for ($i = 0; $i < $cardCount; $i++) {
                        $tinyBacks .= '<span class="tiny-back"></span>';
                    }
                }
                ?>
                <div class="seat seat-<?= $pos ?> <?= $isTurn ? '' : '' ?>">
                    <div class="seat-head">
                        <div class="seat-avatar team-<?= $teamNum ?>">
                            <?= htmlspecialchars($avatar) ?>
                        </div>
                        <div class="seat-pill team-<?= $teamNum ?> <?= $isTurn ? 'active' : '' ?>">
                            <span class="seat-name">
                                <?= htmlspecialchars($p['Name']) ?>
                            </span>
                            <span class="seat-team">T
                                <?= $teamNum ?>
                                <?= htmlspecialchars($cpuTag) ?>
                            </span>
                        </div>
                    </div>
                    <div class="seat-meta">
                        <?= $tinyBacks ?>
                    </div>
                </div>
            <?php endforeach; ?>

            <!-- Center: vira + manilha + played cards -->
            <div class="table-center">
                <div class="center-meta">
                    <div class="meta-chip">
                        <span>
                            <?= tr('vira') ?>
                        </span>
                        <div class="meta-card">
                            <?php if ($vira): ?>
                                <?= renderCard($vira, true) ?>
                            <?php endif; ?>
                        </div>
                    </div>
                    <div class="meta-chip">
                        <span>
                            <?= tr('manilha') ?>
                        </span>
                        <div class="manilha-pill <?= $manilha !== '-' ? 'hot' : '' ?>">
                            <?php if ($manilha !== '-'): ?>
                                <span class="spark">✦</span><strong>
                                    <?= htmlspecialchars($manilha) ?>
                                </strong><span class="spark">✦</span>
                            <?php else: ?>
                                -
                            <?php endif; ?>
                        </div>
                    </div>
                </div>

                <div class="played-layer">
                    <?php foreach ($roundCards as $pc):
                        $pcPos = ($numPlayers === 2)
                            ? ($pc['PlayerID'] === 0 ? 'bottom' : 'top')
                            : (['bottom', 'right', 'top', 'left'][$pc['PlayerID']] ?? 'top');
                        $coords = [
                            'top' => ['x' => 50, 'y' => 24],
                            'right' => ['x' => 73, 'y' => 50],
                            'bottom' => ['x' => 50, 'y' => 76],
                            'left' => ['x' => 27, 'y' => 50],
                        ];
                        $anchor = $coords[$pcPos] ?? $coords['top'];
                        $ownerName = '';
                        foreach ($players as $pp) {
                            if ($pp['ID'] === $pc['PlayerID']) {
                                $ownerName = $pp['Name'];
                                break;
                            }
                        }
                        ?>
                        <div class="played-card" style="left:<?= $anchor['x'] ?>%;top:<?= $anchor['y'] ?>%;">
                            <div class="owner">
                                <?= htmlspecialchars($ownerName) ?>
                            </div>
                            <?= renderCard($pc['Card'], true) ?>
                        </div>
                    <?php endforeach; ?>
                </div>
            </div>
        </div>

        <!-- Sidebar -->
        <aside class="side-panel">
            <div class="side-block">
                <h3>
                    <?= tr('status_title') ?>
                </h3>
                <div class="status-line">
                    <?= htmlspecialchars($statusMsg) ?>
                </div>
            </div>
            <div class="side-block">
                <h3>
                    <?= tr('tricks_title') ?>
                </h3>
                <div class="small-line">
                    <?= tr('score_line', $tw0, $tw1) ?>
                </div>
                <div class="trick-history">
                    <?= renderTrickHistory($snap) ?>
                </div>
            </div>
            <div class="side-block side-log">
                <h3>
                    <?= tr('log_title') ?>
                </h3>
                <pre><?= htmlspecialchars(implode("\n", array_slice($logs, -22))) ?></pre>
            </div>
            <?php if (!empty($errorMsg)): ?>
                <div class="error-log">
                    <?= htmlspecialchars($errorMsg) ?>
                </div>
            <?php endif; ?>
        </aside>
    </div>

    <!-- Action buttons -->
    <div class="action-row">
        <form method="post" action="index.php" style="display:inline">
            <input type="hidden" name="action" value="truco">
            <button type="submit" class="btn btn-truco <?= $canTruco ? 'armed' : '' ?>" <?= $canTruco ? '' : 'disabled' ?>>⚡
                <?= $canAccept ? tr('btn_raise') : tr('btn_truco') ?>
            </button>
        </form>
        <form method="post" action="index.php" style="display:inline">
            <input type="hidden" name="action" value="accept">
            <button type="submit" class="btn btn-accept" <?= $canAccept ? '' : 'disabled' ?>>✓
                <?= tr('btn_accept') ?>
            </button>
        </form>
        <form method="post" action="index.php" style="display:inline">
            <input type="hidden" name="action" value="refuse">
            <button type="submit" class="btn btn-refuse" <?= $canRefuse ? '' : 'disabled' ?>>✕
                <?= tr('btn_refuse') ?>
            </button>
        </form>
        <form method="post" action="index.php" style="display:inline">
            <input type="hidden" name="action" value="autoCpuAndRefresh">
            <button type="submit" class="btn btn-neutral">↺
                <?= tr('refresh') ?>
            </button>
        </form>
    </div>

    <!-- Player's hand -->
    <div class="hand-block">
        <div class="hand-title">
            <?= tr('hand_title') ?>
        </div>
        <div class="my-hand" role="list">
            <?php foreach ($myCards as $idx => $card): ?>
                <?php if ($canPlayCard): ?>
                    <form method="post" action="index.php" style="display:inline" class="card-form">
                        <input type="hidden" name="action" value="play">
                        <input type="hidden" name="cardIndex" value="<?= $idx ?>">
                        <button type="submit" class="card-btn" role="listitem"
                            title="<?= htmlspecialchars(($card['Rank'] ?? '?') . ' ' . suitSymbol($card['Suit'] ?? '')) ?>">
                            <?= renderCard($card, false, (string) ($idx + 1)) ?>
                        </button>
                    </form>
                <?php else: ?>
                    <div class="card-btn" role="listitem" style="opacity:0.56">
                        <?= renderCard($card, false, (string) ($idx + 1)) ?>
                    </div>
                <?php endif; ?>
            <?php endforeach; ?>
        </div>
    </div>

    <?php if ($matchFinished): ?>
        <!-- Match ended overlay -->
        <div class="overlay" style="display:flex">
            <div class="overlay-card match-card">
                <h2>
                    <?= $winnerTeam === $myTeam ? tr('overlay_match_win') : tr('overlay_match_loss') ?>
                </h2>
                <p>
                    <?= tr('overlay_match_detail', $snap['MatchPoints'][0] ?? 0, $snap['MatchPoints'][1] ?? 0) ?>
                </p>
                <form method="post" action="index.php">
                    <input type="hidden" name="action" value="reset">
                    <button type="submit" class="btn btn-primary">🔄
                        <?= tr('btn_play_again') ?>
                    </button>
                </form>
            </div>
        </div>
    <?php endif; ?>
</section>
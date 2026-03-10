<?php
/**
 * Render the HUD (scores, stake, turn indicator).
 *
 * @param array $snap The game snapshot
 */
function renderHud(array $snap): string
{
    $points = $snap['MatchPoints'] ?? [0 => 0, 1 => 0];
    $t1 = $points[0] ?? $points['0'] ?? 0;
    $t2 = $points[1] ?? $points['1'] ?? 0;
    $stake = $snap['CurrentHand']['Stake'] ?? 1;
    $turnPlayer = null;
    foreach (($snap['Players'] ?? []) as $p) {
        if ($p['ID'] === ($snap['TurnPlayer'] ?? -1)) {
            $turnPlayer = $p;
        }
    }
    $turnName = htmlspecialchars($turnPlayer['Name'] ?? '?');
    $myTurn = ($snap['TurnPlayer'] ?? -1) === 0;
    $turnClass = $myTurn ? 'my-turn' : '';

    $stakeSteps = [1, 3, 6, 9, 12];
    $ladderHtml = '';
    foreach ($stakeSteps as $step) {
        $cls = 'future';
        if ($step === $stake) {
            $cls = 'current';
        } elseif ($step < $stake) {
            $cls = 'done';
        }
        $ladderHtml .= '<span class="stake-step ' . $cls . '"><span class="stake-dot"></span><span class="stake-label">' . $step . '</span></span>';
    }

    return <<<HTML
<div class="hud">
  <div class="score-card team-a">
    <span class="score-label">{$GLOBALS['_tr_team1']}</span>
    <strong class="score-value">{$t1}</strong>
  </div>
  <div class="stake-card">
    <div class="stake-main"><span>{$GLOBALS['_tr_stake']}</span> <strong>{$stake}</strong></div>
    <div class="stake-ladder">{$ladderHtml}</div>
  </div>
  <div class="score-card team-b">
    <span class="score-label">{$GLOBALS['_tr_team2']}</span>
    <strong class="score-value">{$t2}</strong>
  </div>
</div>
<div class="turn-line {$turnClass}">{$GLOBALS['_tr_turn_prefix']}{$turnName}</div>
HTML;
}

/**
 * Render trick history badges.
 */
function renderTrickHistory(array $snap): string
{
    $results = $snap['CurrentHand']['TrickResults'] ?? [];
    $myTeam = 0; // local player is always seat 0, team 0
    $short = tr('trick_short');
    $out = '';
    for ($i = 0; $i < 3; $i++) {
        $prefix = $short . ($i + 1);
        if (!isset($results[$i])) {
            $out .= '<span class="trick-badge pending">' . $prefix . ' •••</span>';
            continue;
        }
        $r = $results[$i];
        if ($r === -1) {
            $out .= '<span class="trick-badge tie">' . $prefix . ' =</span>';
            continue;
        }
        $cls = ($r === $myTeam) ? 'win' : 'loss';
        $mark = ($r === $myTeam) ? '✓' : '✗';
        $out .= '<span class="trick-badge ' . $cls . '">' . $prefix . ' ' . $mark . '</span>';
    }
    return $out;
}

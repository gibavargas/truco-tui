<?php
$bundle = $_SESSION['runtime_bundle'] ?? [];
$session = $bundle['lobby'] ?? [];
$events = $_SESSION['runtime_events'] ?? [];
$mode = $bundle['mode'] ?? 'client_lobby';
$isHost = strpos($mode, 'host_') === 0;
$slots = $session['slots'] ?? [];
$connectedSeats = $session['connected_seats'] ?? [];
$assignedSeat = $session['assigned_seat'] ?? -1;
$hostSeat = $session['host_seat'] ?? 0;
$ui = $bundle['ui'] ?? [];
$slotStates = $ui['lobby_slots'] ?? [];
?>
<section class="panel lobby-panel">
    <h2><?= tr('lobby_title') ?></h2>

    <div class="lobby-meta">
        <div><strong><?= htmlspecialchars($mode) ?></strong></div>
        <div><span><?= tr('lobby_invite') ?></span>: <code><?= htmlspecialchars($session['invite_key'] ?? '-') ?></code></div>
    </div>

    <div class="lobby-grid">
        <div class="lobby-block">
            <h3><?= tr('lobby_slots_title') ?></h3>
            <div class="lobby-slots">
                <?php foreach ($slots as $idx => $slotName): ?>
                    <?php
                    $slotState = $slotStates[$idx] ?? [];
                    $connected = (bool) ($slotState['is_connected'] ?? (!empty($connectedSeats[(string) $idx]) || !empty($connectedSeats[$idx])));
                    $isHostSeat = (bool) ($slotState['is_host'] ?? ($idx === $hostSeat));
                    $isLocalSeat = (bool) ($slotState['is_local'] ?? ($idx === $assignedSeat));
                    $canVote = (bool) ($slotState['can_vote_host'] ?? (trim($slotName) !== '' && $idx !== $assignedSeat));
                    $canReplace = (bool) ($slotState['can_request_replacement'] ?? false);
                    $isProvisionalCPU = (bool) ($slotState['is_provisional_cpu'] ?? false);
                    $displayName = trim($slotName) !== '' ? $slotName : tr('lobby_slot_empty');
                    ?>
                    <div class="lobby-slot">
                        <div class="top">
                            <strong>Slot <?= $idx + 1 ?></strong>
                            <span><?= htmlspecialchars($displayName) ?></span>
                        </div>
                        <div class="roles">
                            <?php if ($isLocalSeat): ?><span>you</span><?php endif; ?>
                            <?php if ($isHostSeat): ?><span>host</span><?php endif; ?>
                            <span><?= $connected ? 'online' : 'offline' ?></span>
                            <?php if ($isProvisionalCPU): ?><span>cpu</span><?php endif; ?>
                        </div>
                        <div class="roles">
                            <?php if ($canVote): ?>
                                <form method="post" action="index.php" data-ajax="true">
                                    <input type="hidden" name="action" value="voteHost">
                                    <input type="hidden" name="slot" value="<?= $idx ?>">
                                    <button type="submit" class="btn btn-neutral"><?= tr('action_vote_host') ?></button>
                                </form>
                            <?php endif; ?>
                            <?php if ($canReplace): ?>
                                <form method="post" action="index.php" data-ajax="true">
                                    <input type="hidden" name="action" value="requestReplacementInvite">
                                    <input type="hidden" name="slot" value="<?= $idx ?>">
                                    <button type="submit" class="btn btn-truco"><?= tr('action_replacement_invite') ?></button>
                                </form>
                            <?php endif; ?>
                        </div>
                    </div>
                <?php endforeach; ?>
            </div>
        </div>
        <div class="lobby-block">
            <h3><?= tr('lobby_events_title') ?></h3>
            <pre id="lobby-events"><?php
                if (!$events) {
                    echo htmlspecialchars(tr('lobby_events_empty'));
                } else {
                    foreach ($events as $ev) {
                        $label = '[' . htmlspecialchars(substr((string) ($ev['timestamp'] ?? ''), 11, 8)) . '] ';
                        $label .= htmlspecialchars((string) ($ev['kind'] ?? 'event'));
                        if (!empty($ev['payload']['text'])) {
                            $label .= ' · ' . htmlspecialchars((string) $ev['payload']['text']);
                        } elseif (!empty($ev['payload']['invite_key'])) {
                            $label .= ' · ' . htmlspecialchars((string) $ev['payload']['invite_key']);
                        }
                        echo $label . "\n";
                    }
                }
            ?></pre>
        </div>
    </div>

    <div class="action-row" style="margin-top:12px;">
        <?php if ($isHost): ?>
            <form method="post" action="index.php" style="display:inline" data-ajax="true">
                <input type="hidden" name="action" value="startOnlineMatch">
                <button type="submit" class="btn btn-primary">▶ <?= tr('lobby_start') ?></button>
            </form>
        <?php endif; ?>
        <form method="post" action="index.php" style="display:inline" data-ajax="true">
            <input type="hidden" name="action" value="refreshLobby">
            <button type="submit" class="btn btn-neutral">⟳ <?= tr('lobby_refresh') ?></button>
        </form>
        <form method="post" action="index.php" style="display:inline" data-ajax="true">
            <input type="hidden" name="action" value="leaveLobby">
            <button type="submit" class="btn btn-refuse">⎋ <?= tr('lobby_leave') ?></button>
        </form>
    </div>

    <form method="post" action="index.php" class="lobby-chat-row" style="margin-top:8px;" data-ajax="true">
        <input type="hidden" name="action" value="sendChat">
        <input name="message" class="field" type="text" autocomplete="off" placeholder="<?= tr('chat_placeholder') ?>">
        <button type="submit" class="btn btn-neutral">💬 <?= tr('lobby_chat_send') ?></button>
    </form>
</section>

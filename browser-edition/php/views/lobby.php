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
$connection = $bundle['connection'] ?? [];
$diagnostics = $bundle['diagnostics'] ?? [];
$inviteKey = (string) ($session['invite_key'] ?? '');
$connectionStatus = (string) ($connection['status'] ?? ($bundle['mode'] ?? 'idle'));
$roleLabel = trim((string) ($session['role'] ?? ''));
?>
<section class="panel lobby-panel">
    <h2><?= tr('lobby_title') ?></h2>

    <div class="lobby-meta">
        <div><strong><?= htmlspecialchars($mode) ?></strong></div>
        <div class="invite-row">
            <span><?= tr('lobby_invite') ?></span>:
            <code><?= htmlspecialchars($inviteKey !== '' ? $inviteKey : '-') ?></code>
            <?php if ($inviteKey !== ''): ?>
                <button type="button" class="btn btn-neutral btn-copy" data-copy-text="<?= htmlspecialchars($inviteKey) ?>"><?= tr('invite_copy') ?></button>
            <?php endif; ?>
        </div>
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
                    $status = (string) ($slotState['status'] ?? ($connected ? 'occupied_online' : (trim($slotName) === '' ? 'empty' : 'occupied_offline')));
                    $displayName = trim((string) ($slotState['name'] ?? $slotName)) !== '' ? (string) ($slotState['name'] ?? $slotName) : tr('lobby_slot_empty');
                    ?>
                    <div class="lobby-slot">
                        <div class="top">
                            <strong>Slot <?= $idx + 1 ?></strong>
                            <span><?= htmlspecialchars($displayName) ?></span>
                        </div>
                        <div class="roles">
                            <span><?= htmlspecialchars(tr('slot_status_' . $status)) ?></span>
                            <?php if ($isLocalSeat): ?><span><?= tr('slot_you') ?></span><?php endif; ?>
                            <?php if ($isHostSeat): ?><span><?= tr('slot_host') ?></span><?php endif; ?>
                            <span><?= $connected ? tr('slot_online') : tr('slot_offline') ?></span>
                            <?php if ($isProvisionalCPU): ?><span><?= tr('slot_cpu') ?></span><?php endif; ?>
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
            <h3><?= tr('connection_title') ?></h3>
            <div class="connection-grid">
                <div><span><?= tr('connection_status') ?></span><strong><?= htmlspecialchars($connectionStatus) ?></strong></div>
                <div><span><?= tr('connection_mode') ?></span><strong><?= !empty($connection['is_online']) ? tr('connection_online') : tr('connection_offline') ?></strong></div>
                <?php if ($roleLabel !== ''): ?>
                    <div><span><?= tr('connection_role') ?></span><strong><?= htmlspecialchars($roleLabel) ?></strong></div>
                <?php endif; ?>
                <div><span><?= tr('connection_backlog') ?></span><strong><?= (int) ($diagnostics['event_backlog'] ?? 0) ?></strong></div>
                <?php if (!empty($connection['last_error']['message'])): ?>
                    <div class="connection-error"><span><?= tr('connection_error') ?></span><strong><?= htmlspecialchars((string) $connection['last_error']['message']) ?></strong></div>
                <?php endif; ?>
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
                        $label .= htmlspecialchars(formatEventLine($ev));
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

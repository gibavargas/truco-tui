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
$network = $connection['network'] ?? [];
$inviteKey = (string) ($session['invite_key'] ?? '');
$connectionStatus = (string) ($connection['status'] ?? ($bundle['mode'] ?? 'idle'));
$roleLabel = trim((string) ($session['role'] ?? ''));
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
?>
<section class="panel lobby-panel">
    <div class="section-head">
        <div>
            <span class="section-kicker"><?= tr('setup_online_title') ?></span>
            <h2><?= tr('lobby_title') ?></h2>
        </div>
        <strong class="mode-pill"><?= htmlspecialchars($mode) ?></strong>
    </div>

    <div class="invite-stage">
        <div class="invite-copy">
            <span class="invite-label"><?= tr('lobby_invite') ?></span>
            <code><?= htmlspecialchars($inviteKey !== '' ? $inviteKey : '-') ?></code>
            <p><?= tr('lobby_invite_hint') ?></p>
        </div>
        <div class="invite-actions">
            <?php if ($inviteKey !== ''): ?>
                <button type="button" class="btn btn-neutral btn-copy" data-copy-text="<?= htmlspecialchars($inviteKey) ?>"><?= tr('invite_copy') ?></button>
            <?php endif; ?>
            <?php if ($isHost): ?>
                <form method="post" action="index.php" data-ajax="true">
                    <input type="hidden" name="action" value="startOnlineMatch">
                    <button type="submit" class="btn btn-primary">▶ <?= tr('lobby_start') ?></button>
                </form>
            <?php endif; ?>
        </div>
    </div>

    <div class="lobby-grid">
        <article class="lobby-block lobby-tableau">
            <div class="block-head">
                <h3><?= tr('lobby_slots_title') ?></h3>
                <span class="table-tag"><?= tr('lobby_slots_count', count($slots)) ?></span>
            </div>
            <div class="lobby-slots">
                <?php foreach ($slots as $idx => $slotName): ?>
                    <?php
                    $slotState = $slotStates[$idx] ?? [];
                    $connected = (bool) ($slotState['is_connected'] ?? false);
                    $isHostSeat = (bool) ($slotState['is_host'] ?? false);
                    $isLocalSeat = (bool) ($slotState['is_local'] ?? false);
                    $canVote = (bool) ($slotState['can_vote_host'] ?? false);
                    $canReplace = (bool) ($slotState['can_request_replacement'] ?? false);
                    $isProvisionalCPU = (bool) ($slotState['is_provisional_cpu'] ?? false);
                    $status = (string) ($slotState['status'] ?? 'empty');
                    $displayName = trim((string) ($slotState['name'] ?? $slotName)) !== '' ? (string) ($slotState['name'] ?? $slotName) : tr('lobby_slot_empty');
                    ?>
                    <article class="lobby-slot <?= $isLocalSeat ? 'local-seat' : '' ?>">
                        <div class="slot-crest">
                            <span class="slot-index">0<?= $idx + 1 ?></span>
                            <strong><?= htmlspecialchars($displayName) ?></strong>
                        </div>
                        <div class="slot-tags roles">
                            <span><?= htmlspecialchars(tr('slot_status_' . $status)) ?></span>
                            <?php if ($isLocalSeat): ?><span><?= tr('slot_you') ?></span><?php endif; ?>
                            <?php if ($isHostSeat): ?><span><?= tr('slot_host') ?></span><?php endif; ?>
                            <span><?= $connected ? tr('slot_online') : tr('slot_offline') ?></span>
                            <?php if ($isProvisionalCPU): ?><span><?= tr('slot_cpu') ?></span><?php endif; ?>
                        </div>
                        <div class="slot-actions roles">
                            <?php if ($canVote): ?>
                                <form method="post" action="index.php" data-ajax="true">
                                    <input type="hidden" name="action" value="voteHost">
                                    <input type="hidden" name="slot" value="<?= htmlspecialchars($idx) ?>">
                                    <button type="submit" class="btn btn-neutral"><?= tr('action_vote_host') ?></button>
                                </form>
                            <?php endif; ?>
                            <?php if ($canReplace): ?>
                                <form method="post" action="index.php" data-ajax="true">
                                    <input type="hidden" name="action" value="requestReplacementInvite">
                                    <input type="hidden" name="slot" value="<?= htmlspecialchars($idx) ?>">
                                    <button type="submit" class="btn btn-truco"><?= tr('action_replacement_invite') ?></button>
                                </form>
                            <?php endif; ?>
                        </div>
                    </article>
                <?php endforeach; ?>
            </div>
        </article>

        <article class="lobby-block lobby-side-stack">
            <div class="lobby-mini-grid">
                <section class="side-block">
                    <h3><?= tr('connection_title') ?></h3>
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

                <section class="side-block side-log">
                    <h3><?= tr('lobby_events_title') ?></h3>
                    <pre id="lobby-events" class="event-feed" data-focus-target="lobby-events" tabindex="-1" role="log" aria-live="polite" aria-atomic="false"><?php
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
                </section>
            </div>

            <section class="side-block chat-block">
                <h3><?= tr('lobby_chat_send') ?></h3>
                <form method="post" action="index.php" class="lobby-chat-row" data-ajax="true">
                    <input type="hidden" name="action" value="sendChat">
                    <input name="message" class="field" type="text" autocomplete="off" placeholder="<?= htmlspecialchars(tr('chat_placeholder')) ?>">
                    <button type="submit" class="btn btn-neutral">💬 <?= tr('lobby_chat_send') ?></button>
                </form>
            </section>
        </article>
    </div>

    <div class="action-row lobby-actions">
        <form method="post" action="index.php" data-ajax="true">
            <input type="hidden" name="action" value="refreshLobby">
            <button type="submit" class="btn btn-neutral">⟳ <?= tr('lobby_refresh') ?></button>
        </form>
        <form method="post" action="index.php" data-ajax="true">
            <input type="hidden" name="action" value="leaveLobby">
            <button type="submit" class="btn btn-refuse">⎋ <?= tr('lobby_leave') ?></button>
        </form>
    </div>
</section>

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
$network = $bundle['connection']['network'] ?? [];
$seatProtocolVersions = $network['seat_protocol_versions'] ?? [];
$supportedVersions = array_map(
    static fn($version) => 'v' . (int) $version,
    is_array($network['supported_protocol_versions'] ?? null) ? $network['supported_protocol_versions'] : []
);
$supportedVersionsText = $supportedVersions ? implode('/', $supportedVersions) : '-';
$transportKey = 'network_transport_' . ($network['transport'] ?? 'tcp_tls');
$transportLabel = tr($transportKey);
$compatibilityText = '';
if ($isHost) {
    $uniqueVersions = [];
    foreach ($seatProtocolVersions as $version) {
        if ((int) $version > 0) {
            $uniqueVersions['v' . (int) $version] = true;
        }
    }
    $hostVersionsText = $uniqueVersions ? implode('/', array_keys($uniqueVersions)) : $supportedVersionsText;
    $compatibilityText = !empty($network['mixed_protocol_session'])
        ? tr('network_compatibility_host_mixed', $hostVersionsText)
        : tr('network_compatibility_host_uniform', $hostVersionsText);
} elseif (!empty($network['negotiated_protocol_version'])) {
    $compatibilityText = tr('network_compatibility_client', 'v' . (int) $network['negotiated_protocol_version']);
}
?>
<section class="panel lobby-panel">
    <h2><?= tr('lobby_title') ?></h2>

    <div class="lobby-meta">
        <div><strong><?= htmlspecialchars($mode) ?></strong></div>
        <div><span><?= tr('lobby_invite') ?></span>: <code><?= htmlspecialchars($session['invite_key'] ?? '-') ?></code></div>
        <div><span><?= tr('network_transport_label') ?></span>: <strong><?= htmlspecialchars($transportLabel) ?></strong></div>
        <div><span><?= tr('network_compatibility_label') ?></span>: <strong><?= htmlspecialchars($compatibilityText !== '' ? $compatibilityText : $supportedVersionsText) ?></strong></div>
        <div><span><?= tr('network_supported_versions_label') ?></span>: <code><?= htmlspecialchars($supportedVersionsText) ?></code></div>
    </div>

    <div class="lobby-grid">
        <div class="lobby-block">
            <h3><?= tr('lobby_slots_title') ?></h3>
            <div class="lobby-slots">
                <?php foreach ($slots as $idx => $slotName): ?>
                    <?php
                    $connected = !empty($connectedSeats[(string) $idx]) || !empty($connectedSeats[$idx]);
                    $displayName = trim($slotName) !== '' ? $slotName : tr('lobby_slot_empty');
                    $seatProtocolVersion = $seatProtocolVersions[(string) $idx] ?? $seatProtocolVersions[$idx] ?? null;
                    ?>
                    <div class="lobby-slot">
                        <div class="top">
                            <strong>Slot <?= $idx + 1 ?></strong>
                            <span>
                                <?= htmlspecialchars($displayName) ?>
                                <?php if ($seatProtocolVersion !== null): ?>
                                    <small>· v<?= (int) $seatProtocolVersion ?></small>
                                <?php endif; ?>
                            </span>
                        </div>
                        <div class="roles">
                            <?php if ($idx === $assignedSeat): ?><span>you</span><?php endif; ?>
                            <?php if ($idx === $hostSeat): ?><span>host</span><?php endif; ?>
                            <span><?= $connected ? 'online' : 'offline' ?></span>
                        </div>
                        <div class="roles">
                            <?php if (trim($slotName) !== '' && $idx !== $assignedSeat): ?>
                                <form method="post" action="index.php" data-ajax="true">
                                    <input type="hidden" name="action" value="voteHost">
                                    <input type="hidden" name="slot" value="<?= $idx ?>">
                                    <button type="submit" class="btn btn-neutral"><?= tr('action_vote_host') ?></button>
                                </form>
                            <?php endif; ?>
                            <?php if ($isHost && trim($slotName) === ''): ?>
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

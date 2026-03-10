<?php /** Lobby view — Online lobby */ ?>
<section class="panel lobby-panel">
    <h2>
        <?= tr('lobby_title') ?>
    </h2>

    <?php
    $session = $_SESSION['online_session'] ?? null;
    $events = $_SESSION['online_events'] ?? [];
    ?>

    <?php if ($session): ?>
        <div class="lobby-meta">
            <div><strong>
                    <?= htmlspecialchars($session['mode'] ?? 'host') ?>
                </strong></div>
            <div><span>
                    <?= tr('lobby_invite') ?>
                </span>: <code><?= htmlspecialchars($session['inviteKey'] ?? '-') ?></code></div>
        </div>

        <div class="lobby-grid">
            <div class="lobby-block">
                <h3>
                    <?= tr('lobby_slots_title') ?>
                </h3>
                <div class="lobby-slots">
                    <?php foreach (($session['slots'] ?? []) as $idx => $slotName): ?>
                        <div class="lobby-slot">
                            <div class="top">
                                <strong>Slot
                                    <?= $idx + 1 ?>
                                </strong>
                                <span>
                                    <?= htmlspecialchars(trim($slotName) ?: tr('lobby_slot_empty')) ?>
                                </span>
                            </div>
                        </div>
                    <?php endforeach; ?>
                </div>
            </div>
            <div class="lobby-block">
                <h3>
                    <?= tr('lobby_events_title') ?>
                </h3>
                <pre><?= htmlspecialchars(count($events) ? implode("\n", $events) : tr('lobby_events_empty')) ?></pre>
            </div>
        </div>
    <?php endif; ?>

    <div class="action-row" style="margin-top:12px;">
        <form method="post" action="index.php" style="display:inline">
            <input type="hidden" name="action" value="startOnlineMatch">
            <button type="submit" class="btn btn-primary">▶
                <?= tr('lobby_start') ?>
            </button>
        </form>
        <form method="post" action="index.php" style="display:inline">
            <input type="hidden" name="action" value="refreshLobby">
            <button type="submit" class="btn btn-neutral">⟳
                <?= tr('lobby_refresh') ?>
            </button>
        </form>
        <form method="post" action="index.php" style="display:inline">
            <input type="hidden" name="action" value="leaveLobby">
            <button type="submit" class="btn btn-refuse">⎋
                <?= tr('lobby_leave') ?>
            </button>
        </form>
    </div>

    <!-- Chat -->
    <form method="post" action="index.php" class="lobby-chat-row" style="margin-top:8px;">
        <input type="hidden" name="action" value="sendChat">
        <input name="message" class="field" type="text" autocomplete="off" placeholder="Mensagem">
        <button type="submit" class="btn btn-neutral">💬
            <?= tr('lobby_chat_send') ?>
        </button>
    </form>
</section>
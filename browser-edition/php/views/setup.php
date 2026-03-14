<?php /** Setup page — New match form */ ?>
<section class="panel setup-panel">
    <h2>
        <?= tr('setup_title') ?>
    </h2>
    <form method="post" action="index.php" data-ajax="true">
        <input type="hidden" name="action" value="startGame">
        <div class="setup-grid">
            <div>
                <label for="player-name">
                    <?= tr('setup_name') ?>
                </label>
                <input id="player-name" name="name" class="field" type="text"
                    value="<?= htmlspecialchars($_SESSION['player_name'] ?? 'Você') ?>" autocomplete="off">
            </div>
            <div>
                <label for="num-players">
                    <?= tr('setup_players') ?>
                </label>
                <select id="num-players" name="numPlayers" class="field">
                    <option value="2">2</option>
                    <option value="4">4</option>
                </select>
            </div>
        </div>
        <div class="setup-actions">
            <button type="submit" class="btn btn-primary">▶
                <?= tr('setup_start') ?>
            </button>
        </div>
        <p class="setup-help">
            <?= tr('setup_help') ?>
        </p>
    </form>

    <div class="online-setup">
        <h3><?= tr('lobby_title') ?></h3>
        <div class="setup-grid online-grid">
            <form method="post" action="index.php" data-ajax="true">
                <input type="hidden" name="action" value="startOnlineHost">
                <input type="hidden" name="name" value="<?= htmlspecialchars($_SESSION['player_name'] ?? 'Você') ?>">
                <input type="hidden" name="numPlayers" value="2">
                <button type="submit" class="btn btn-neutral"><?= tr('online_action_host') ?> · 2P</button>
            </form>
            <form method="post" action="index.php" data-ajax="true">
                <input type="hidden" name="action" value="startOnlineHost">
                <input type="hidden" name="name" value="<?= htmlspecialchars($_SESSION['player_name'] ?? 'Você') ?>">
                <input type="hidden" name="numPlayers" value="4">
                <button type="submit" class="btn btn-neutral"><?= tr('online_action_host') ?> · 4P</button>
            </form>
        </div>
        <form method="post" action="index.php" data-ajax="true" style="margin-top:10px;">
            <input type="hidden" name="action" value="joinOnline">
            <div class="setup-grid">
                <div>
                    <label for="invite-key"><?= tr('setup_online_key') ?></label>
                    <input id="invite-key" name="key" class="field" type="text" autocomplete="off">
                </div>
                <div>
                    <label for="join-role"><?= tr('setup_online_role') ?></label>
                    <select id="join-role" name="role" class="field">
                        <option value="auto"><?= tr('online_role_auto') ?></option>
                        <option value="partner"><?= tr('online_role_partner') ?></option>
                        <option value="opponent"><?= tr('online_role_opponent') ?></option>
                    </select>
                </div>
            </div>
            <input type="hidden" name="name" value="<?= htmlspecialchars($_SESSION['player_name'] ?? 'Você') ?>">
            <div class="setup-actions">
                <button type="submit" class="btn btn-truco"><?= tr('online_action_join') ?></button>
            </div>
        </form>
    </div>

    <!-- Language switcher -->
    <form method="post" action="index.php" style="margin-top:12px;">
        <input type="hidden" name="action" value="setLocale">
        <label for="locale-select">
            <?= tr('locale_label') ?>
        </label>
        <select id="locale-select" name="locale" class="field field-sm" onchange="this.form.submit()">
            <option value="pt-BR" <?= ($_SESSION['locale'] ?? 'pt-BR') === 'pt-BR' ? 'selected' : '' ?>>Português (BR)
            </option>
            <option value="en-US" <?= ($_SESSION['locale'] ?? 'pt-BR') === 'en-US' ? 'selected' : '' ?>>English (US)
            </option>
        </select>
        <noscript><button type="submit" class="btn btn-neutral">OK</button></noscript>
    </form>
</section>

<?php /** Setup page — New match form */ ?>
<section class="panel setup-panel">
    <h2>
        <?= tr('setup_title') ?>
    </h2>
    <div class="setup-sections">
        <div class="setup-card">
            <h3><?= tr('setup_offline_title') ?></h3>
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
            </form>
        </div>

        <div class="setup-card online-setup">
            <h3><?= tr('setup_online_title') ?></h3>
            <form method="post" action="index.php" data-ajax="true">
                <input type="hidden" name="action" value="startOnlineHost">
                <input type="hidden" name="name" value="<?= htmlspecialchars($_SESSION['player_name'] ?? 'Você') ?>">
                <div class="setup-grid host-grid">
                    <div>
                        <label for="host-num-players"><?= tr('setup_online_players') ?></label>
                        <select id="host-num-players" name="numPlayers" class="field">
                            <option value="2">2</option>
                            <option value="4">4</option>
                        </select>
                    </div>
                    <div class="span-2">
                        <label for="relay-url"><?= tr('setup_online_relay') ?></label>
                        <input id="relay-url" name="relay_url" class="field" type="text" autocomplete="off" placeholder="https://relay.example.com">
                    </div>
                </div>
                <div class="setup-actions">
                    <button type="submit" class="btn btn-neutral"><?= tr('online_action_host') ?></button>
                </div>
            </form>
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
    <p class="setup-help">
        <?= tr('setup_help') ?>
    </p>
</section>

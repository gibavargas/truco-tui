<?php
$defaultName = htmlspecialchars($_SESSION['player_name'] ?? 'Você');
?>
<section class="panel setup-panel">
    <div class="setup-shell">
        <div class="setup-hero">
            <span class="setup-kicker"><?= tr('setup_kicker') ?></span>
            <h2><?= tr('setup_intro_title') ?></h2>
            <p class="setup-intro"><?= tr('setup_intro_body') ?></p>
            <div class="setup-hero-pills">
                <span>2p / 4p</span>
                <span>P2P + Relay</span>
                <span><?= tr('setup_live_sync') ?></span>
            </div>
        </div>

        <div class="setup-card-grid">
            <article class="setup-card setup-card-offline">
                <div class="setup-card-head">
                    <span class="setup-card-kicker"><?= tr('setup_mode_offline') ?></span>
                    <h3><?= tr('setup_offline_title') ?></h3>
                    <p><?= tr('setup_offline_body') ?></p>
                </div>
                <form method="post" action="index.php" data-ajax="true">
                    <input type="hidden" name="action" value="startGame">
                    <div class="setup-form-grid">
                        <div>
                            <label for="player-name"><?= tr('setup_name') ?></label>
                            <input id="player-name" name="name" class="field" type="text" value="<?= $defaultName ?>" autocomplete="off">
                            <p class="field-help"><?= tr('setup_name_hint') ?></p>
                        </div>
                        <div>
                            <label for="num-players"><?= tr('setup_players') ?></label>
                            <select id="num-players" name="numPlayers" class="field">
                                <option value="2">2</option>
                                <option value="4">4</option>
                            </select>
                        </div>
                    </div>
                    <div class="setup-actions">
                        <button type="submit" class="btn btn-primary">▶ <?= tr('setup_start') ?></button>
                    </div>
                    <p class="setup-help"><?= tr('setup_help') ?></p>
                </form>
            </article>

            <article class="setup-card setup-card-host">
                <div class="setup-card-head">
                    <span class="setup-card-kicker"><?= tr('setup_mode_online') ?></span>
                    <h3><?= tr('setup_host_title') ?></h3>
                    <p><?= tr('setup_host_body') ?></p>
                </div>
                <form method="post" action="index.php" data-ajax="true">
                    <input type="hidden" name="action" value="startOnlineHost">
                    <div class="setup-form-grid">
                        <div>
                            <label for="host-name"><?= tr('setup_name') ?></label>
                            <input id="host-name" name="name" class="field" type="text" value="<?= $defaultName ?>" autocomplete="off">
                            <p class="field-help"><?= tr('setup_name_hint') ?></p>
                        </div>
                        <div>
                            <label for="host-num-players"><?= tr('setup_players') ?></label>
                            <select id="host-num-players" name="numPlayers" class="field">
                                <option value="2">2</option>
                                <option value="4">4</option>
                            </select>
                        </div>
                    </div>
                    <div class="setup-actions">
                        <button type="submit" class="btn btn-neutral">⌁ <?= tr('setup_host_start') ?></button>
                    </div>
                    <p class="setup-help"><?= tr('setup_host_hint') ?></p>
                </form>
            </article>

            <article class="setup-card setup-card-join">
                <div class="setup-card-head">
                    <span class="setup-card-kicker"><?= tr('lobby_title') ?></span>
                    <h3><?= tr('setup_join_title') ?></h3>
                    <p><?= tr('setup_join_body') ?></p>
                </div>
                <form method="post" action="index.php" data-ajax="true">
                    <input type="hidden" name="action" value="joinOnline">
                    <div class="setup-form-grid">
                        <div>
                            <label for="join-name"><?= tr('setup_name') ?></label>
                            <input id="join-name" name="name" class="field" type="text" value="<?= $defaultName ?>" autocomplete="off">
                            <p class="field-help"><?= tr('setup_name_hint') ?></p>
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
                    <div class="setup-form-grid setup-form-grid-single">
                        <div>
                            <label for="invite-key"><?= tr('setup_online_key') ?></label>
                            <input id="invite-key" name="key" class="field" type="text" autocomplete="off" placeholder="<?= tr('setup_join_key_placeholder') ?>">
                            <p class="field-help"><?= tr('setup_join_hint') ?></p>
                        </div>
                    </div>
                    <div class="setup-actions">
                        <button type="submit" class="btn btn-truco">⇢ <?= tr('setup_join_start') ?></button>
                    </div>
                </form>
            </article>
        </div>
    </div>
</section>

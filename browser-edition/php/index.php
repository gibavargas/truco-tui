<?php
/**
 * Truco Browser Edition — PHP Frontend (no JavaScript)
 * Main router: handles form POSTs, calls Go API, renders views.
 */

session_start();

require_once __DIR__ . '/api_client.php';
require_once __DIR__ . '/i18n.php';

// Defaults
if (!isset($_SESSION['locale'])) {
    $_SESSION['locale'] = 'pt-BR';
}

$apiUrl = getenv('TRUCO_API_URL') ?: 'http://localhost:9090';
$api = new TrucoApiClient($apiUrl);

// Ensure we have a Go API session
if (empty($_SESSION['api_session_id'])) {
    $res = $api->call('createSession');
    if (!empty($res['ok']) && !empty($res['sessionId'])) {
        $_SESSION['api_session_id'] = $res['sessionId'];
    }
}
$sid = $_SESSION['api_session_id'] ?? '';

// State tracking
$view = $_SESSION['current_view'] ?? 'setup'; // setup | game | lobby
$snap = null;
$errorMsg = '';
$statusMsg = '';

// ---------------------------------------------------------------------------
// Handle POST actions
// ---------------------------------------------------------------------------
if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $action = $_POST['action'] ?? '';

    switch ($action) {
        case 'setLocale':
            $loc = $_POST['locale'] ?? 'pt-BR';
            $_SESSION['locale'] = in_array($loc, ['pt-BR', 'en-US']) ? $loc : 'pt-BR';
            break;

        case 'startGame':
            $name = trim($_POST['name'] ?? 'Você');
            $numPlayers = (int) ($_POST['numPlayers'] ?? 2);
            $_SESSION['player_name'] = $name;
            $res = $api->call('startGame', $sid, [
                'numPlayers' => $numPlayers,
                'name' => $name,
            ]);
            if (!empty($res['ok'])) {
                $view = 'game';
                $_SESSION['current_view'] = 'game';
                // Run CPU turns immediately
                $api->call('autoCpuLoopTick', $sid);
            } else {
                $errorMsg = $res['error'] ?? 'Failed to start game';
            }
            break;

        case 'play':
            $idx = (int) ($_POST['cardIndex'] ?? 0);
            $res = $api->call('play', $sid, ['cardIndex' => $idx]);
            if (empty($res['ok'])) {
                $errorMsg = $res['error'] ?? 'Failed to play card';
            }
            // Run CPU turns after player action
            $api->call('autoCpuLoopTick', $sid);
            break;

        case 'truco':
            $res = $api->call('truco', $sid);
            if (empty($res['ok'])) {
                $errorMsg = $res['error'] ?? 'Truco failed';
            }
            $api->call('autoCpuLoopTick', $sid);
            break;

        case 'accept':
            $res = $api->call('accept', $sid);
            if (empty($res['ok'])) {
                $errorMsg = $res['error'] ?? 'Accept failed';
            }
            $api->call('autoCpuLoopTick', $sid);
            break;

        case 'refuse':
            $res = $api->call('refuse', $sid);
            if (empty($res['ok'])) {
                $errorMsg = $res['error'] ?? 'Refuse failed';
            }
            $api->call('autoCpuLoopTick', $sid);
            break;

        case 'autoCpuAndRefresh':
            $api->call('autoCpuLoopTick', $sid);
            break;

        case 'newHand':
            $api->call('newHand', $sid);
            $api->call('autoCpuLoopTick', $sid);
            break;

        case 'reset':
            $api->call('reset', $sid);
            $view = 'setup';
            $_SESSION['current_view'] = 'setup';
            break;

        // Online lobby actions
        case 'startOnlineHost':
            $name = trim($_POST['name'] ?? 'Host');
            $numPlayers = (int) ($_POST['numPlayers'] ?? 2);
            $res = $api->call('startOnlineHost', $sid, [
                'name' => $name,
                'numPlayers' => $numPlayers,
            ]);
            if (!empty($res['ok']) && !empty($res['session'])) {
                $_SESSION['online_session'] = $res['session'];
                $view = 'lobby';
                $_SESSION['current_view'] = 'lobby';
            } else {
                $errorMsg = $res['error'] ?? 'Failed to create lobby';
            }
            break;

        case 'joinOnline':
            $name = trim($_POST['name'] ?? 'Player');
            $key = trim($_POST['key'] ?? '');
            $role = $_POST['role'] ?? 'auto';
            $numPlayers = (int) ($_POST['numPlayers'] ?? 2);
            $res = $api->call('joinOnline', $sid, [
                'name' => $name,
                'key' => $key,
                'role' => $role,
                'numPlayers' => $numPlayers,
            ]);
            if (!empty($res['ok']) && !empty($res['session'])) {
                $_SESSION['online_session'] = $res['session'];
                $view = 'lobby';
                $_SESSION['current_view'] = 'lobby';
            } else {
                $errorMsg = $res['error'] ?? 'Failed to join lobby';
            }
            break;

        case 'startOnlineMatch':
            $res = $api->call('startOnlineMatch', $sid);
            if (!empty($res['ok'])) {
                if (!empty($res['session'])) {
                    $_SESSION['online_session'] = $res['session'];
                }
                $view = 'game';
                $_SESSION['current_view'] = 'game';
                $api->call('autoCpuLoopTick', $sid);
            } else {
                $errorMsg = $res['error'] ?? 'Failed to start match';
            }
            break;

        case 'refreshLobby':
            $res = $api->call('onlineState', $sid);
            if (!empty($res['ok']) && !empty($res['session'])) {
                $_SESSION['online_session'] = $res['session'];
            }
            $evRes = $api->call('pullOnlineEvents', $sid);
            if (!empty($evRes['ok']) && !empty($evRes['events'])) {
                $trail = $_SESSION['online_events'] ?? [];
                foreach ($evRes['events'] as $ev) {
                    $ts = date('H:i', (int) (($ev['timestamp'] ?? 0) / 1000));
                    $trail[] = "[{$ts}] " . ($ev['type'] ?? 'event') . ' · ' . ($ev['message'] ?? '');
                }
                $_SESSION['online_events'] = array_slice($trail, -24);
            }
            break;

        case 'sendChat':
            $msg = trim($_POST['message'] ?? '');
            if ($msg !== '') {
                $api->call('sendChat', $sid, ['message' => $msg]);
            }
            // Refresh events
            $evRes = $api->call('pullOnlineEvents', $sid);
            if (!empty($evRes['ok']) && !empty($evRes['events'])) {
                $trail = $_SESSION['online_events'] ?? [];
                foreach ($evRes['events'] as $ev) {
                    $ts = date('H:i', (int) (($ev['timestamp'] ?? 0) / 1000));
                    $trail[] = "[{$ts}] " . ($ev['type'] ?? 'event') . ' · ' . ($ev['message'] ?? '');
                }
                $_SESSION['online_events'] = array_slice($trail, -24);
            }
            break;

        case 'leaveLobby':
            $api->call('leaveSession', $sid);
            unset($_SESSION['online_session'], $_SESSION['online_events']);
            $view = 'setup';
            $_SESSION['current_view'] = 'setup';
            break;
    }

    // Post-redirect-get (PRG): redirect to prevent form re-submission
    header('Location: index.php');
    exit;
}

// ---------------------------------------------------------------------------
// Fetch current snapshot for game view
// ---------------------------------------------------------------------------
if ($view === 'game') {
    $res = $api->call('snapshot', $sid);
    if (!empty($res['ok'])) {
        $snap = TrucoApiClient::parseSnapshot($res);
    }
    if ($snap === null) {
        // Game was reset or errored — go back to setup
        $view = 'setup';
        $_SESSION['current_view'] = 'setup';
    }
}

// Store flash error if any
$errorMsg = $_SESSION['flash_error'] ?? '';
unset($_SESSION['flash_error']);

$locale = $_SESSION['locale'] ?? 'pt-BR';
$langAttr = ($locale === 'en-US') ? 'en' : 'pt-BR';
?>
<!doctype html>
<html lang="<?= $langAttr ?>">

<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <meta http-equiv="Cache-Control" content="no-cache, no-store, must-revalidate">
    <title>
        <?= tr('title_main') ?>
    </title>
    <meta name="description" content="Truco Paulista — PHP edition, no JavaScript.">
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link
        href="https://fonts.googleapis.com/css2?family=Bebas+Neue&family=Sora:wght@400;600;700;800&family=IBM+Plex+Mono:wght@400;600&display=swap"
        rel="stylesheet">
    <link rel="stylesheet" href="style.css">
</head>

<body>
    <div class="ambient ambient-a"></div>
    <div class="ambient ambient-b"></div>

    <main class="app-shell">
        <header class="topbar">
            <div class="brand">
                <span class="brand-mark">🂡</span>
                <div>
                    <h1>
                        <?= tr('title_main') ?>
                    </h1>
                    <p>
                        <?= tr('title_sub') ?>
                    </p>
                </div>
            </div>
            <div class="topbar-actions">
                <form method="post" action="index.php" style="display:inline">
                    <input type="hidden" name="action" value="setLocale">
                    <label for="locale-sel">
                        <?= tr('locale_label') ?>
                    </label>
                    <select id="locale-sel" name="locale" class="field field-sm">
                        <option value="pt-BR" <?= $locale === 'pt-BR' ? 'selected' : '' ?>>Português (BR)</option>
                        <option value="en-US" <?= $locale === 'en-US' ? 'selected' : '' ?>>English (US)</option>
                    </select>
                    <button type="submit" class="btn btn-neutral" style="padding:6px 10px;">OK</button>
                </form>
            </div>
        </header>

        <?php if ($view === 'game' && $snap !== null): ?>
            <?php include __DIR__ . '/views/game.php'; ?>
        <?php elseif ($view === 'lobby'): ?>
            <?php include __DIR__ . '/views/lobby.php'; ?>
        <?php else: ?>
            <?php include __DIR__ . '/views/setup.php'; ?>
        <?php endif; ?>

        <?php if ($errorMsg): ?>
            <div class="error-log" style="margin-top:12px; padding:10px;">
                <?= htmlspecialchars($errorMsg) ?>
            </div>
        <?php endif; ?>
    </main>
</body>

</html>
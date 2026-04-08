<?php
// Secure session management
$isSecure = isset($_SERVER['HTTPS']) && $_SERVER['HTTPS'] === 'on';
session_set_cookie_params([
    'lifetime' => 0,
    'path' => '/',
    'domain' => '',
    'secure' => $isSecure,
    'httponly' => true,
    'samesite' => 'Lax'
]);
session_start();

require_once __DIR__ . '/api_client.php';
require_once __DIR__ . '/i18n.php';

if (!isset($_SESSION['locale'])) {
    $_SESSION['locale'] = 'pt-BR';
}

$apiUrl = getenv('TRUCO_API_URL') ?: 'http://localhost:9090';
$api = new TrucoApiClient($apiUrl);

function ensureApiSession(TrucoApiClient $api): string
{
    if (empty($_SESSION['api_session_id'])) {
        $res = $api->call('createSession');
        if (!empty($res['ok']) && !empty($res['sessionId'])) {
            $_SESSION['api_session_id'] = $res['sessionId'];
        }
    }

    return (string) ($_SESSION['api_session_id'] ?? '');
}

$sid = ensureApiSession($api);

function storeBrowserState(array $result): void
{
    $bundle = TrucoApiClient::parseBundle($result);
    if (is_array($bundle)) {
        $_SESSION['runtime_bundle'] = $bundle;
        $_SESSION['online_session'] = $result['session'] ?? [];
    }
    $_SESSION['runtime_state_valid'] = true;
    unset($_SESSION['runtime_error']);
    if (!empty($result['events']) && is_array($result['events'])) {
        $trail = $_SESSION['runtime_events'] ?? [];
        foreach ($result['events'] as $ev) {
            $trail[] = $ev;
        }
        $_SESSION['runtime_events'] = array_slice($trail, -80);
    }
}

function markBrowserRuntimeError(string $message, bool $stale = false): void
{
    $_SESSION['runtime_error'] = trim($message);
    $_SESSION['runtime_state_valid'] = !$stale;
}

function refreshRuntimeState(TrucoApiClient $api, string $sid, bool $pullEvents = false): ?array
{
    $res = $api->call('snapshot', $sid);
    if (empty($res['ok'])) {
        markBrowserRuntimeError((string) ($res['error'] ?? tr('game_runtime_stale_copy')), true);
        return null;
    }
    storeBrowserState($res);
    if ($pullEvents) {
        $evRes = $api->call('pollEvents', $sid);
        if (!empty($evRes['ok'])) {
            storeBrowserState($evRes);
        }
    }
    return $_SESSION['runtime_bundle'] ?? null;
}

function currentViewFromBundle(?array $bundle): string
{
    $mode = $bundle['mode'] ?? 'idle';
    if (strpos($mode, 'lobby') !== false) {
        return 'lobby';
    }
    if (strpos($mode, 'match') !== false) {
        return 'game';
    }
    return 'setup';
}

function renderRuntimeNotice(): string
{
    $message = trim((string) ($_SESSION['runtime_error'] ?? ''));
    $stateValid = $_SESSION['runtime_state_valid'] ?? true;

    $isStale = !$stateValid;
    if ($message === '' && $stateValid) {
        return '<div id="runtime-notice" class="runtime-banner" data-focus-target="runtime-banner" tabindex="-1" role="status" aria-live="polite" aria-atomic="true"></div>';
    }

    $title = $isStale ? tr('game_runtime_stale_title') : tr('connection_error');
    if ($message === '') {
        $message = $isStale ? tr('game_runtime_stale_copy') : tr('connection_error');
    }

    $class = $isStale ? 'runtime-banner stale' : 'runtime-banner error';
    return '<div id="runtime-notice" class="' . $class . '" data-focus-target="runtime-banner" tabindex="-1" role="status" aria-live="polite" aria-atomic="true">'
        . '<strong>' . htmlspecialchars($title) . '</strong>'
        . '<span>' . htmlspecialchars($message) . '</span>'
        . '</div>';
}

function formatEventLine(array $ev): string
{
    $kind = (string) ($ev['kind'] ?? 'event');
    $payload = is_array($ev['payload'] ?? null) ? $ev['payload'] : [];
    $kindLabel = match ($kind) {
        'chat' => tr('event_chat'),
        'system' => tr('event_system'),
        'replacement_invite' => tr('event_replacement_invite'),
        'error' => tr('event_error'),
        'lobby_updated' => tr('event_lobby_updated'),
        'match_updated' => tr('event_match_updated'),
        default => $kind,
    };
    $parts = [$kindLabel];
    if (!empty($payload['author']) && !empty($payload['text'])) {
        $parts[] = $payload['author'] . ': ' . $payload['text'];
    } elseif (!empty($payload['text'])) {
        $parts[] = (string) $payload['text'];
    } elseif (!empty($payload['message'])) {
        $parts[] = (string) $payload['message'];
    } elseif (!empty($payload['invite_key'])) {
        $parts[] = (string) $payload['invite_key'];
    }
    return implode(' · ', $parts);
}

$errorMsg = '';
$ajaxRequest = !empty($_REQUEST['ajax']);

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $action = $_POST['action'] ?? '';

    switch ($action) {
        case 'setLocale':
            $loc = $_POST['locale'] ?? 'pt-BR';
            $_SESSION['locale'] = in_array($loc, ['pt-BR', 'en-US'], true) ? $loc : 'pt-BR';
            $res = $api->call('setLocale', $sid, ['locale' => $_SESSION['locale']]);
            if (!empty($res['ok'])) {
                storeBrowserState($res);
            } else {
                markBrowserRuntimeError((string) ($res['error'] ?? tr('connection_error')));
            }
            break;

        case 'startGame':
            $_SESSION['player_name'] = trim($_POST['name'] ?? tr('default_player_name'));
            $res = $api->call('startGame', $sid, [
                'numPlayers' => (int) ($_POST['numPlayers'] ?? 2),
                'name' => $_SESSION['player_name'],
            ]);
            if (!empty($res['ok'])) {
                storeBrowserState($res);
            } else {
                $errorMsg = $res['error'] ?? 'Failed to start game';
                markBrowserRuntimeError($errorMsg);
            }
            break;

        case 'play':
        case 'truco':
        case 'accept':
        case 'refuse':
            $payload = [];
            if ($action === 'play') {
                $payload['cardIndex'] = (int) ($_POST['cardIndex'] ?? -1);
                $payload['faceDown'] = !empty($_POST['faceDown']);
            }
            $res = $api->call($action, $sid, $payload);
            if (!empty($res['ok'])) {
                storeBrowserState($res);
                refreshRuntimeState($api, $sid, true);
            } else {
                $errorMsg = $res['error'] ?? 'Action failed';
                markBrowserRuntimeError($errorMsg);
            }
            break;

        case 'reset':
            $res = $api->call('reset', $sid);
            if (!empty($res['ok'])) {
                storeBrowserState($res);
                unset($_SESSION['runtime_events']);
            } else {
                $errorMsg = $res['error'] ?? 'Failed to reset session';
                markBrowserRuntimeError($errorMsg);
            }
            break;

        case 'leaveLobby':
            $res = $api->call('closeSession', $sid);
            if (!empty($res['ok'])) {
                unset($_SESSION['api_session_id']);
                unset($_SESSION['runtime_bundle']);
                unset($_SESSION['online_session']);
                unset($_SESSION['runtime_events']);
                unset($_SESSION['runtime_error']);
                $_SESSION['runtime_state_valid'] = true;
                $sid = '';
            } else {
                $errorMsg = $res['error'] ?? 'Failed to close session';
                markBrowserRuntimeError($errorMsg);
            }
            break;

        case 'startOnlineHost':
            $res = $api->call('startOnlineHost', $sid, [
                'name' => trim($_POST['name'] ?? tr('default_player_name')),
                'numPlayers' => (int) ($_POST['numPlayers'] ?? 2),
                'relay_url' => trim($_POST['relay_url'] ?? ''),
            ]);
            if (!empty($res['ok'])) {
                storeBrowserState($res);
            } else {
                $errorMsg = $res['error'] ?? 'Failed to create lobby';
                markBrowserRuntimeError($errorMsg);
            }
            break;

        case 'joinOnline':
            $res = $api->call('joinOnline', $sid, [
                'name' => trim($_POST['name'] ?? tr('default_player_name')),
                'key' => trim($_POST['key'] ?? ''),
                'role' => $_POST['role'] ?? 'auto',
            ]);
            if (!empty($res['ok'])) {
                storeBrowserState($res);
            } else {
                $errorMsg = $res['error'] ?? 'Failed to join lobby';
                markBrowserRuntimeError($errorMsg);
            }
            break;

        case 'startOnlineMatch':
            $res = $api->call('startOnlineMatch', $sid);
            if (!empty($res['ok'])) {
                storeBrowserState($res);
                refreshRuntimeState($api, $sid, true);
            } else {
                $errorMsg = $res['error'] ?? 'Failed to start match';
                markBrowserRuntimeError($errorMsg);
            }
            break;

        case 'refreshLobby':
        case 'refreshGame':
            refreshRuntimeState($api, $sid, true);
            break;

        case 'sendChat':
            $msg = trim($_POST['message'] ?? '');
            if ($msg !== '') {
                $res = $api->call('sendChat', $sid, ['message' => $msg]);
                if (!empty($res['ok'])) {
                    storeBrowserState($res);
                } else {
                    $errorMsg = $res['error'] ?? 'Failed to send chat';
                    markBrowserRuntimeError($errorMsg);
                }
            }
            refreshRuntimeState($api, $sid, true);
            break;

        case 'voteHost':
            $res = $api->call('sendHostVote', $sid, ['slot' => (int) ($_POST['slot'] ?? -1)]);
            if (!empty($res['ok'])) {
                storeBrowserState($res);
            } else {
                $errorMsg = $res['error'] ?? 'Failed to vote host';
                markBrowserRuntimeError($errorMsg);
            }
            refreshRuntimeState($api, $sid, true);
            break;

        case 'requestReplacementInvite':
            $res = $api->call('requestReplacementInvite', $sid, ['slot' => (int) ($_POST['slot'] ?? -1)]);
            if (!empty($res['ok'])) {
                storeBrowserState($res);
            } else {
                $errorMsg = $res['error'] ?? 'Failed to request replacement invite';
                markBrowserRuntimeError($errorMsg);
            }
            refreshRuntimeState($api, $sid, true);
            break;
    }

    if (!$ajaxRequest) {
        header('Location: index.php');
        exit;
    }
}

$sid = ensureApiSession($api);
$bundle = refreshRuntimeState($api, $sid, true) ?? ($_SESSION['runtime_bundle'] ?? null);
$view = currentViewFromBundle($bundle);
$snap = is_array($bundle['match'] ?? null) ? $bundle['match'] : null;

ob_start();
echo renderRuntimeNotice();
if ($view === 'game' && $snap !== null) {
    include __DIR__ . '/views/game.php';
} elseif ($view === 'lobby') {
    include __DIR__ . '/views/lobby.php';
} else {
    include __DIR__ . '/views/setup.php';
}
$viewHtml = ob_get_clean();

if ($ajaxRequest) {
    header('Content-Type: application/json');
    echo json_encode([
        'ok' => true,
        'view' => $view,
        'viewHtml' => $viewHtml,
    ]);
    exit;
}

$locale = $_SESSION['locale'] ?? 'pt-BR';
$langAttr = ($locale === 'en-US') ? 'en' : 'pt-BR';
?>
<!doctype html>
<html lang="<?= $langAttr ?>">

<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <meta http-equiv="Cache-Control" content="no-cache, no-store, must-revalidate">
    <title><?= tr('title_main') ?></title>
    <meta name="description" content="Truco Paulista no navegador, com mesa offline e online em uma direção visual de botequim refinado.">
    <link rel="icon" type="image/svg+xml" href="favicon.svg">
    <link rel="icon" type="image/png" sizes="512x512" href="favicon.png">
    <link rel="icon" href="favicon.ico" sizes="any">
    <link rel="apple-touch-icon" href="apple-touch-icon.png">
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Barlow:wght@400;500;600;700;800&family=Cormorant+Garamond:wght@500;600;700&family=IBM+Plex+Mono:wght@400;600&display=swap" rel="stylesheet">
    <link rel="stylesheet" href="style.css">
</head>

<body>
    <div class="ambient ambient-a"></div>
    <div class="ambient ambient-b"></div>
    <div class="ambient ambient-c"></div>
    <div class="grain"></div>

    <main class="app-shell">
        <header class="topbar">
            <div class="brand">
                <span class="brand-mark">♣</span>
                <div class="brand-copy">
                    <div class="brand-row">
                        <span class="brand-kicker"><?= tr('brand_kicker') ?></span>
                        <span class="brand-stamp"><?= tr('brand_stamp') ?></span>
                    </div>
                    <h1><?= tr('title_main') ?></h1>
                    <p id="title-sub"><?= tr('title_sub') ?></p>
                </div>
            </div>
            <div class="topbar-actions">
                <form method="post" action="index.php" class="locale-form">
                    <input type="hidden" name="action" value="setLocale">
                    <label for="locale-sel"><?= tr('locale_label') ?></label>
                    <select id="locale-sel" name="locale" class="field field-sm">
                        <option value="pt-BR" <?= $locale === 'pt-BR' ? 'selected' : '' ?>>Português (BR)</option>
                        <option value="en-US" <?= $locale === 'en-US' ? 'selected' : '' ?>>English (US)</option>
                    </select>
                    <button type="submit" class="btn btn-neutral btn-mini">OK</button>
                </form>
            </div>
        </header>

        <div id="view-root">
            <?= $viewHtml ?>
        </div>
    </main>
    <script src="ajax.js" defer></script>
</body>

</html>

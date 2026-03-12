<?php
/**
 * TrucoApiClient — thin wrapper for the Go HTTP API.
 */
class TrucoApiClient {
    private string $baseUrl;

    public function __construct(string $baseUrl = 'http://localhost:9090') {
        $this->baseUrl = rtrim($baseUrl, '/');
    }

    /**
     * Call the Go API. Returns decoded JSON as associative array.
     */
    public function call(string $action, string $sessionId = '', array $payload = []): array {
        $url = $this->baseUrl . '/api/' . $action;
        $json = json_encode($payload ?: new \stdClass());
        $headers = [
            'Content-Type: application/json',
        ];
        if ($sessionId !== '') {
            $headers[] = 'X-Session-ID: ' . $sessionId;
        }

        $ch = curl_init($url);
        curl_setopt_array($ch, [
            CURLOPT_POST           => true,
            CURLOPT_POSTFIELDS     => $json,
            CURLOPT_HTTPHEADER     => $headers,
            CURLOPT_RETURNTRANSFER => true,
            CURLOPT_TIMEOUT        => 5,
            CURLOPT_CONNECTTIMEOUT => 3,
        ]);
        $response = curl_exec($ch);
        $err = curl_error($ch);
        
        if (PHP_VERSION_ID < 80000) {
            curl_close($ch);
        }

        if ($response === false) {
            return ['ok' => false, 'error' => 'API unavailable: ' . $err];
        }

        $decoded = json_decode($response, true);
        if (!is_array($decoded)) {
            return ['ok' => false, 'error' => 'Invalid API response'];
        }
        return $decoded;
    }

    /**
     * Parse the snapshot JSON string from an API response.
     */
    public static function parseSnapshot(array $result): ?array {
        if (empty($result['snapshot'])) {
            return null;
        }
        $snap = json_decode($result['snapshot'], true);
        return is_array($snap) ? $snap : null;
    }

    /**
     * Parse the runtime bundle from an API response.
     */
    public static function parseBundle(array $result): ?array {
        if (empty($result['bundle']) || !is_array($result['bundle'])) {
            return null;
        }
        return $result['bundle'];
    }
}

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
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        $err = curl_error($ch);
        curl_close($ch);

        if ($response === false) {
            return ['ok' => false, 'error' => 'API unavailable: ' . $err, 'action' => $action];
        }

        $decoded = json_decode($response, true);
        if (!is_array($decoded)) {
            return ['ok' => false, 'error' => 'Invalid API response', 'httpCode' => $httpCode, 'action' => $action];
        }

        // Add HTTP code to response for debugging
        if ($httpCode >= 400) {
            $decoded['httpCode'] = $httpCode;
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
}

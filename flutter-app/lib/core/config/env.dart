/// Environment configuration for the BRVM Trading app.
class Env {
  Env._();

  /// Base URL for the Go API server.
  static const String apiBaseUrl = String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: 'http://localhost:8080/api/v1',
  );

  /// WebSocket URL for real-time signals.
  static const String wsBaseUrl = String.fromEnvironment(
    'WS_BASE_URL',
    defaultValue: 'ws://localhost:8080/api/v1/signals/ws',
  );

  /// API request timeout in seconds.
  static const int apiTimeoutSeconds = int.fromEnvironment(
    'API_TIMEOUT',
    defaultValue: 15,
  );
}

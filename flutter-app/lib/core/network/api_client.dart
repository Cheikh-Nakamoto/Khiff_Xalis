import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../config/env.dart';

/// Dio-based HTTP client configured for the BRVM Trading API.
///
/// - Sets base URL from [Env.apiBaseUrl]
/// - Adds JWT auth token via interceptor
/// - Logs requests/responses in debug mode
/// - Handles token expiry with automatic 401 detection
class ApiClient {
  ApiClient({String? token}) : _token = token {
    _dio = Dio(BaseOptions(
      baseUrl: Env.apiBaseUrl,
      connectTimeout: const Duration(seconds: 10),
      receiveTimeout: Duration(seconds: Env.apiTimeoutSeconds),
      sendTimeout: Duration(seconds: Env.apiTimeoutSeconds),
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
      },
    ));

    _dio.interceptors.addAll([
      _AuthInterceptor(this),
      LogInterceptor(
        requestBody: true,
        responseBody: true,
        logPrint: (obj) => print('[ApiClient] $obj'),
      ),
    ]);
  }

  late final Dio _dio;
  String? _token;

  /// Current JWT access token.
  String? get token => _token;

  /// Update the JWT token (called after login/refresh).
  void setToken(String? token) {
    _token = token;
  }

  // ── Public endpoints (no auth) ──

  /// GET /market/tickers
  Future<Response> getTickers() => _dio.get('/market/tickers');

  /// GET /market/data/:ticker
  Future<Response> getMarketData(String ticker, {int days = 90}) =>
      _dio.get('/market/data/$ticker', queryParameters: {'days': days});

  /// GET /market/latest/:ticker
  Future<Response> getLatestData(String ticker) =>
      _dio.get('/market/latest/$ticker');

  /// GET /fundamental/:ticker
  Future<Response> getFundamentals(String ticker) =>
      _dio.get('/fundamental/$ticker');

  /// GET /signals/:ticker
  Future<Response> getSignal(String ticker) =>
      _dio.get('/signals/$ticker');

  /// GET /signals/scan
  Future<Response> scanSignals({double minScore = 60, String? signalType}) =>
      _dio.get('/signals/scan', queryParameters: {
        'min_score': minScore,
        if (signalType != null) 'signal_type': signalType,
      });

  // ── Auth endpoints ──

  /// POST /auth/login
  Future<Response> login(String email, String password) =>
      _dio.post('/auth/login', data: {
        'email': email,
        'password': password,
      });

  /// POST /auth/register
  Future<Response> register({
    required String email,
    required String password,
    required String firstName,
    required String lastName,
  }) =>
      _dio.post('/auth/register', data: {
        'email': email,
        'password': password,
        'first_name': firstName,
        'last_name': lastName,
      });

  /// POST /auth/refresh
  Future<Response> refreshToken(String refreshToken) =>
      _dio.post('/auth/refresh', data: {
        'refresh_token': refreshToken,
      });

  // ── Protected endpoints (JWT required) ──

  /// GET /portfolio
  Future<Response> getPortfolio() => _dio.get('/portfolio');

  /// GET /orders
  Future<Response> getOrders({String status = 'ALL', int limit = 50}) =>
      _dio.get('/orders', queryParameters: {
        'status': status,
        'limit': limit,
      });

  /// POST /orders
  Future<Response> createOrder({
    required String ticker,
    required String side,
    required int quantity,
    required double price,
    required String orderType,
    double? stopPrice,
  }) =>
      _dio.post('/orders', data: {
        'ticker': ticker,
        'side': side,
        'quantity': quantity,
        'price': price,
        'order_type': orderType,
        if (stopPrice != null) 'stop_price': stopPrice,
      });

  /// Dispose the Dio client.
  void dispose() {
    _dio.close();
  }
}

/// Interceptor that attaches the JWT bearer token to every request.
class _AuthInterceptor extends Interceptor {
  _AuthInterceptor(this._client);
  final ApiClient _client;

  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) {
    final token = _client.token;
    if (token != null && token.isNotEmpty) {
      options.headers['Authorization'] = 'Bearer $token';
    }
    handler.next(options);
  }

  @override
  void onError(DioException err, ErrorInterceptorHandler handler) {
    if (err.response?.statusCode == 401) {
      // Token expired or invalid — callers should handle re-login.
      print('[ApiClient] 401 Unauthorized — token may be expired');
    }
    handler.next(err);
  }
}

/// Riverpod provider for [ApiClient].
///
/// The client is created once and shared across the app.
/// Update the token via `ref.read(apiClientProvider).setToken(newToken)`.
final apiClientProvider = Provider<ApiClient>((ref) {
  final client = ApiClient();
  ref.onDispose(() => client.dispose());
  return client;
});

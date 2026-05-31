import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/network/api_client.dart';

/// Authentication state.
class AuthState {
  const AuthState({
    this.accessToken,
    this.refreshToken,
    this.email,
    this.isAuthenticated = false,
    this.isLoading = false,
    this.error,
  });

  final String? accessToken;
  final String? refreshToken;
  final String? email;
  final bool isAuthenticated;
  final bool isLoading;
  final String? error;

  AuthState copyWith({
    String? accessToken,
    String? refreshToken,
    String? email,
    bool? isAuthenticated,
    bool? isLoading,
    String? error,
  }) {
    return AuthState(
      accessToken: accessToken ?? this.accessToken,
      refreshToken: refreshToken ?? this.refreshToken,
      email: email ?? this.email,
      isAuthenticated: isAuthenticated ?? this.isAuthenticated,
      isLoading: isLoading ?? this.isLoading,
      error: error,
    );
  }
}

/// Auth state notifier managing login/register/logout flows.
class AuthNotifier extends StateNotifier<AuthState> {
  AuthNotifier(this._api) : super(const AuthState());

  final ApiClient _api;

  Future<bool> login(String email, String password) async {
    state = state.copyWith(isLoading: true, error: null);
    try {
      final response = await _api.login(email, password);
      final data = response.data as Map<String, dynamic>;
      final accessToken = data['access_token'] as String;
      final refreshToken = data['refresh_token'] as String;

      _api.setToken(accessToken);

      state = AuthState(
        accessToken: accessToken,
        refreshToken: refreshToken,
        email: email,
        isAuthenticated: true,
      );
      return true;
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: _extractError(e),
      );
      return false;
    }
  }

  Future<bool> register({
    required String email,
    required String password,
    required String firstName,
    required String lastName,
  }) async {
    state = state.copyWith(isLoading: true, error: null);
    try {
      final response = await _api.register(
        email: email,
        password: password,
        firstName: firstName,
        lastName: lastName,
      );
      final data = response.data as Map<String, dynamic>;
      final accessToken = data['access_token'] as String;
      final refreshToken = data['refresh_token'] as String;

      _api.setToken(accessToken);

      state = AuthState(
        accessToken: accessToken,
        refreshToken: refreshToken,
        email: email,
        isAuthenticated: true,
      );
      return true;
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: _extractError(e),
      );
      return false;
    }
  }

  Future<void> refreshAccessToken() async {
    final refreshToken = state.refreshToken;
    if (refreshToken == null) return;

    try {
      final response = await _api.refreshToken(refreshToken);
      final data = response.data as Map<String, dynamic>;
      final newToken = data['access_token'] as String;

      _api.setToken(newToken);
      state = state.copyWith(accessToken: newToken);
    } catch (e) {
      // Refresh failed — user must re-login.
      logout();
    }
  }

  void logout() {
    _api.setToken(null);
    state = const AuthState();
  }

  String _extractError(dynamic e) {
    if (e is Exception) {
      try {
        // Dio error
        final dynamic dioError = e;
        final responseData = dioError.response?.data;
        if (responseData is Map<String, dynamic>) {
          return responseData['error'] as String? ?? 'Erreur inconnue';
        }
      } catch (_) {}
    }
    return 'Erreur de connexion au serveur';
  }
}

/// Provider for auth state.
final authProvider = StateNotifierProvider<AuthNotifier, AuthState>((ref) {
  final api = ref.read(apiClientProvider);
  return AuthNotifier(api);
});

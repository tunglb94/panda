import 'dart:convert';

import 'package:shared_preferences/shared_preferences.dart';

import '../../../core/network/api_client.dart';
import '../domain/models/bank_account.dart';
import '../domain/models/payout_request.dart';
import '../domain/models/wallet_statement.dart';
import '../domain/models/wallet_summary.dart';
import '../domain/models/wallet_transaction.dart';

/// Driver Finance / Settlement Engine — Wallet is a read-only projection
/// (Phần 3/13); every figure comes from the backend's ledger, never
/// computed here. Phần 11 — Offline: Wallet Summary and Transaction History
/// are cached to [SharedPreferences] and served as a fallback whenever a
/// live fetch fails with the app's universal offline/timeout sentinel
/// (`ApiException.statusCode == 0` — see `ApiClient`), mirroring the same
/// resilience pattern as the Communication Module's `OfflineMessageQueue`.
class WalletRepository {
  const WalletRepository(this._client);

  final ApiClient _client;

  static const _summaryCacheKey = 'wallet_summary_cache_v1';
  static const _transactionsCacheKey = 'wallet_transactions_cache_v1';

  Future<WalletSummary> fetchSummary() async {
    try {
      final body = await _client.get('/api/v1/driver/wallet/summary');
      final summary = WalletSummary.fromJson(body);
      await _cacheSummary(summary);
      return summary;
    } on ApiException catch (e) {
      if (e.statusCode == 0) {
        final cached = await _readCachedSummary();
        if (cached != null) return cached;
      }
      rethrow;
    }
  }

  Future<void> _cacheSummary(WalletSummary summary) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_summaryCacheKey, jsonEncode(summary.toJson()));
  }

  Future<WalletSummary?> _readCachedSummary() async {
    final prefs = await SharedPreferences.getInstance();
    final raw = prefs.getString(_summaryCacheKey);
    if (raw == null || raw.isEmpty) return null;
    try {
      return WalletSummary.fromJson(jsonDecode(raw) as Map<String, dynamic>);
    } catch (_) {
      return null;
    }
  }

  Future<WalletStatement> fetchStatement({required DateTime from, required DateTime to}) async {
    final body = await _client.get(
      '/api/v1/driver/wallet/statement'
      '?from=${from.toUtc().millisecondsSinceEpoch ~/ 1000}'
      '&to=${to.toUtc().millisecondsSinceEpoch ~/ 1000}',
    );
    return WalletStatement.fromJson(body);
  }

  Future<List<WalletTransaction>> fetchTransactions({int limit = 50}) async {
    try {
      final body = await _client.get('/api/v1/driver/wallet/transactions?limit=$limit');
      final raw = (body['transactions'] as List<dynamic>?) ?? const [];
      final list = raw.map((e) => WalletTransaction.fromJson(e as Map<String, dynamic>)).toList();
      await _cacheTransactions(list);
      return list;
    } on ApiException catch (e) {
      if (e.statusCode == 0) {
        final cached = await _readCachedTransactions();
        if (cached != null) return cached;
      }
      rethrow;
    }
  }

  Future<void> _cacheTransactions(List<WalletTransaction> list) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_transactionsCacheKey, jsonEncode(list.map((e) => e.toJson()).toList()));
  }

  Future<List<WalletTransaction>?> _readCachedTransactions() async {
    final prefs = await SharedPreferences.getInstance();
    final raw = prefs.getString(_transactionsCacheKey);
    if (raw == null || raw.isEmpty) return null;
    try {
      final list = jsonDecode(raw) as List<dynamic>;
      return list.map((e) => WalletTransaction.fromJson(e as Map<String, dynamic>)).toList();
    } catch (_) {
      return null;
    }
  }

  Future<BankAccount?> fetchBankAccount() async {
    try {
      final body = await _client.get('/api/v1/driver/wallet/bank-account');
      return BankAccount.fromJson(body);
    } on ApiException catch (e) {
      if (e.statusCode == 404) return null;
      rethrow;
    }
  }

  Future<BankAccount> setBankAccount({
    required String bankName,
    required String accountHolderName,
    required String accountNumber,
    String branchName = '',
  }) async {
    final body = await _client.post('/api/v1/driver/wallet/bank-account', body: {
      'bank_name': bankName,
      'account_holder_name': accountHolderName,
      'account_number': accountNumber,
      'branch_name': branchName,
    });
    return BankAccount.fromJson(body);
  }

  Future<PayoutRequest> createPayoutRequest(int amountCents) async {
    final body = await _client.post('/api/v1/driver/wallet/payouts', body: {'amount_cents': amountCents});
    return PayoutRequest.fromJson(body);
  }

  Future<List<PayoutRequest>> fetchMyPayoutRequests({int limit = 20}) async {
    final body = await _client.get('/api/v1/driver/wallet/payouts?limit=$limit');
    final raw = (body['payout_requests'] as List<dynamic>?) ?? const [];
    return raw.map((e) => PayoutRequest.fromJson(e as Map<String, dynamic>)).toList();
  }
}

// Unit tests for the Driver Finance domain models' JSON parsing — these
// mirror the exact gateway response shapes from
// backend/services/gateway/http/handlers/wallet_handler.go.
import 'package:flutter_test/flutter_test.dart';

import 'package:driver/features/wallet/domain/models/bank_account.dart';
import 'package:driver/features/wallet/domain/models/payout_request.dart';
import 'package:driver/features/wallet/domain/models/wallet_summary.dart';
import 'package:driver/features/wallet/domain/models/wallet_transaction.dart';

void main() {
  test('WalletSummary.fromJson parses every field', () {
    final s = WalletSummary.fromJson({
      'currency': 'VND',
      'available_cents': 800000,
      'pending_cents': 100000,
      'outstanding_cents': 24000,
      'net_cents': 776000,
      'lifetime_earned_cents': 2000000,
      'lifetime_withdrawn_cents': 500000,
    });
    expect(s.currency, 'VND');
    expect(s.availableCents, 800000);
    expect(s.pendingCents, 100000);
    expect(s.outstandingCents, 24000);
    expect(s.netCents, 776000);
    expect(s.lifetimeEarnedCents, 2000000);
    expect(s.lifetimeWithdrawnCents, 500000);
  });

  test('WalletSummary.empty is all zero', () {
    expect(WalletSummary.empty.availableCents, 0);
    expect(WalletSummary.empty.currency, 'VND');
  });

  test('BankAccount.fromJson never carries a raw account number field', () {
    final b = BankAccount.fromJson({
      'bank_name': 'Vietcombank',
      'account_holder_name': 'Nguyen Van A',
      'masked_account_number': '••••6789',
      'branch_name': 'Chi nhánh Q1',
      'updated_at': '2026-07-01T00:00:00Z',
    });
    expect(b.maskedAccountNumber, '••••6789');
    expect(b.bankName, 'Vietcombank');
  });

  test('PayoutRequest.fromJson parses status and reject reason', () {
    final p = PayoutRequest.fromJson({
      'payout_request_id': 'p1',
      'amount_cents': 300000,
      'currency': 'VND',
      'bank_name': 'Vietcombank',
      'masked_account_number': '••••6789',
      'status': 'rejected',
      'requested_at': '2026-07-01T00:00:00Z',
      'reject_reason': 'Sai thông tin ngân hàng',
    });
    expect(p.status, 'rejected');
    expect(p.rejectReason, 'Sai thông tin ngân hàng');
    expect(payoutStatusLabel(p.status), 'Bị từ chối');
  });

  test('WalletTransaction.fromJson computes isCredit correctly', () {
    final credit = WalletTransaction.fromJson({
      'type': 'ride_income',
      'direction': 'credit',
      'amount_cents': 96000,
      'currency': 'VND',
      'description': '',
      'payment_method': 'wallet',
      'created_at': '2026-07-01T00:00:00Z',
    });
    final debit = WalletTransaction.fromJson({
      'type': 'withdrawal',
      'direction': 'debit',
      'amount_cents': 300000,
      'currency': 'VND',
      'description': '',
      'payment_method': '',
      'created_at': '2026-07-01T00:00:00Z',
    });
    expect(credit.isCredit, isTrue);
    expect(debit.isCredit, isFalse);
    expect(walletTransactionTypeLabel(credit.type), 'Thu nhập chuyến xe');
    expect(walletTransactionTypeLabel(debit.type), 'Rút tiền');
  });
}

/// A driver's single default payout bank account (Phần 6). Only ever
/// carries the masked account number — the raw number never leaves the
/// backend (Phần 10/12).
class BankAccount {
  const BankAccount({
    required this.bankName,
    required this.accountHolderName,
    required this.maskedAccountNumber,
    required this.branchName,
    required this.updatedAt,
  });

  final String bankName;
  final String accountHolderName;
  final String maskedAccountNumber;
  final String branchName;
  final DateTime? updatedAt;

  factory BankAccount.fromJson(Map<String, dynamic> json) => BankAccount(
        bankName: json['bank_name'] as String? ?? '',
        accountHolderName: json['account_holder_name'] as String? ?? '',
        maskedAccountNumber: json['masked_account_number'] as String? ?? '',
        branchName: json['branch_name'] as String? ?? '',
        updatedAt: DateTime.tryParse(json['updated_at'] as String? ?? ''),
      );
}

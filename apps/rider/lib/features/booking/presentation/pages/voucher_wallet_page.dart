import 'package:flutter/material.dart';

import 'package:rider/core/network/api_client.dart';
import 'package:rider/core/theme/app_colors.dart';
import 'package:rider/core/theme/app_spacing.dart';
import 'package:rider/shared/utils/currency_format.dart';
import 'package:rider/shared/widgets/app_empty_state.dart';
import 'package:rider/shared/widgets/app_loading_view.dart';

import '../../data/promotion_repository.dart';
import '../../domain/models/voucher.dart';
import '../widgets/voucher_card.dart';

/// Ví Voucher — Available / Used / Expired, backed by
/// `GET /api/v1/rider/vouchers`.
class VoucherWalletPage extends StatefulWidget {
  const VoucherWalletPage({super.key, required this.apiClient});

  final ApiClient apiClient;

  @override
  State<VoucherWalletPage> createState() => _VoucherWalletPageState();
}

class _VoucherWalletPageState extends State<VoucherWalletPage> {
  bool _loading = true;
  String? _error;
  List<Voucher> _available = const [];
  List<Voucher> _expired = const [];
  List<_UsedEntry> _used = const [];

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final body = await PromotionRepository(widget.apiClient).myVouchers();
      final available = (body['available'] as List<dynamic>? ?? [])
          .map((e) => Voucher.fromApi(e as Map<String, dynamic>, status: VoucherStatus.available))
          .toList();
      final expired = (body['expired'] as List<dynamic>? ?? [])
          .map((e) => Voucher.fromApi(e as Map<String, dynamic>, status: VoucherStatus.expired))
          .toList();
      final used = (body['used'] as List<dynamic>? ?? [])
          .map((e) => _UsedEntry.fromJson(e as Map<String, dynamic>))
          .toList();
      if (!mounted) return;
      setState(() {
        _available = available;
        _expired = expired;
        _used = used;
        _loading = false;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() {
        _loading = false;
        _error = 'Không thể tải ví voucher.';
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return DefaultTabController(
      length: 3,
      child: Scaffold(
        appBar: AppBar(
          title: const Text('Ví voucher'),
          bottom: const TabBar(
            tabs: [
              Tab(text: 'Có thể dùng'),
              Tab(text: 'Đã dùng'),
              Tab(text: 'Hết hạn'),
            ],
          ),
        ),
        body: SafeArea(
          child: _loading
              ? const AppLoadingView(label: 'Đang tải…')
              : _error != null
                  ? Center(child: Text(_error!, style: const TextStyle(color: AppColors.error)))
                  : RefreshIndicator(
                      onRefresh: _load,
                      child: TabBarView(
                        children: [
                          _VoucherList(vouchers: _available, emptyText: 'Chưa có voucher nào'),
                          _UsedList(entries: _used),
                          _VoucherList(vouchers: _expired, emptyText: 'Chưa có voucher hết hạn'),
                        ],
                      ),
                    ),
        ),
      ),
    );
  }
}

class _UsedEntry {
  const _UsedEntry({required this.voucherName, required this.voucherCode, required this.discountAmount, required this.tripId});

  factory _UsedEntry.fromJson(Map<String, dynamic> json) => _UsedEntry(
        voucherName: json['voucher_name'] as String? ?? '',
        voucherCode: json['voucher_code'] as String? ?? '',
        discountAmount: (json['discount_amount'] as num?)?.toInt() ?? 0,
        tripId: json['trip_id'] as String? ?? '',
      );

  final String voucherName;
  final String voucherCode;
  final int discountAmount;
  final String tripId;
}

class _VoucherList extends StatelessWidget {
  const _VoucherList({required this.vouchers, required this.emptyText});

  final List<Voucher> vouchers;
  final String emptyText;

  @override
  Widget build(BuildContext context) {
    if (vouchers.isEmpty) {
      return AppEmptyState(icon: Icons.local_offer_outlined, title: emptyText, mascotAsset: 'mascot_voucher.png');
    }
    return ListView.separated(
      padding: const EdgeInsets.all(AppSpacing.lg),
      itemCount: vouchers.length,
      separatorBuilder: (_, _) => const SizedBox(height: AppSpacing.md),
      itemBuilder: (context, i) => VoucherCard(voucher: vouchers[i]),
    );
  }
}

class _UsedList extends StatelessWidget {
  const _UsedList({required this.entries});

  final List<_UsedEntry> entries;

  @override
  Widget build(BuildContext context) {
    if (entries.isEmpty) {
      return const AppEmptyState(
        icon: Icons.local_offer_outlined,
        title: 'Chưa dùng voucher nào',
        mascotAsset: 'mascot_voucher.png',
      );
    }
    return ListView.separated(
      padding: const EdgeInsets.all(AppSpacing.lg),
      itemCount: entries.length,
      separatorBuilder: (_, _) => const SizedBox(height: AppSpacing.sm),
      itemBuilder: (context, i) {
        final e = entries[i];
        return ListTile(
          leading: const Icon(Icons.check_circle, color: AppColors.primary),
          title: Text(e.voucherName.isNotEmpty ? e.voucherName : e.voucherCode),
          subtitle: Text('Mã: ${e.voucherCode} · Chuyến ${e.tripId}'),
          trailing: Text(
            '-${formatMoney(e.discountAmount, 'VND')}',
            style: const TextStyle(color: AppColors.primary, fontWeight: FontWeight.w600),
          ),
        );
      },
    );
  }
}

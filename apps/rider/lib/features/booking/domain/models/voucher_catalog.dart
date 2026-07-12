import 'voucher.dart';

/// The rider's available vouchers.
///
/// Always empty today: `backend/services/promotion` has real domain logic
/// (Voucher entity, VoucherValidator, PromotionService) but no gRPC handler
/// or REST route exists for any client to call — see the Promotion Engine
/// CHANGELOG entry. Rather than fabricate sample vouchers to make this
/// screen look populated, [VoucherCatalog.mine] stays honestly empty until
/// a real `GET /api/v1/rider/vouchers`-shaped endpoint exists; the rider
/// sees "Chưa có voucher nào" (see `voucher_list_sheet.dart`), which is
/// simply true.
///
/// [VoucherCard]/[VoucherListSheet] are fully built against the [Voucher]
/// model so wiring a real repository later is a one-line swap of this
/// source, not a UI rewrite.
abstract final class VoucherCatalog {
  static const List<Voucher> mine = [];
}

import 'package:flutter/material.dart';

import '../../../../core/network/api_client.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_spacing.dart';
import '../../../../shared/widgets/app_button.dart';
import '../../../../shared/widgets/app_snackbar.dart';
import '../../data/kyc_repository.dart';
import '../../domain/models/document_type.dart';
import '../../domain/models/license_class.dart';
import '../widgets/document_upload_tile.dart';
import 'kyc_status_page.dart';

/// "Become Driver" — the 6-step KYC + Vehicle Verification wizard (Phần 5).
/// Documents upload immediately per-step (Phần 4's document endpoints);
/// personal-info and vehicle-info are only sent to the backend on the final
/// Step 6 Submit, which the backend accepts as a create-or-resubmit (so
/// re-opening this wizard after a Rejected status works with the same
/// single Submit action — no separate "edit" mode needed client-side).
class BecomeDriverPage extends StatefulWidget {
  const BecomeDriverPage({super.key, required this.apiClient});

  final ApiClient apiClient;

  @override
  State<BecomeDriverPage> createState() => _BecomeDriverPageState();
}

class _BecomeDriverPageState extends State<BecomeDriverPage> {
  late final KYCRepository _repo = KYCRepository(widget.apiClient);
  int _step = 0;
  bool _loadingExisting = true;

  // Step 1 — personal info.
  final _fullNameCtrl = TextEditingController();
  final _addressCtrl = TextEditingController();
  final _nationalIdCtrl = TextEditingController();
  DateTime? _dob;

  // Step 3 — GPLX.
  final _licenseNumberCtrl = TextEditingController();
  String? _licenseClass;

  // Step 4 — vehicle info.
  final _brandCtrl = TextEditingController();
  final _modelCtrl = TextEditingController();
  final _yearCtrl = TextEditingController();
  final _colorCtrl = TextEditingController();
  final _plateCtrl = TextEditingController();
  final _vinCtrl = TextEditingController();
  final _engineNumberCtrl = TextEditingController();
  final _chassisNumberCtrl = TextEditingController();
  String _vehicleType = 'motorcycle';
  String _serviceType = 'motorcycle'; // "bike" wire alias — see backend ServiceTypeBike
  bool _rideEnabled = true;
  bool _deliveryEnabled = false;

  Map<String, bool> _uploaded = {};
  Map<String, DocumentChecklistItem> _docItems = {};
  bool _submitting = false;
  String? _submitError;

  @override
  void initState() {
    super.initState();
    _loadExisting();
  }

  @override
  void dispose() {
    _fullNameCtrl.dispose();
    _addressCtrl.dispose();
    _nationalIdCtrl.dispose();
    _licenseNumberCtrl.dispose();
    _brandCtrl.dispose();
    _modelCtrl.dispose();
    _yearCtrl.dispose();
    _colorCtrl.dispose();
    _plateCtrl.dispose();
    _vinCtrl.dispose();
    _engineNumberCtrl.dispose();
    _chassisNumberCtrl.dispose();
    super.dispose();
  }

  /// Best-effort prefill from a prior Pending/Rejected submission, so
  /// re-opening this wizard doesn't lose already-entered data.
  Future<void> _loadExisting() async {
    try {
      final dv = await _repo.getDriverVerification();
      if (dv != null) {
        _fullNameCtrl.text = dv.fullName;
        _addressCtrl.text = dv.address;
        _nationalIdCtrl.text = dv.nationalIdNumber;
        _licenseNumberCtrl.text = dv.licenseNumber;
        _dob = dv.dateOfBirth;
      }
    } catch (_) {
      // Ignore — wizard just starts blank.
    }
    try {
      final vv = await _repo.getVehicleVerification();
      if (vv != null) {
        _brandCtrl.text = vv.brand;
        _modelCtrl.text = vv.model;
        _yearCtrl.text = vv.year > 0 ? vv.year.toString() : '';
        _colorCtrl.text = vv.color;
        _plateCtrl.text = vv.plateNumber;
        _vinCtrl.text = vv.vin;
        _engineNumberCtrl.text = vv.engineNumber;
        _chassisNumberCtrl.text = vv.chassisNumber;
        if (vv.vehicleType.isNotEmpty) _vehicleType = vv.vehicleType;
        if (vv.serviceType.isNotEmpty) _serviceType = vv.serviceType;
        if (vv.licenseClass.isNotEmpty) _licenseClass = vv.licenseClass;
        _rideEnabled = vv.rideEnabled;
        _deliveryEnabled = vv.deliveryEnabled;
      }
    } catch (_) {
      // Ignore.
    }
    await _refreshDocs();
    if (mounted) setState(() => _loadingExisting = false);
  }

  Future<void> _refreshDocs() async {
    try {
      final list = await _repo.listDocuments();
      final uploaded = {for (final d in list) d.documentType: d.uploaded};
      final items = {for (final d in list) d.documentType: d};
      if (mounted) {
        setState(() {
          _uploaded = uploaded;
          _docItems = items;
        });
      }
    } catch (_) {
      // Ignore — checklist just stays as last known.
    }
  }

  bool get _step1Valid =>
      _fullNameCtrl.text.trim().isNotEmpty &&
      _addressCtrl.text.trim().isNotEmpty &&
      _nationalIdCtrl.text.trim().isNotEmpty &&
      _dob != null;

  bool get _step2Valid =>
      (_uploaded[DocumentType.cccdFront] ?? false) &&
      (_uploaded[DocumentType.cccdBack] ?? false) &&
      (_uploaded[DocumentType.selfie] ?? false);

  // Phần 10: GPLX bắt buộc nếu chạy Ride.
  bool get _step3Valid =>
      !_rideEnabled ||
      ((_uploaded[DocumentType.license] ?? false) &&
          _licenseClass != null &&
          _licenseNumberCtrl.text.trim().isNotEmpty);

  bool get _step4Valid =>
      _brandCtrl.text.trim().isNotEmpty &&
      _modelCtrl.text.trim().isNotEmpty &&
      _plateCtrl.text.trim().isNotEmpty &&
      (_rideEnabled || _deliveryEnabled);

  bool get _step5Valid =>
      (_uploaded[DocumentType.vehicleRegistration] ?? false) && (_uploaded[DocumentType.vehicleInsurance] ?? false);

  bool get _canContinue => switch (_step) {
        0 => _step1Valid,
        1 => _step2Valid,
        2 => _step3Valid,
        3 => _step4Valid,
        4 => _step5Valid,
        _ => true,
      };

  void _next() {
    if (!_canContinue) return;
    if (_step < 5) setState(() => _step++);
  }

  void _back() {
    if (_step > 0) setState(() => _step--);
  }

  Future<void> _pickDateOfBirth() async {
    final now = DateTime.now();
    final picked = await showDatePicker(
      context: context,
      initialDate: _dob ?? DateTime(now.year - 25),
      firstDate: DateTime(now.year - 100),
      lastDate: DateTime(now.year - 18, now.month, now.day),
      helpText: 'Ngày sinh',
    );
    if (picked != null) setState(() => _dob = picked);
  }

  Future<void> _submit() async {
    if (_submitting) return;
    setState(() {
      _submitting = true;
      _submitError = null;
    });
    try {
      await _repo.submitDriverVerification(
        fullName: _fullNameCtrl.text.trim(),
        dateOfBirth: _dob!,
        address: _addressCtrl.text.trim(),
        nationalIdNumber: _nationalIdCtrl.text.trim(),
        licenseNumber: _licenseNumberCtrl.text.trim(),
      );
      final year = int.tryParse(_yearCtrl.text.trim()) ?? 0;
      await _repo.submitVehicleVerification(VehicleVerificationInput(
        vehicleType: _vehicleType,
        serviceType: _serviceType,
        brand: _brandCtrl.text.trim(),
        model: _modelCtrl.text.trim(),
        year: year,
        color: _colorCtrl.text.trim(),
        plateNumber: _plateCtrl.text.trim(),
        vin: _vinCtrl.text.trim(),
        engineNumber: _engineNumberCtrl.text.trim(),
        chassisNumber: _chassisNumberCtrl.text.trim(),
        licenseClass: _rideEnabled ? (_licenseClass ?? '') : '',
        rideEnabled: _rideEnabled,
        deliveryEnabled: _deliveryEnabled,
      ));
      if (!mounted) return;
      AppSnackbar.success(context, 'Đã gửi hồ sơ xác minh thành công!');
      Navigator.of(context).pushReplacement(
        MaterialPageRoute(builder: (_) => KYCStatusPage(apiClient: widget.apiClient)),
      );
    } on ApiException catch (e) {
      if (mounted) {
        setState(() {
          _submitting = false;
          _submitError = e.statusCode == 0 ? e.message : 'Không thể gửi hồ sơ. Vui lòng thử lại.';
        });
      }
    } catch (_) {
      if (mounted) {
        setState(() {
          _submitting = false;
          _submitError = 'Không thể gửi hồ sơ. Vui lòng thử lại.';
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    if (_loadingExisting) {
      return const Scaffold(body: Center(child: CircularProgressIndicator()));
    }
    return Scaffold(
      appBar: AppBar(title: Text('Đăng ký tài xế — Bước ${_step + 1}/6')),
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 480),
            child: Column(
              children: [
                _ProgressBar(step: _step, total: 6),
                Expanded(
                  child: SingleChildScrollView(
                    padding: const EdgeInsets.all(AppSpacing.lg),
                    child: _buildStep(),
                  ),
                ),
                _buildControls(),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildStep() {
    return switch (_step) {
      0 => _Step1PersonalInfo(
          fullNameCtrl: _fullNameCtrl,
          addressCtrl: _addressCtrl,
          nationalIdCtrl: _nationalIdCtrl,
          dob: _dob,
          onPickDob: _pickDateOfBirth,
          onChanged: () => setState(() {}),
        ),
      1 => _Step2Documents(repository: _repo, uploaded: _uploaded, onUploaded: _refreshDocs),
      2 => _Step3License(
          repository: _repo,
          uploaded: _uploaded,
          docItems: _docItems,
          onUploaded: _refreshDocs,
          licenseNumberCtrl: _licenseNumberCtrl,
          licenseClass: _licenseClass,
          rideEnabled: _rideEnabled,
          onLicenseClassChanged: (v) => setState(() => _licenseClass = v),
          onChanged: () => setState(() {}),
        ),
      3 => _Step4VehicleInfo(
          brandCtrl: _brandCtrl,
          modelCtrl: _modelCtrl,
          yearCtrl: _yearCtrl,
          colorCtrl: _colorCtrl,
          plateCtrl: _plateCtrl,
          vinCtrl: _vinCtrl,
          engineNumberCtrl: _engineNumberCtrl,
          chassisNumberCtrl: _chassisNumberCtrl,
          vehicleType: _vehicleType,
          serviceType: _serviceType,
          rideEnabled: _rideEnabled,
          deliveryEnabled: _deliveryEnabled,
          onVehicleTypeChanged: (vt, st) => setState(() {
            _vehicleType = vt;
            _serviceType = st;
            if (_licenseClass != null && !licenseAllowsServiceType(_licenseClass!, _serviceType)) {
              _licenseClass = null;
            }
          }),
          onRideChanged: (v) => setState(() => _rideEnabled = v),
          onDeliveryChanged: (v) => setState(() => _deliveryEnabled = v),
          onChanged: () => setState(() {}),
        ),
      4 => _Step5VehicleDocuments(repository: _repo, uploaded: _uploaded, docItems: _docItems, onUploaded: _refreshDocs),
      _ => _Step6Review(
          fullName: _fullNameCtrl.text,
          address: _addressCtrl.text,
          nationalIdNumber: _nationalIdCtrl.text,
          dob: _dob,
          vehicleType: _vehicleType,
          serviceType: _serviceType,
          brand: _brandCtrl.text,
          model: _modelCtrl.text,
          plateNumber: _plateCtrl.text,
          licenseClass: _licenseClass,
          rideEnabled: _rideEnabled,
          deliveryEnabled: _deliveryEnabled,
          error: _submitError,
        ),
    };
  }

  Widget _buildControls() {
    return Padding(
      padding: const EdgeInsets.fromLTRB(AppSpacing.lg, 0, AppSpacing.lg, AppSpacing.lg),
      child: Row(
        children: [
          if (_step > 0)
            Expanded(child: AppButton.outline(label: 'Quay lại', onPressed: _submitting ? null : _back)),
          if (_step > 0) const SizedBox(width: AppSpacing.md),
          Expanded(
            child: _step < 5
                ? AppButton.primary(label: 'Tiếp tục', onPressed: _canContinue ? _next : null)
                : AppButton.primary(label: 'Gửi hồ sơ', isLoading: _submitting, onPressed: _submit),
          ),
        ],
      ),
    );
  }
}

class _ProgressBar extends StatelessWidget {
  const _ProgressBar({required this.step, required this.total});

  final int step;
  final int total;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(AppSpacing.lg, AppSpacing.sm, AppSpacing.lg, AppSpacing.sm),
      child: Semantics(
        label: 'Bước ${step + 1} trên $total',
        child: ClipRRect(
          borderRadius: BorderRadius.circular(4),
          child: LinearProgressIndicator(
            value: (step + 1) / total,
            minHeight: 6,
            backgroundColor: AppColors.divider,
            valueColor: const AlwaysStoppedAnimation(AppColors.primary),
          ),
        ),
      ),
    );
  }
}

// ─── Step 1 — Personal info ─────────────────────────────────────────────────

class _Step1PersonalInfo extends StatelessWidget {
  const _Step1PersonalInfo({
    required this.fullNameCtrl,
    required this.addressCtrl,
    required this.nationalIdCtrl,
    required this.dob,
    required this.onPickDob,
    required this.onChanged,
  });

  final TextEditingController fullNameCtrl;
  final TextEditingController addressCtrl;
  final TextEditingController nationalIdCtrl;
  final DateTime? dob;
  final VoidCallback onPickDob;

  /// Fired on every keystroke in any field here — without it, the parent
  /// never rebuilds after typing, so "Tiếp tục" stays disabled (stuck on
  /// whatever _canContinue evaluated to when the page last rebuilt) even
  /// after every field is genuinely filled in.
  final VoidCallback onChanged;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Thông tin cá nhân', style: Theme.of(context).textTheme.titleLarge),
        const SizedBox(height: AppSpacing.sm),
        Text(
          'Thông tin này phải khớp với CCCD của bạn.',
          style: Theme.of(context).textTheme.bodySmall?.copyWith(color: AppColors.textSecondary),
        ),
        const SizedBox(height: AppSpacing.lg),
        TextField(
          controller: fullNameCtrl,
          textCapitalization: TextCapitalization.words,
          decoration: const InputDecoration(labelText: 'Họ và tên', prefixIcon: Icon(Icons.person_outline)),
          onChanged: (_) => onChanged(),
        ),
        const SizedBox(height: AppSpacing.md),
        InkWell(
          onTap: onPickDob,
          child: InputDecorator(
            decoration: const InputDecoration(labelText: 'Ngày sinh', prefixIcon: Icon(Icons.cake_outlined)),
            child: Text(
              dob == null ? 'Chọn ngày sinh' : '${dob!.day.toString().padLeft(2, '0')}/${dob!.month.toString().padLeft(2, '0')}/${dob!.year}',
            ),
          ),
        ),
        const SizedBox(height: AppSpacing.md),
        TextField(
          controller: addressCtrl,
          maxLines: 2,
          decoration: const InputDecoration(labelText: 'Địa chỉ thường trú', prefixIcon: Icon(Icons.home_outlined)),
          onChanged: (_) => onChanged(),
        ),
        const SizedBox(height: AppSpacing.md),
        TextField(
          controller: nationalIdCtrl,
          keyboardType: TextInputType.number,
          decoration: const InputDecoration(labelText: 'Số CCCD', prefixIcon: Icon(Icons.badge_outlined)),
          onChanged: (_) => onChanged(),
        ),
      ],
    );
  }
}

// ─── Step 2 — CCCD + Selfie ─────────────────────────────────────────────────

class _Step2Documents extends StatelessWidget {
  const _Step2Documents({required this.repository, required this.uploaded, required this.onUploaded});

  final KYCRepository repository;
  final Map<String, bool> uploaded;
  final VoidCallback onUploaded;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Căn cước công dân', style: Theme.of(context).textTheme.titleLarge),
        const SizedBox(height: AppSpacing.sm),
        Text(
          'Chụp rõ 4 góc, không lóa sáng.',
          style: Theme.of(context).textTheme.bodySmall?.copyWith(color: AppColors.textSecondary),
        ),
        const SizedBox(height: AppSpacing.lg),
        DocumentUploadTile(
          label: documentTypeLabel(DocumentType.cccdFront),
          documentType: DocumentType.cccdFront,
          repository: repository,
          uploaded: uploaded[DocumentType.cccdFront] ?? false,
          onUploaded: onUploaded,
        ),
        const SizedBox(height: AppSpacing.sm),
        DocumentUploadTile(
          label: documentTypeLabel(DocumentType.cccdBack),
          documentType: DocumentType.cccdBack,
          repository: repository,
          uploaded: uploaded[DocumentType.cccdBack] ?? false,
          onUploaded: onUploaded,
        ),
        const SizedBox(height: AppSpacing.sm),
        DocumentUploadTile(
          label: documentTypeLabel(DocumentType.selfie),
          documentType: DocumentType.selfie,
          repository: repository,
          uploaded: uploaded[DocumentType.selfie] ?? false,
          onUploaded: onUploaded,
        ),
      ],
    );
  }
}

// ─── Step 3 — GPLX ───────────────────────────────────────────────────────────

class _Step3License extends StatelessWidget {
  const _Step3License({
    required this.repository,
    required this.uploaded,
    required this.docItems,
    required this.onUploaded,
    required this.licenseNumberCtrl,
    required this.licenseClass,
    required this.rideEnabled,
    required this.onLicenseClassChanged,
    required this.onChanged,
  });

  final KYCRepository repository;
  final Map<String, bool> uploaded;
  final Map<String, DocumentChecklistItem> docItems;
  final VoidCallback onUploaded;
  final TextEditingController licenseNumberCtrl;
  final String? licenseClass;
  final bool rideEnabled;
  final ValueChanged<String?> onLicenseClassChanged;

  /// Fired on every keystroke in licenseNumberCtrl — see
  /// _Step1PersonalInfo.onChanged's doc comment for why this is needed.
  final VoidCallback onChanged;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Giấy phép lái xe (GPLX)', style: Theme.of(context).textTheme.titleLarge),
        const SizedBox(height: AppSpacing.sm),
        Text(
          rideEnabled
              ? 'Bắt buộc nếu bạn muốn chở khách (Ride).'
              : 'Không bắt buộc — bạn đang chọn chỉ giao hàng (Delivery) ở bước sau.',
          style: Theme.of(context).textTheme.bodySmall?.copyWith(color: AppColors.textSecondary),
        ),
        const SizedBox(height: AppSpacing.lg),
        DocumentUploadTile(
          label: documentTypeLabel(DocumentType.license),
          documentType: DocumentType.license,
          repository: repository,
          uploaded: uploaded[DocumentType.license] ?? false,
          requiresExpiry: true,
          expiresAt: docItems[DocumentType.license]?.expiresAt,
          expired: docItems[DocumentType.license]?.expired ?? false,
          expiringSoon: docItems[DocumentType.license]?.expiringSoon ?? false,
          version: docItems[DocumentType.license]?.version ?? 0,
          onUploaded: onUploaded,
        ),
        const SizedBox(height: AppSpacing.md),
        TextField(
          controller: licenseNumberCtrl,
          decoration: const InputDecoration(labelText: 'Số GPLX', prefixIcon: Icon(Icons.numbers)),
          onChanged: (_) => onChanged(),
        ),
        const SizedBox(height: AppSpacing.md),
        DropdownButtonFormField<String>(
          initialValue: licenseClass,
          decoration: const InputDecoration(labelText: 'Hạng bằng lái', prefixIcon: Icon(Icons.motorcycle_outlined)),
          items: LicenseClass.all
              .map((c) => DropdownMenuItem(value: c, child: Text(licenseClassLabel(c))))
              .toList(),
          onChanged: onLicenseClassChanged,
        ),
      ],
    );
  }
}

// ─── Step 4 — Vehicle info ───────────────────────────────────────────────────

class _Step4VehicleInfo extends StatelessWidget {
  const _Step4VehicleInfo({
    required this.brandCtrl,
    required this.modelCtrl,
    required this.yearCtrl,
    required this.colorCtrl,
    required this.plateCtrl,
    required this.vinCtrl,
    required this.engineNumberCtrl,
    required this.chassisNumberCtrl,
    required this.vehicleType,
    required this.serviceType,
    required this.rideEnabled,
    required this.deliveryEnabled,
    required this.onVehicleTypeChanged,
    required this.onRideChanged,
    required this.onDeliveryChanged,
    required this.onChanged,
  });

  final TextEditingController brandCtrl;
  final TextEditingController modelCtrl;
  final TextEditingController yearCtrl;
  final TextEditingController colorCtrl;
  final TextEditingController plateCtrl;
  final TextEditingController vinCtrl;
  final TextEditingController engineNumberCtrl;
  final TextEditingController chassisNumberCtrl;
  final String vehicleType;
  final String serviceType;
  final bool rideEnabled;
  final bool deliveryEnabled;
  final void Function(String vehicleType, String serviceType) onVehicleTypeChanged;
  final ValueChanged<bool> onRideChanged;
  final ValueChanged<bool> onDeliveryChanged;

  /// Fired on every keystroke in any text field here — see
  /// _Step1PersonalInfo.onChanged's doc comment for why this is needed.
  final VoidCallback onChanged;

  static const _tiers = [
    ('motorcycle', 'motorcycle', 'Xe máy — Bike'),
    ('motorcycle', 'bike_plus', 'Xe máy — Bike Plus'),
    ('car', 'car', 'Ô tô — Car'),
    ('van', 'car_xl', 'Ô tô — Car XL'),
  ];

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Thông tin xe', style: Theme.of(context).textTheme.titleLarge),
        const SizedBox(height: AppSpacing.lg),
        DropdownButtonFormField<String>(
          initialValue: serviceType,
          decoration: const InputDecoration(labelText: 'Loại dịch vụ', prefixIcon: Icon(Icons.two_wheeler)),
          items: _tiers
              .map((t) => DropdownMenuItem(value: t.$2, child: Text(t.$3)))
              .toList(),
          onChanged: (st) {
            final tier = _tiers.firstWhere((t) => t.$2 == st);
            onVehicleTypeChanged(tier.$1, tier.$2);
          },
        ),
        const SizedBox(height: AppSpacing.md),
        TextField(
          controller: brandCtrl,
          decoration: const InputDecoration(labelText: 'Hãng xe', prefixIcon: Icon(Icons.directions_car_filled_outlined)),
          onChanged: (_) => onChanged(),
        ),
        const SizedBox(height: AppSpacing.md),
        TextField(
          controller: modelCtrl,
          decoration: const InputDecoration(labelText: 'Model'),
          onChanged: (_) => onChanged(),
        ),
        const SizedBox(height: AppSpacing.md),
        Row(
          children: [
            Expanded(
              child: TextField(
                controller: yearCtrl,
                keyboardType: TextInputType.number,
                decoration: const InputDecoration(labelText: 'Năm sản xuất'),
                onChanged: (_) => onChanged(),
              ),
            ),
            const SizedBox(width: AppSpacing.md),
            Expanded(
              child: TextField(
                controller: colorCtrl,
                decoration: const InputDecoration(labelText: 'Màu xe'),
                onChanged: (_) => onChanged(),
              ),
            ),
          ],
        ),
        const SizedBox(height: AppSpacing.md),
        TextField(
          controller: plateCtrl,
          textCapitalization: TextCapitalization.characters,
          decoration: const InputDecoration(labelText: 'Biển số xe', prefixIcon: Icon(Icons.pin_outlined)),
          onChanged: (_) => onChanged(),
        ),
        const SizedBox(height: AppSpacing.lg),
        Text('Thông tin định danh xe (không bắt buộc)', style: Theme.of(context).textTheme.titleSmall),
        const SizedBox(height: AppSpacing.sm),
        TextField(
          controller: vinCtrl,
          textCapitalization: TextCapitalization.characters,
          decoration: const InputDecoration(labelText: 'Số VIN'),
          onChanged: (_) => onChanged(),
        ),
        const SizedBox(height: AppSpacing.md),
        TextField(
          controller: engineNumberCtrl,
          textCapitalization: TextCapitalization.characters,
          decoration: const InputDecoration(labelText: 'Số máy'),
          onChanged: (_) => onChanged(),
        ),
        const SizedBox(height: AppSpacing.md),
        TextField(
          controller: chassisNumberCtrl,
          textCapitalization: TextCapitalization.characters,
          decoration: const InputDecoration(labelText: 'Số khung'),
          onChanged: (_) => onChanged(),
        ),
        const SizedBox(height: AppSpacing.lg),
        Text('Bạn muốn nhận loại đơn nào?', style: Theme.of(context).textTheme.titleSmall),
        CheckboxListTile(
          value: rideEnabled,
          onChanged: (v) => onRideChanged(v ?? false),
          title: const Text('Chở khách (Ride)'),
          contentPadding: EdgeInsets.zero,
          controlAffinity: ListTileControlAffinity.leading,
        ),
        CheckboxListTile(
          value: deliveryEnabled,
          onChanged: (v) => onDeliveryChanged(v ?? false),
          title: const Text('Giao hàng (Delivery)'),
          contentPadding: EdgeInsets.zero,
          controlAffinity: ListTileControlAffinity.leading,
        ),
      ],
    );
  }
}

// ─── Step 5 — Vehicle documents ─────────────────────────────────────────────

class _Step5VehicleDocuments extends StatelessWidget {
  const _Step5VehicleDocuments({
    required this.repository,
    required this.uploaded,
    required this.docItems,
    required this.onUploaded,
  });

  final KYCRepository repository;
  final Map<String, bool> uploaded;
  final Map<String, DocumentChecklistItem> docItems;
  final VoidCallback onUploaded;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Giấy tờ xe', style: Theme.of(context).textTheme.titleLarge),
        const SizedBox(height: AppSpacing.lg),
        DocumentUploadTile(
          label: documentTypeLabel(DocumentType.vehicleRegistration),
          documentType: DocumentType.vehicleRegistration,
          repository: repository,
          uploaded: uploaded[DocumentType.vehicleRegistration] ?? false,
          requiresExpiry: true,
          expiresAt: docItems[DocumentType.vehicleRegistration]?.expiresAt,
          expired: docItems[DocumentType.vehicleRegistration]?.expired ?? false,
          expiringSoon: docItems[DocumentType.vehicleRegistration]?.expiringSoon ?? false,
          version: docItems[DocumentType.vehicleRegistration]?.version ?? 0,
          onUploaded: onUploaded,
        ),
        const SizedBox(height: AppSpacing.sm),
        DocumentUploadTile(
          label: documentTypeLabel(DocumentType.vehicleInsurance),
          documentType: DocumentType.vehicleInsurance,
          repository: repository,
          uploaded: uploaded[DocumentType.vehicleInsurance] ?? false,
          requiresExpiry: true,
          expiresAt: docItems[DocumentType.vehicleInsurance]?.expiresAt,
          expired: docItems[DocumentType.vehicleInsurance]?.expired ?? false,
          expiringSoon: docItems[DocumentType.vehicleInsurance]?.expiringSoon ?? false,
          version: docItems[DocumentType.vehicleInsurance]?.version ?? 0,
          onUploaded: onUploaded,
        ),
        const SizedBox(height: AppSpacing.sm),
        DocumentUploadTile(
          label: '${documentTypeLabel(DocumentType.vehicleInspection)} (không bắt buộc)',
          documentType: DocumentType.vehicleInspection,
          repository: repository,
          uploaded: uploaded[DocumentType.vehicleInspection] ?? false,
          requiresExpiry: true,
          expiresAt: docItems[DocumentType.vehicleInspection]?.expiresAt,
          expired: docItems[DocumentType.vehicleInspection]?.expired ?? false,
          expiringSoon: docItems[DocumentType.vehicleInspection]?.expiringSoon ?? false,
          version: docItems[DocumentType.vehicleInspection]?.version ?? 0,
          onUploaded: onUploaded,
        ),
      ],
    );
  }
}

// ─── Step 6 — Review ─────────────────────────────────────────────────────────

class _Step6Review extends StatelessWidget {
  const _Step6Review({
    required this.fullName,
    required this.address,
    required this.nationalIdNumber,
    required this.dob,
    required this.vehicleType,
    required this.serviceType,
    required this.brand,
    required this.model,
    required this.plateNumber,
    required this.licenseClass,
    required this.rideEnabled,
    required this.deliveryEnabled,
    required this.error,
  });

  final String fullName;
  final String address;
  final String nationalIdNumber;
  final DateTime? dob;
  final String vehicleType;
  final String serviceType;
  final String brand;
  final String model;
  final String plateNumber;
  final String? licenseClass;
  final bool rideEnabled;
  final bool deliveryEnabled;
  final String? error;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Xem lại & Gửi hồ sơ', style: theme.textTheme.titleLarge),
        const SizedBox(height: AppSpacing.lg),
        _ReviewRow(label: 'Họ tên', value: fullName),
        _ReviewRow(
          label: 'Ngày sinh',
          value: dob == null ? '—' : '${dob!.day.toString().padLeft(2, '0')}/${dob!.month.toString().padLeft(2, '0')}/${dob!.year}',
        ),
        _ReviewRow(label: 'Địa chỉ', value: address),
        _ReviewRow(label: 'Số CCCD', value: nationalIdNumber),
        const Divider(height: AppSpacing.xl),
        _ReviewRow(label: 'Xe', value: '$brand $model'.trim()),
        _ReviewRow(label: 'Biển số', value: plateNumber),
        _ReviewRow(label: 'Hạng bằng', value: licenseClass == null ? '—' : licenseClassLabel(licenseClass!)),
        _ReviewRow(
          label: 'Nhận đơn',
          value: [if (rideEnabled) 'Chở khách', if (deliveryEnabled) 'Giao hàng'].join(', '),
        ),
        if (error != null) ...[
          const SizedBox(height: AppSpacing.md),
          Text(error!, style: const TextStyle(color: AppColors.error)),
        ],
        const SizedBox(height: AppSpacing.md),
        Text(
          'Bằng việc gửi hồ sơ, bạn xác nhận thông tin trên là chính xác. Panda sẽ xét duyệt trong vòng 24–48 giờ.',
          style: theme.textTheme.bodySmall?.copyWith(color: AppColors.textSecondary),
        ),
      ],
    );
  }
}

class _ReviewRow extends StatelessWidget {
  const _ReviewRow({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(width: 110, child: Text(label, style: const TextStyle(color: AppColors.textSecondary))),
          Expanded(child: Text(value.isEmpty ? '—' : value, style: const TextStyle(fontWeight: FontWeight.w600))),
        ],
      ),
    );
  }
}

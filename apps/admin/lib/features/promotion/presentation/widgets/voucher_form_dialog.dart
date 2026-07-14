import 'package:flutter/material.dart';

import '../../data/promotion_repository.dart';
import '../../domain/models/voucher.dart';

/// Create/Edit voucher form. Pops `true` on save so the caller reloads the list.
class VoucherFormDialog extends StatefulWidget {
  const VoucherFormDialog({super.key, required this.repository, this.existing});

  final PromotionRepository repository;
  final Voucher? existing;

  @override
  State<VoucherFormDialog> createState() => _VoucherFormDialogState();
}

const _serviceTypeOptions = ['motorcycle', 'bike_plus', 'car', 'car_xl'];
const _tripTypeOptions = ['ride', 'delivery'];

class _VoucherFormDialogState extends State<VoucherFormDialog> {
  final _formKey = GlobalKey<FormState>();
  late final TextEditingController _code;
  late final TextEditingController _title;
  late final TextEditingController _description;
  late final TextEditingController _value;
  late final TextEditingController _maxDiscount;
  late final TextEditingController _minOrder;
  late final TextEditingController _usageLimit;
  late final TextEditingController _perUserLimit;
  late final TextEditingController _budget;
  late final TextEditingController _campaign;
  late String _type;
  late DateTime _start;
  late DateTime _end;
  late Set<String> _serviceTypes;
  late Set<String> _tripTypes;
  late bool _enabled;
  bool _saving = false;
  String? _error;

  bool get _isEdit => widget.existing != null;

  @override
  void initState() {
    super.initState();
    final v = widget.existing;
    _code = TextEditingController(text: v?.code ?? '');
    _title = TextEditingController(text: v?.title ?? '');
    _description = TextEditingController(text: v?.description ?? '');
    _value = TextEditingController(text: v != null ? v.value.toString() : '10');
    _maxDiscount = TextEditingController(text: v != null ? v.maxDiscount.toString() : '0');
    _minOrder = TextEditingController(text: v != null ? v.minOrder.toString() : '0');
    _usageLimit = TextEditingController(text: v != null ? v.usageLimit.toString() : '0');
    _perUserLimit = TextEditingController(text: v != null ? v.perUserLimit.toString() : '1');
    _budget = TextEditingController(text: v != null ? v.budget.toString() : '1000000');
    _campaign = TextEditingController(text: v?.campaign ?? '');
    _type = v?.type ?? 'percentage';
    _start = v != null && v.start.isNotEmpty ? DateTime.tryParse(v.start) ?? DateTime.now() : DateTime.now();
    _end = v != null && v.end.isNotEmpty
        ? DateTime.tryParse(v.end) ?? DateTime.now().add(const Duration(days: 30))
        : DateTime.now().add(const Duration(days: 30));
    _serviceTypes = {...(v?.serviceType ?? const [])};
    _tripTypes = {...(v?.tripType ?? const [])};
    _enabled = v?.enabled ?? false;
  }

  @override
  void dispose() {
    for (final c in [_code, _title, _description, _value, _maxDiscount, _minOrder, _usageLimit, _perUserLimit, _budget, _campaign]) {
      c.dispose();
    }
    super.dispose();
  }

  Future<void> _pickDate({required bool isStart}) async {
    final initial = isStart ? _start : _end;
    final picked = await showDatePicker(
      context: context,
      initialDate: initial,
      firstDate: DateTime(2020),
      lastDate: DateTime(2100),
    );
    if (picked == null) return;
    setState(() {
      if (isStart) {
        _start = DateTime(picked.year, picked.month, picked.day, _start.hour, _start.minute);
      } else {
        _end = DateTime(picked.year, picked.month, picked.day, 23, 59, 59);
      }
    });
  }

  Future<void> _save() async {
    if (!_formKey.currentState!.validate()) return;
    setState(() {
      _saving = true;
      _error = null;
    });
    final form = {
      'code': _code.text.trim(),
      'title': _title.text.trim(),
      'description': _description.text.trim(),
      'type': _type,
      'value': int.tryParse(_value.text) ?? 0,
      'max_discount': int.tryParse(_maxDiscount.text) ?? 0,
      'min_order': int.tryParse(_minOrder.text) ?? 0,
      'start': _start.toUtc().toIso8601String(),
      'end': _end.toUtc().toIso8601String(),
      'usage_limit': int.tryParse(_usageLimit.text) ?? 0,
      'per_user_limit': int.tryParse(_perUserLimit.text) ?? 0,
      'budget': int.tryParse(_budget.text) ?? 0,
      'service_type': _serviceTypes.toList(),
      'trip_type': _tripTypes.toList(),
      'campaign': _campaign.text.trim(),
      'enabled': _enabled,
    };
    try {
      if (_isEdit) {
        await widget.repository.updateVoucher(widget.existing!.id, form);
      } else {
        await widget.repository.createVoucher(form);
      }
      if (mounted) Navigator.pop(context, true);
    } catch (e) {
      setState(() {
        _saving = false;
        _error = '$e';
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Dialog(
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 520, maxHeight: 640),
        child: Padding(
          padding: const EdgeInsets.all(20),
          child: Form(
            key: _formKey,
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(_isEdit ? 'Sửa voucher' : 'Tạo voucher', style: Theme.of(context).textTheme.titleLarge),
                const SizedBox(height: 12),
                Expanded(
                  child: SingleChildScrollView(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        TextFormField(
                          controller: _code,
                          decoration: const InputDecoration(labelText: 'Mã voucher (để trống = tự động áp dụng)'),
                        ),
                        TextFormField(
                          controller: _title,
                          decoration: const InputDecoration(labelText: 'Tiêu đề'),
                          validator: (v) => (v == null || v.trim().isEmpty) ? 'Bắt buộc' : null,
                        ),
                        TextFormField(
                          controller: _description,
                          decoration: const InputDecoration(labelText: 'Mô tả'),
                        ),
                        TextFormField(
                          controller: _campaign,
                          decoration: const InputDecoration(
                            labelText: 'Campaign (VD: WELCOME, TET2027, AIRPORT, BIRTHDAY, REFERRAL)',
                          ),
                        ),
                        const SizedBox(height: 12),
                        Row(
                          children: [
                            Expanded(
                              child: DropdownButtonFormField<String>(
                                initialValue: _type,
                                decoration: const InputDecoration(labelText: 'Loại giảm giá'),
                                items: const [
                                  DropdownMenuItem(value: 'percentage', child: Text('Phần trăm')),
                                  DropdownMenuItem(value: 'fixed', child: Text('Số tiền cố định')),
                                ],
                                onChanged: (v) => setState(() => _type = v ?? 'percentage'),
                              ),
                            ),
                            const SizedBox(width: 12),
                            Expanded(
                              child: TextFormField(
                                controller: _value,
                                keyboardType: TextInputType.number,
                                decoration: InputDecoration(labelText: _type == 'fixed' ? 'Số tiền (đ)' : 'Phần trăm (%)'),
                                validator: (v) => (int.tryParse(v ?? '') == null) ? 'Số không hợp lệ' : null,
                              ),
                            ),
                          ],
                        ),
                        Row(
                          children: [
                            Expanded(
                              child: TextFormField(
                                controller: _maxDiscount,
                                keyboardType: TextInputType.number,
                                decoration: const InputDecoration(labelText: 'Giảm tối đa (đ, 0 = không giới hạn)'),
                              ),
                            ),
                            const SizedBox(width: 12),
                            Expanded(
                              child: TextFormField(
                                controller: _minOrder,
                                keyboardType: TextInputType.number,
                                decoration: const InputDecoration(labelText: 'Đơn tối thiểu (đ)'),
                              ),
                            ),
                          ],
                        ),
                        Row(
                          children: [
                            Expanded(
                              child: ListTile(
                                contentPadding: EdgeInsets.zero,
                                title: const Text('Bắt đầu'),
                                subtitle: Text(_start.toLocal().toString().split('.').first),
                                onTap: () => _pickDate(isStart: true),
                              ),
                            ),
                            Expanded(
                              child: ListTile(
                                contentPadding: EdgeInsets.zero,
                                title: const Text('Kết thúc'),
                                subtitle: Text(_end.toLocal().toString().split('.').first),
                                onTap: () => _pickDate(isStart: false),
                              ),
                            ),
                          ],
                        ),
                        Row(
                          children: [
                            Expanded(
                              child: TextFormField(
                                controller: _usageLimit,
                                keyboardType: TextInputType.number,
                                decoration: const InputDecoration(labelText: 'Tổng lượt dùng (0 = ∞)'),
                              ),
                            ),
                            const SizedBox(width: 12),
                            Expanded(
                              child: TextFormField(
                                controller: _perUserLimit,
                                keyboardType: TextInputType.number,
                                decoration: const InputDecoration(labelText: 'Lượt/người dùng'),
                              ),
                            ),
                          ],
                        ),
                        TextFormField(
                          controller: _budget,
                          keyboardType: TextInputType.number,
                          decoration: const InputDecoration(labelText: 'Ngân sách chiến dịch (đ)'),
                          validator: (v) => (int.tryParse(v ?? '') ?? 0) <= 0 ? 'Ngân sách phải > 0' : null,
                        ),
                        const SizedBox(height: 8),
                        Text('Loại dịch vụ (bỏ trống = tất cả)', style: Theme.of(context).textTheme.labelMedium),
                        Wrap(
                          spacing: 8,
                          children: _serviceTypeOptions
                              .map((t) => FilterChip(
                                    label: Text(t),
                                    selected: _serviceTypes.contains(t),
                                    onSelected: (sel) => setState(() => sel ? _serviceTypes.add(t) : _serviceTypes.remove(t)),
                                  ))
                              .toList(),
                        ),
                        const SizedBox(height: 8),
                        Text('Loại chuyến (bỏ trống = cả hai)', style: Theme.of(context).textTheme.labelMedium),
                        Wrap(
                          spacing: 8,
                          children: _tripTypeOptions
                              .map((t) => FilterChip(
                                    label: Text(t == 'ride' ? 'Chở khách' : 'Giao hàng'),
                                    selected: _tripTypes.contains(t),
                                    onSelected: (sel) => setState(() => sel ? _tripTypes.add(t) : _tripTypes.remove(t)),
                                  ))
                              .toList(),
                        ),
                        SwitchListTile(
                          contentPadding: EdgeInsets.zero,
                          title: const Text('Kích hoạt ngay'),
                          value: _enabled,
                          onChanged: (v) => setState(() => _enabled = v),
                        ),
                        if (_error != null)
                          Padding(
                            padding: const EdgeInsets.only(top: 8),
                            child: Text(_error!, style: const TextStyle(color: Colors.red)),
                          ),
                      ],
                    ),
                  ),
                ),
                const SizedBox(height: 12),
                Row(
                  mainAxisAlignment: MainAxisAlignment.end,
                  children: [
                    TextButton(onPressed: () => Navigator.pop(context), child: const Text('Hủy')),
                    const SizedBox(width: 8),
                    FilledButton(
                      onPressed: _saving ? null : _save,
                      child: _saving
                          ? const SizedBox(width: 16, height: 16, child: CircularProgressIndicator(strokeWidth: 2))
                          : const Text('Lưu'),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

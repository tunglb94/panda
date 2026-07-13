import 'package:flutter/material.dart';

/// Reject reason picker (Phần 3) — fixed dropdown vocabulary (client-side
/// only; the backend's `reason` field is plain free text) plus a free-text
/// box for "Khác" or extra detail. Returns the composed reason string, or
/// null if the admin cancelled.
class RejectReasonDialog extends StatefulWidget {
  const RejectReasonDialog({super.key});

  static const _options = [
    'Ảnh mờ',
    'Sai GPLX',
    'Sai CCCD',
    'Không trùng Selfie',
    'Xe không đúng',
    'Khác',
  ];

  static Future<String?> show(BuildContext context) {
    return showDialog<String>(context: context, builder: (_) => const RejectReasonDialog());
  }

  @override
  State<RejectReasonDialog> createState() => _RejectReasonDialogState();
}

class _RejectReasonDialogState extends State<RejectReasonDialog> {
  String _selected = RejectReasonDialog._options.first;
  final _detailCtrl = TextEditingController();

  @override
  void dispose() {
    _detailCtrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: const Text('Lý do từ chối'),
      content: SizedBox(
        width: 360,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            DropdownButtonFormField<String>(
              initialValue: _selected,
              items: RejectReasonDialog._options
                  .map((o) => DropdownMenuItem(value: o, child: Text(o)))
                  .toList(),
              onChanged: (v) => setState(() => _selected = v ?? _selected),
              decoration: const InputDecoration(labelText: 'Chọn lý do'),
            ),
            const SizedBox(height: 12),
            TextField(
              controller: _detailCtrl,
              maxLines: 3,
              decoration: const InputDecoration(labelText: 'Ghi chú thêm (tuỳ chọn)', alignLabelWithHint: true),
            ),
          ],
        ),
      ),
      actions: [
        TextButton(onPressed: () => Navigator.of(context).pop(), child: const Text('Huỷ')),
        FilledButton(
          onPressed: () {
            final detail = _detailCtrl.text.trim();
            final reason = detail.isEmpty ? _selected : '$_selected: $detail';
            Navigator.of(context).pop(reason);
          },
          child: const Text('Từ chối'),
        ),
      ],
    );
  }
}

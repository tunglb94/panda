import 'dart:typed_data';
import 'package:flutter/material.dart';

import '../../data/kyc_repository.dart';
import '../../domain/models/kyc_document_item.dart';
import 'fullscreen_image_viewer.dart';

/// One checklist entry in the review Drawer's document grid (Phần 2/6/13).
/// Lazily fetches the image bytes (auth-gated — never a plain `<img src>`)
/// only when [item] is uploaded, then shows a thumbnail; tapping opens the
/// fullscreen zoom/pan viewer.
class DocumentThumbnail extends StatefulWidget {
  const DocumentThumbnail({super.key, required this.repository, required this.item});

  final KYCRepository repository;
  final KYCDocumentItem item;

  @override
  State<DocumentThumbnail> createState() => _DocumentThumbnailState();
}

class _DocumentThumbnailState extends State<DocumentThumbnail> {
  Future<Uint8List>? _bytesFuture;

  @override
  void initState() {
    super.initState();
    final id = widget.item.documentId;
    if (id != null) _bytesFuture = widget.repository.getDocumentBytes(id);
  }

  @override
  Widget build(BuildContext context) {
    final label = kDocumentTypeLabels[widget.item.documentType] ?? widget.item.documentType;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(label, style: const TextStyle(fontSize: 12, fontWeight: FontWeight.w600)),
        const SizedBox(height: 4),
        AspectRatio(
          aspectRatio: 4 / 3,
          child: DecoratedBox(
            decoration: BoxDecoration(
              color: const Color(0xFFF3F4F6),
              borderRadius: BorderRadius.circular(8),
              border: Border.all(color: const Color(0xFFE5E7EB)),
            ),
            child: !widget.item.uploaded
                ? const Center(
                    child: Icon(Icons.image_not_supported_outlined, color: Color(0xFF9CA3AF)),
                  )
                : ClipRRect(
                    borderRadius: BorderRadius.circular(8),
                    child: FutureBuilder<Uint8List>(
                      future: _bytesFuture,
                      builder: (context, snapshot) {
                        if (!snapshot.hasData) {
                          return const Center(child: CircularProgressIndicator(strokeWidth: 2));
                        }
                        final bytes = snapshot.data!;
                        return InkWell(
                          onTap: () => FullscreenImageViewer.show(context, bytes: bytes, title: label),
                          child: Image.memory(bytes, fit: BoxFit.cover, width: double.infinity),
                        );
                      },
                    ),
                  ),
          ),
        ),
        if (widget.item.uploaded) ...[
          const SizedBox(height: 2),
          Text(
            'v${widget.item.version ?? 1}'
            '${widget.item.expired == true ? ' • Hết hạn' : widget.item.expiringSoon == true ? ' • Sắp hết hạn' : ''}',
            style: TextStyle(
              fontSize: 11,
              color: widget.item.expired == true ? const Color(0xFFDC2626) : const Color(0xFF6B7280),
            ),
          ),
        ],
      ],
    );
  }
}

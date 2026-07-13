import 'dart:typed_data';
import 'package:flutter/material.dart';

import '../../data/kyc_repository.dart';

/// Small circular selfie thumbnail for the list table (Phần 6). Lazily
/// fetches only when this row is actually built — paired with
/// `ListView.builder` in the parent so off-screen rows never fetch.
class AvatarThumbnail extends StatefulWidget {
  const AvatarThumbnail({super.key, required this.repository, required this.driverId});

  final KYCRepository repository;
  final String driverId;

  @override
  State<AvatarThumbnail> createState() => _AvatarThumbnailState();
}

class _AvatarThumbnailState extends State<AvatarThumbnail> {
  late Future<Uint8List?> _future;

  @override
  void initState() {
    super.initState();
    _future = _load();
  }

  Future<Uint8List?> _load() async {
    final selfie = await widget.repository.getSelfieDocument(widget.driverId);
    if (selfie?.documentId == null) return null;
    try {
      return await widget.repository.getDocumentBytes(selfie!.documentId!);
    } catch (_) {
      return null;
    }
  }

  @override
  Widget build(BuildContext context) {
    return FutureBuilder<Uint8List?>(
      future: _future,
      builder: (context, snapshot) {
        final bytes = snapshot.data;
        return CircleAvatar(
          radius: 18,
          backgroundColor: const Color(0xFFE5E7EB),
          backgroundImage: bytes != null ? MemoryImage(bytes) : null,
          child: bytes == null ? const Icon(Icons.person_outline, size: 18, color: Color(0xFF9CA3AF)) : null,
        );
      },
    );
  }
}

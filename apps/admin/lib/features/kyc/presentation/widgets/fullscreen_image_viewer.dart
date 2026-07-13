import 'dart:typed_data';
import 'package:flutter/material.dart';

/// Fullscreen zoom/pan viewer (Phần 13). Uses [InteractiveViewer], built
/// into the Flutter SDK — no new package needed.
class FullscreenImageViewer extends StatelessWidget {
  const FullscreenImageViewer({super.key, required this.bytes, required this.title});

  final Uint8List bytes;
  final String title;

  static Future<void> show(BuildContext context, {required Uint8List bytes, required String title}) {
    return Navigator.of(context).push(
      PageRouteBuilder(
        opaque: false,
        barrierColor: Colors.black87,
        pageBuilder: (context, _, _) => FullscreenImageViewer(bytes: bytes, title: title),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.black87,
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        foregroundColor: Colors.white,
        title: Text(title),
      ),
      body: Center(
        child: InteractiveViewer(
          minScale: 0.5,
          maxScale: 5,
          child: Image.memory(bytes, fit: BoxFit.contain),
        ),
      ),
    );
  }
}

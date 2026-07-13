// ignore_for_file: deprecated_member_use, avoid_web_libraries_in_flutter
import 'dart:html' as html;
import 'dart:typed_data';

/// Triggers a browser "Save As" download of [bytes] as [filename]. Uses
/// `dart:html`'s Blob + anchor-click trick — safe here since apps/admin is
/// Flutter-Web-only (no other platform target configured for this project).
void downloadBytes(Uint8List bytes, String filename) {
  final blob = html.Blob([bytes]);
  final url = html.Url.createObjectUrlFromBlob(blob);
  html.AnchorElement(href: url)
    ..setAttribute('download', filename)
    ..click();
  html.Url.revokeObjectUrl(url);
}

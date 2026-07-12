import 'dart:convert';

import 'package:google_maps_flutter/google_maps_flutter.dart';
import 'package:http/http.dart' as http;

import 'place_suggestion.dart';

class PlacesServiceException implements Exception {
  const PlacesServiceException(this.message);

  final String message;

  @override
  String toString() => message;
}

/// Address search backed by OpenStreetMap's Nominatim — no API key, no
/// billing account, no Google Cloud setup required.
///
/// Replaces the former `GooglePlacesService` (Google Places Autocomplete +
/// Details), which needs a billing-enabled Google Cloud project. Getting
/// that billing account working hit a persistent, unresolved Google-side
/// issue (see chat history 2026-07-10), so the app searches addresses via
/// Nominatim instead. The native map itself (tiles/markers) is unaffected —
/// it still renders through `google_maps_flutter`, which uses the separate
/// Maps SDK for Android key already working in `AndroidManifest.xml`.
///
/// Usage policy: the public Nominatim instance asks for a descriptive
/// User-Agent and reasonable request volume (nominatim.org/release-docs/latest/api/Search/).
/// Fine for dev/test and light production use; a high-volume production
/// deployment should self-host Nominatim or use a paid OSM-backed provider
/// (e.g. Goong.io, LocationIQ) — swap the base URLs below when that's needed.
class NominatimPlacesService {
  const NominatimPlacesService();

  static const _searchUrl = 'https://nominatim.openstreetmap.org/search';
  static const _reverseUrl = 'https://nominatim.openstreetmap.org/reverse';
  static const _userAgent = 'PandaRiderApp/1.0 (+https://github.com/fairride)';

  Future<List<PlaceSuggestion>> autocomplete(String input, {LatLng? bias}) async {
    if (input.trim().isEmpty) return const [];
    final params = <String, String>{
      'q': input,
      'format': 'jsonv2',
      'addressdetails': '0',
      'limit': '6',
      'countrycodes': 'vn',
      'accept-language': 'vi',
    };
    // Nominatim has no radius-bias search; a loose viewbox around the rider's
    // current position nudges ranking toward nearby results without hard
    // excluding anything outside it (bounded=0).
    if (bias != null) {
      const delta = 0.5; // ~50km box
      params['viewbox'] =
          '${bias.longitude - delta},${bias.latitude + delta},${bias.longitude + delta},${bias.latitude - delta}';
      params['bounded'] = '0';
    }

    final uri = Uri.parse(_searchUrl).replace(queryParameters: params);
    final response = await http.get(uri, headers: {'User-Agent': _userAgent});
    if (response.statusCode != 200) {
      throw PlacesServiceException('Lỗi tìm kiếm địa điểm (${response.statusCode}).');
    }
    final results = jsonDecode(response.body) as List<dynamic>;
    return results.map((r) {
      final item = r as Map<String, dynamic>;
      final lat = item['lat'] as String;
      final lon = item['lon'] as String;
      final displayName = item['display_name'] as String? ?? '';
      final name = item['name'] as String?;
      final (mainText, secondaryText) = _splitDisplayName(displayName, name);
      return PlaceSuggestion(
        // Encodes the coordinate directly — Nominatim's search response
        // already includes lat/lon, so details() doesn't need to look the
        // place back up by an opaque ID the way Google Places does.
        placeId: '$lat,$lon',
        mainText: mainText,
        secondaryText: secondaryText,
      );
    }).toList();
  }

  Future<PlaceDetailsResult> details(String placeId) async {
    final parts = placeId.split(',');
    if (parts.length != 2) {
      throw const PlacesServiceException('Địa điểm không hợp lệ.');
    }
    final lat = double.tryParse(parts[0]);
    final lon = double.tryParse(parts[1]);
    if (lat == null || lon == null) {
      throw const PlacesServiceException('Địa điểm không hợp lệ.');
    }

    final uri = Uri.parse(_reverseUrl).replace(queryParameters: {
      'lat': parts[0],
      'lon': parts[1],
      'format': 'jsonv2',
      'accept-language': 'vi',
    });
    final response = await http.get(uri, headers: {'User-Agent': _userAgent});
    if (response.statusCode != 200) {
      throw PlacesServiceException('Lỗi tải chi tiết địa điểm (${response.statusCode}).');
    }
    final data = jsonDecode(response.body) as Map<String, dynamic>;
    final displayName = data['display_name'] as String? ?? '';
    return PlaceDetailsResult(
      location: LatLng(lat, lon),
      formattedAddress: displayName,
    );
  }

  /// Nominatim doesn't separate "name" from "rest of address" the way
  /// Google's structured_formatting does — this derives the same shape:
  /// short label up front, fuller address as the secondary line.
  (String, String) _splitDisplayName(String displayName, String? name) {
    if (name != null && name.isNotEmpty && displayName.startsWith(name)) {
      final rest = displayName.substring(name.length).replaceFirst(RegExp(r'^,\s*'), '');
      return (name, rest);
    }
    final parts = displayName.split(', ');
    if (parts.isEmpty) return (displayName, '');
    return (parts.first, parts.skip(1).join(', '));
  }
}

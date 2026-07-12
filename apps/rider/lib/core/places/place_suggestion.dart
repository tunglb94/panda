import 'package:google_maps_flutter/google_maps_flutter.dart';

/// A single address/place search result.
class PlaceSuggestion {
  const PlaceSuggestion({
    required this.placeId,
    required this.mainText,
    required this.secondaryText,
  });

  final String placeId;
  final String mainText;
  final String secondaryText;

  String get description =>
      secondaryText.isEmpty ? mainText : '$mainText, $secondaryText';
}

/// Resolved coordinates + formatted address for a [PlaceSuggestion], fetched
/// via the Places Details API once the rider picks a suggestion.
class PlaceDetailsResult {
  const PlaceDetailsResult({required this.location, required this.formattedAddress});

  final LatLng location;
  final String formattedAddress;
}

import 'dart:async';

import 'package:flutter/material.dart';
import 'package:google_maps_flutter/google_maps_flutter.dart';

import 'package:rider/core/places/nominatim_places_service.dart';
import 'package:rider/core/places/place_suggestion.dart';

/// Text field with live address suggestions (OpenStreetMap Nominatim) — lets
/// the rider type an address or landmark name (e.g. "Chợ Bến Thành") instead
/// of having to drag the map pin to the right spot. Picking a suggestion
/// resolves its coordinates and reports both through [onSelected].
class PlaceSearchField extends StatefulWidget {
  const PlaceSearchField({
    super.key,
    required this.placesService,
    required this.hintText,
    required this.onSelected,
    this.biasCenter,
    this.initialText,
  });

  final NominatimPlacesService placesService;
  final String hintText;
  final LatLng? biasCenter;
  final void Function(String address, LatLng location) onSelected;

  /// Pre-fills the field — e.g. after the rider picks a location via
  /// [MapAddressPickerPage] rather than by typing. Only read once, at
  /// construction (see this widget's own internal `_controller`, which
  /// owns keystroke-by-keystroke state) — callers that need to update the
  /// text after a discrete external event (not typing) should pass a new
  /// `key` (e.g. `ValueKey(initialText)`) so Flutter treats it as a fresh
  /// instance instead of trying to diff-update the live controller.
  final String? initialText;

  @override
  State<PlaceSearchField> createState() => _PlaceSearchFieldState();
}

class _PlaceSearchFieldState extends State<PlaceSearchField> {
  late final _controller = TextEditingController(text: widget.initialText);
  Timer? _debounce;
  List<PlaceSuggestion> _suggestions = [];
  bool _searching = false;
  bool _resolving = false;
  String? _error;

  @override
  void dispose() {
    _debounce?.cancel();
    _controller.dispose();
    super.dispose();
  }

  void _onChanged(String value) {
    _debounce?.cancel();
    setState(() => _error = null);
    final query = value.trim();
    if (query.length < 2) {
      setState(() => _suggestions = []);
      return;
    }
    _debounce = Timer(const Duration(milliseconds: 350), () => _search(query));
  }

  Future<void> _search(String query) async {
    setState(() => _searching = true);
    try {
      final results =
          await widget.placesService.autocomplete(query, bias: widget.biasCenter);
      if (!mounted) return;
      setState(() {
        _suggestions = results;
        _searching = false;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() {
        _searching = false;
        _suggestions = [];
        _error = 'Không thể tìm kiếm địa điểm. Vui lòng thử lại.';
      });
    }
  }

  Future<void> _select(PlaceSuggestion suggestion) async {
    FocusScope.of(context).unfocus();
    setState(() {
      _suggestions = [];
      _resolving = true;
      _controller.text = suggestion.description;
      _controller.selection =
          TextSelection.collapsed(offset: _controller.text.length);
    });
    try {
      final details = await widget.placesService.details(suggestion.placeId);
      if (!mounted) return;
      setState(() => _resolving = false);
      widget.onSelected(
        details.formattedAddress.isNotEmpty
            ? details.formattedAddress
            : suggestion.description,
        details.location,
      );
    } catch (_) {
      if (!mounted) return;
      setState(() {
        _resolving = false;
        _error = 'Không thể lấy vị trí của địa điểm này.';
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        TextField(
          controller: _controller,
          onChanged: _onChanged,
          textInputAction: TextInputAction.search,
          decoration: InputDecoration(
            hintText: widget.hintText,
            isDense: true,
            prefixIcon: const Icon(Icons.search, size: 20),
            suffixIcon: (_searching || _resolving)
                ? const Padding(
                    padding: EdgeInsets.all(14),
                    child: SizedBox(
                      width: 16,
                      height: 16,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    ),
                  )
                : (_controller.text.isNotEmpty
                    ? IconButton(
                        icon: const Icon(Icons.clear, size: 18),
                        onPressed: () => setState(() {
                          _controller.clear();
                          _suggestions = [];
                        }),
                      )
                    : null),
            contentPadding:
                const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
          ),
        ),
        if (_suggestions.isNotEmpty)
          Container(
            margin: const EdgeInsets.only(top: 6),
            constraints: const BoxConstraints(maxHeight: 220),
            decoration: BoxDecoration(
              borderRadius: BorderRadius.circular(12),
              border: Border.all(color: Colors.grey.shade200),
            ),
            child: ListView.separated(
              padding: EdgeInsets.zero,
              shrinkWrap: true,
              itemCount: _suggestions.length,
              separatorBuilder: (_, _) => const Divider(height: 1),
              itemBuilder: (context, index) {
                final s = _suggestions[index];
                return ListTile(
                  dense: true,
                  leading: const Icon(Icons.place_outlined),
                  title: Text(s.mainText, maxLines: 1, overflow: TextOverflow.ellipsis),
                  subtitle: s.secondaryText.isEmpty
                      ? null
                      : Text(s.secondaryText,
                          maxLines: 1, overflow: TextOverflow.ellipsis),
                  onTap: () => _select(s),
                );
              },
            ),
          ),
        if (_error != null)
          Padding(
            padding: const EdgeInsets.only(top: 4, left: 4),
            child: Text(
              _error!,
              style:
                  TextStyle(color: Theme.of(context).colorScheme.error, fontSize: 12),
            ),
          ),
      ],
    );
  }
}

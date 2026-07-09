import 'package:shared_preferences/shared_preferences.dart';

class TripStorage {
  static const _keyActiveTripId = 'active_trip_id';

  Future<void> saveActiveTripId(String tripId) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_keyActiveTripId, tripId);
  }

  Future<String?> loadActiveTripId() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString(_keyActiveTripId);
  }

  Future<void> clearActiveTripId() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_keyActiveTripId);
  }
}

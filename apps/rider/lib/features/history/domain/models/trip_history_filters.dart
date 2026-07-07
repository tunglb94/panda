/// Status filter chips on the Trip History screen.
enum TripHistoryStatusFilter { all, completed, cancelled }

extension TripHistoryStatusFilterX on TripHistoryStatusFilter {
  String get label => switch (this) {
        TripHistoryStatusFilter.all => 'All',
        TripHistoryStatusFilter.completed => 'Completed',
        TripHistoryStatusFilter.cancelled => 'Cancelled',
      };
}

/// Date-range filter chips on the Trip History screen.
enum TripHistoryDateFilter { all, today, thisWeek, thisMonth }

extension TripHistoryDateFilterX on TripHistoryDateFilter {
  String get label => switch (this) {
        TripHistoryDateFilter.all => 'All time',
        TripHistoryDateFilter.today => 'Today',
        TripHistoryDateFilter.thisWeek => 'This Week',
        TripHistoryDateFilter.thisMonth => 'This Month',
      };

  /// Whether [dateTime] falls inside this range, evaluated against "now".
  bool matches(DateTime dateTime) {
    final now = DateTime.now();
    switch (this) {
      case TripHistoryDateFilter.all:
        return true;
      case TripHistoryDateFilter.today:
        return dateTime.year == now.year &&
            dateTime.month == now.month &&
            dateTime.day == now.day;
      case TripHistoryDateFilter.thisWeek:
        return now.difference(dateTime).inDays < 7;
      case TripHistoryDateFilter.thisMonth:
        return dateTime.year == now.year && dateTime.month == now.month;
    }
  }
}

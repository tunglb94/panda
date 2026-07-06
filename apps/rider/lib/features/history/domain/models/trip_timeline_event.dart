import 'package:flutter/material.dart';

/// A single milestone in a past trip's timeline (Trip Detail screen).
class TripTimelineEvent {
  const TripTimelineEvent({
    required this.label,
    required this.time,
    required this.icon,
  });

  final String label;
  final DateTime time;
  final IconData icon;
}

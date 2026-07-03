import 'package:flutter/material.dart';

class BookingPage extends StatefulWidget {
  const BookingPage({super.key});

  @override
  State<BookingPage> createState() => _BookingPageState();
}

class _BookingPageState extends State<BookingPage> {
  int _selectedVehicle = 0;

  static const List<_VehicleOption> _vehicles = [
    _VehicleOption(
        label: 'Car',
        icon: Icons.directions_car,
        fare: '\$5.00–8.00',
        eta: '3 min'),
    _VehicleOption(
        label: 'Moto',
        icon: Icons.two_wheeler,
        fare: '\$2.50–4.00',
        eta: '2 min'),
    _VehicleOption(
        label: 'Van',
        icon: Icons.airport_shuttle,
        fare: '\$9.00–14.00',
        eta: '5 min'),
  ];

  @override
  Widget build(BuildContext context) {
    final textTheme = Theme.of(context).textTheme;
    return Scaffold(
      appBar: AppBar(title: const Text('Book a Ride')),
      body: Column(
        children: [
          Expanded(
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const _AddressCard(
                    label: 'Pickup',
                    address: '123 Main Street',
                    dotColor: Color(0xFF1A8C4E),
                  ),
                  const SizedBox(height: 8),
                  const _AddressCard(
                    label: 'Dropoff',
                    address: 'Enter destination',
                    dotColor: Colors.red,
                  ),
                  const SizedBox(height: 24),
                  Text(
                    'Choose a ride',
                    style: textTheme.titleMedium
                        ?.copyWith(fontWeight: FontWeight.w600),
                  ),
                  const SizedBox(height: 12),
                  ...List.generate(_vehicles.length, (i) {
                    return _VehicleCard(
                      option: _vehicles[i],
                      isSelected: _selectedVehicle == i,
                      onTap: () => setState(() => _selectedVehicle = i),
                    );
                  }),
                  const SizedBox(height: 16),
                  _FareBreakdown(vehicle: _vehicles[_selectedVehicle]),
                ],
              ),
            ),
          ),
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 8, 16, 36),
            child: ElevatedButton(
              onPressed: () => _showConfirmSheet(context),
              child: const Text('Confirm Booking'),
            ),
          ),
        ],
      ),
    );
  }

  void _showConfirmSheet(BuildContext context) {
    final vehicle = _vehicles[_selectedVehicle];
    showModalBottomSheet<void>(
      context: context,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (_) => Padding(
        padding: const EdgeInsets.fromLTRB(24, 24, 24, 40),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              width: 40,
              height: 4,
              decoration: BoxDecoration(
                color: Colors.grey.shade300,
                borderRadius: BorderRadius.circular(2),
              ),
            ),
            const SizedBox(height: 20),
            const Text(
              'Confirm your ride?',
              style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 8),
            Text(
              '${vehicle.label} · ${vehicle.fare} · ETA ${vehicle.eta}',
              style: TextStyle(color: Colors.grey.shade600),
            ),
            const SizedBox(height: 24),
            ElevatedButton(
              onPressed: () => Navigator.pop(context),
              child: const Text('Confirm'),
            ),
            const SizedBox(height: 8),
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: const Text('Cancel'),
            ),
          ],
        ),
      ),
    );
  }
}

class _AddressCard extends StatelessWidget {
  const _AddressCard({
    required this.label,
    required this.address,
    required this.dotColor,
  });

  final String label;
  final String address;
  final Color dotColor;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey.shade200),
      ),
      child: Row(
        children: [
          Icon(Icons.circle, size: 12, color: dotColor),
          const SizedBox(width: 12),
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                label,
                style: TextStyle(fontSize: 11, color: Colors.grey.shade500),
              ),
              Text(
                address,
                style: const TextStyle(fontWeight: FontWeight.w500),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class _VehicleOption {
  const _VehicleOption({
    required this.label,
    required this.icon,
    required this.fare,
    required this.eta,
  });

  final String label;
  final IconData icon;
  final String fare;
  final String eta;
}

class _VehicleCard extends StatelessWidget {
  const _VehicleCard({
    required this.option,
    required this.isSelected,
    required this.onTap,
  });

  final _VehicleOption option;
  final bool isSelected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    const primary = Color(0xFF1A8C4E);
    return GestureDetector(
      onTap: onTap,
      child: Container(
        margin: const EdgeInsets.only(bottom: 8),
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
        decoration: BoxDecoration(
          color: isSelected
              ? primary.withValues(alpha: 0.05)
              : Colors.white,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(
            color: isSelected ? primary : Colors.grey.shade200,
            width: isSelected ? 1.5 : 1,
          ),
        ),
        child: Row(
          children: [
            Icon(option.icon, color: primary, size: 28),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(option.label,
                      style: const TextStyle(fontWeight: FontWeight.w600)),
                  Text(
                    'ETA ${option.eta}',
                    style: TextStyle(
                        fontSize: 12, color: Colors.grey.shade500),
                  ),
                ],
              ),
            ),
            Text(option.fare,
                style: const TextStyle(fontWeight: FontWeight.w600)),
          ],
        ),
      ),
    );
  }
}

class _FareBreakdown extends StatelessWidget {
  const _FareBreakdown({required this.vehicle});

  final _VehicleOption vehicle;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.grey.shade50,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey.shade200),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text('Fare breakdown',
              style: TextStyle(fontWeight: FontWeight.w600)),
          const SizedBox(height: 12),
          const _FareRow(label: 'Base fare', value: '\$2.00'),
          const _FareRow(label: 'Distance (est.)', value: '\$2.50'),
          const _FareRow(label: 'Booking fee', value: '\$0.50'),
          const Divider(height: 20),
          _FareRow(
              label: 'Estimated total', value: vehicle.fare, bold: true),
        ],
      ),
    );
  }
}

class _FareRow extends StatelessWidget {
  const _FareRow({
    required this.label,
    required this.value,
    this.bold = false,
  });

  final String label;
  final String value;
  final bool bold;

  @override
  Widget build(BuildContext context) {
    final style = TextStyle(
        fontWeight: bold ? FontWeight.bold : FontWeight.normal);
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 2),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label, style: style),
          Text(value, style: style),
        ],
      ),
    );
  }
}

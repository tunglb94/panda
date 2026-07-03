import 'package:flutter/material.dart';
import 'package:rider/core/router/app_router.dart';
import 'package:rider/core/theme/app_theme.dart';

class RiderApp extends StatelessWidget {
  const RiderApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp.router(
      title: 'FAIRRIDE',
      theme: AppTheme.light,
      routerConfig: AppRouter.router,
      debugShowCheckedModeBanner: false,
    );
  }
}

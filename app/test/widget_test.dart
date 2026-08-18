import 'package:flutter_test/flutter_test.dart';

import 'package:nibtara_app/main.dart';

void main() {
  test('aging buckets by days outstanding', () {
    final today = DateTime(2026, 8, 18);
    expect(Due('A', 100, '2026-08-10').bucket(today), 'current');
    expect(Due('A', 100, '2026-01-10').bucket(today), 'stale');
  });

  test('settlement offers include a 10% waiver', () {
    final opts = settlementOptions(1000);
    expect(opts.length, 3);
    expect(opts[1].value, closeTo(900, 1e-9));
  });

  testWidgets('renders outstanding header', (tester) async {
    await tester.pumpWidget(const NibtaraApp());
    expect(find.textContaining('Outstanding'), findsOneWidget);
  });
}

import 'package:flutter/material.dart';

void main() => runApp(const NibtaraApp());

/// Nibtara — dues-aging register + settlement slip. No interest: an aging board
/// plus 2–3 waiver/instalment offers per due. Mirrors the Go service.
class NibtaraApp extends StatelessWidget {
  const NibtaraApp({super.key});
  @override
  Widget build(BuildContext context) => MaterialApp(
        title: 'Nibtara',
        debugShowCheckedModeBanner: false,
        theme: ThemeData(colorSchemeSeed: const Color(0xFF8E6E3E), useMaterial3: true),
        home: const HomePage(),
      );
}

class Due {
  final String customer, sinceDate;
  final double amount;
  Due(this.customer, this.amount, this.sinceDate);

  int daysOutstanding(DateTime today) {
    final since = DateTime.tryParse(sinceDate);
    if (since == null) return 0;
    return today.difference(since).inDays;
  }

  String bucket(DateTime today) {
    final d = daysOutstanding(today);
    if (d > 90) return 'stale';
    if (d >= 60) return 'aging';
    return 'current';
  }
}

/// settlementOptions mirrors backend/cost.go.
List<MapEntry<String, double>> settlementOptions(double amount) => [
      MapEntry('Pay in full now', amount),
      MapEntry('Clear now, 10% goodwill waiver', amount * 0.90),
      MapEntry('3 monthly instalments', amount / 3),
    ];

class HomePage extends StatefulWidget {
  const HomePage({super.key});
  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  final _dues = <Due>[];
  final _cust = TextEditingController();
  final _amt = TextEditingController();
  final _since = TextEditingController(text: '2026-05-01');

  void _add() {
    final a = double.tryParse(_amt.text.trim()) ?? 0;
    if (_cust.text.trim().isEmpty || a <= 0) return;
    setState(() {
      _dues.insert(0, Due(_cust.text.trim(), a, _since.text.trim()));
      _cust.clear();
      _amt.clear();
    });
  }

  @override
  Widget build(BuildContext context) {
    final now = DateTime.now();
    final total = _dues.fold(0.0, (s, d) => s + d.amount);
    String m(double v) => '₹${v.toStringAsFixed(2)}';
    return Scaffold(
      appBar: AppBar(
        title: const Text('Nibtara · dues & settlement'),
        backgroundColor: Theme.of(context).colorScheme.primaryContainer,
      ),
      body: Column(children: [
        Container(
          width: double.infinity,
          color: Theme.of(context).colorScheme.primaryContainer,
          padding: const EdgeInsets.all(14),
          child: Text('Outstanding ${m(total)}', style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold)),
        ),
        Padding(padding: const EdgeInsets.all(12), child: Row(children: [
          Expanded(child: TextField(controller: _cust, decoration: const InputDecoration(labelText: 'Customer', border: OutlineInputBorder()))),
          const SizedBox(width: 8),
          SizedBox(width: 90, child: TextField(controller: _amt, keyboardType: TextInputType.number, decoration: const InputDecoration(labelText: '₹', border: OutlineInputBorder()))),
          const SizedBox(width: 8),
          SizedBox(width: 130, child: TextField(controller: _since, decoration: const InputDecoration(labelText: 'Since', border: OutlineInputBorder()))),
          const SizedBox(width: 8),
          FilledButton(onPressed: _add, child: const Text('Add')),
        ])),
        const Divider(),
        Expanded(child: ListView.builder(
          itemCount: _dues.length,
          itemBuilder: (_, i) {
            final d = _dues[i];
            final b = d.bucket(now);
            return ExpansionTile(
              leading: Icon(Icons.person, color: b == 'stale' ? Colors.red : (b == 'aging' ? Colors.orange : null)),
              title: Text('${d.customer} · ${m(d.amount)}'),
              subtitle: Text('$b · ${d.daysOutstanding(now)} days'),
              children: [
                for (final o in settlementOptions(d.amount))
                  ListTile(dense: true, title: Text(o.key), trailing: Text(m(o.value))),
              ],
            );
          },
        )),
      ]),
    );
  }
}

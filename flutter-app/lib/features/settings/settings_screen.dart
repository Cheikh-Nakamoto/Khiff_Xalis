import 'package:flutter/material.dart';

/// Settings screen — user preferences and app info.
class SettingsScreen extends StatelessWidget {
  const SettingsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text(
          'Parametres',
          style: TextStyle(fontWeight: FontWeight.bold),
        ),
      ),
      body: ListView(
        children: [
          const ListTile(
            leading: Icon(Icons.person),
            title: Text('Profil'),
            trailing: Icon(Icons.chevron_right),
          ),
          const ListTile(
            leading: Icon(Icons.notifications),
            title: Text('Notifications'),
            trailing: Icon(Icons.chevron_right),
          ),
          const ListTile(
            leading: Icon(Icons.security),
            title: Text('Securite'),
            trailing: Icon(Icons.chevron_right),
          ),
          const ListTile(
            leading: Icon(Icons.account_balance),
            title: Text('Compte SGI'),
            trailing: Icon(Icons.chevron_right),
          ),
          const Divider(),
          ListTile(
            leading: const Icon(Icons.dark_mode),
            title: const Text('Mode Sombre'),
            trailing: Switch(value: false, onChanged: (_) {}),
          ),
          const ListTile(
            leading: Icon(Icons.language),
            title: Text('Langue'),
            trailing: Text('Francais'),
          ),
          const Divider(),
          const ListTile(
            leading: Icon(Icons.info),
            title: Text('A propos'),
            trailing: Text('v1.0.0'),
          ),
        ],
      ),
    );
  }
}

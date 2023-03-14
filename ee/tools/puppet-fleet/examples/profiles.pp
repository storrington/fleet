$template = @(END)
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>PayloadDescription</key>
    <string
      >This profile configuration is designed to apply the CIS Benchmark for
      macOS 10.14 (v2.0.0), 10.15 (v2.0.0), 11.0 (v2.0.0), and 12.0
      (v1.0.0)</string
    >
    <key>PayloadDisplayName</key>
    <string>CIS - Bluetooth Sharing</string>
    <key>PayloadEnabled</key>
    <true />
    <key>PayloadIdentifier</key>
    <string>cis.macOSBenchmark.section2.BluetoothSharing</string>
    <key>PayloadScope</key>
    <string>System</string>
    <key>PayloadType</key>
    <string>Configuration</string>
    <key>PayloadUUID</key>
    <string>5CEBD712-28EB-432B-84C7-AA28A5A383D8</string>
    <key>PayloadVersion</key>
    <integer>1</integer>
    <key>PayloadRemovalDisallowed</key>
    <true />
    <key>PayloadContent</key>
    <array>
      <dict>
        <key>PayloadContent</key>
        <dict>
          <key>com.apple.Bluetooth</key>
          <dict>
            <key>Forced</key>
            <array>
              <dict>
                <key>mcx_preference_settings</key>
                <dict>
                  <key>PrefKeyServicesEnabled</key>
                  <false />
                </dict>
              </dict>
            </array>
          </dict>
        </dict>
        <key>PayloadDescription</key>
        <string>Disables Bluetooth Sharing</string>
        <key>PayloadDisplayName</key>
        <string>Custom</string>
        <key>PayloadEnabled</key>
        <true />
        <key>PayloadIdentifier</key>
        <string>0240DD1C-70DC-4766-9018-04322BFEEAD1</string>
        <key>PayloadType</key>
        <string>com.apple.ManagedClient.preferences</string>
        <key>PayloadUUID</key>
        <string>0240DD1C-70DC-4766-9018-04322BFEEAD1</string>
        <key>PayloadVersion</key>
        <integer>1</integer>
      </dict>
    </array>
  </dict>
</plist>
END

node default {
  fleet::add_to_team{ 'Workstations': }

  fleet::add_profiles {'Workstations':
    profiles => [
      inline_template($template)
    ]
  }

#  fleet::with_team { 'Workstations':
#    profiles => [
#      profile::cis_bt_sharing,
#    ]
#  }
}

package apple_mdm

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/fleetdm/fleet/v4/server/contexts/ctxerr"
	"github.com/fleetdm/fleet/v4/server/mdm/apple/appmanifest"
	"github.com/fleetdm/fleet/v4/server/mdm/apple/mobileconfig"
	"github.com/groob/plist"
	"github.com/micromdm/nanomdm/mdm"
	nanomdm_push "github.com/micromdm/nanomdm/push"
	nanomdm_storage "github.com/micromdm/nanomdm/storage"
)

// commandPayload is the common structure all MDM commands use
type commandPayload struct {
	CommandUUID string
	Command     any
}

// MDMAppleCommander contains methods to enqueue commands managed by Fleet and
// send push notifications to hosts.
//
// It's intentionally decoupled from fleet.Service so it can be used internally
// in crons and other services, leaving authentication/permission handling to
// the caller.
type MDMAppleCommander struct {
	storage nanomdm_storage.AllStorage
	pusher  nanomdm_push.Pusher
}

// NewMDMAppleCommander creates a new commander instance.
func NewMDMAppleCommander(mdmStorage nanomdm_storage.AllStorage, mdmPushService nanomdm_push.Pusher) *MDMAppleCommander {
	return &MDMAppleCommander{
		storage: mdmStorage,
		pusher:  mdmPushService,
	}
}

// InstallProfile sends the homonymous MDM command to the given hosts, it also
// takes care of the base64 encoding of the provided profile bytes.
func (svc *MDMAppleCommander) InstallProfile(ctx context.Context, hostUUIDs []string, profile mobileconfig.Mobileconfig, uuid string) error {
	base64Profile := base64.StdEncoding.EncodeToString(profile)
	raw := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CommandUUID</key>
	<string>%s</string>
	<key>Command</key>
	<dict>
		<key>RequestType</key>
		<string>InstallProfile</string>
		<key>Payload</key>
		<data>%s</data>
	</dict>
</dict>
</plist>`, uuid, base64Profile)
	err := svc.EnqueueCommand(ctx, hostUUIDs, raw)
	return ctxerr.Wrap(ctx, err, "commander install profile")
}

// InstallProfile sends the homonymous MDM command to the given hosts.
func (svc *MDMAppleCommander) RemoveProfile(ctx context.Context, hostUUIDs []string, profileIdentifier string, uuid string) error {
	raw := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CommandUUID</key>
	<string>%s</string>
	<key>Command</key>
	<dict>
		<key>RequestType</key>
		<string>RemoveProfile</string>
		<key>Identifier</key>
		<string>%s</string>
	</dict>
</dict>
</plist>`, uuid, profileIdentifier)
	err := svc.EnqueueCommand(ctx, hostUUIDs, raw)
	return ctxerr.Wrap(ctx, err, "commander remove profile")
}

func (svc *MDMAppleCommander) DeviceLock(ctx context.Context, hostUUIDs []string, uuid string) error {
	pin := GenerateRandomPin(6)
	raw := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>CommandUUID</key>
    <string>%s</string>
    <key>Command</key>
    <dict>
      <key>RequestType</key>
      <string>DeviceLock</string>
      <key>PIN</key>
      <string>%s</string>
    </dict>
  </dict>
</plist>`, uuid, pin)
	return svc.EnqueueCommand(ctx, hostUUIDs, raw)
}

func (svc *MDMAppleCommander) EraseDevice(ctx context.Context, hostUUIDs []string, uuid string) error {
	pin := GenerateRandomPin(6)
	raw := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>CommandUUID</key>
    <string>%s</string>
    <key>Command</key>
    <dict>
      <key>RequestType</key>
      <string>EraseDevice</string>
      <key>PIN</key>
      <string>%s</string>
    </dict>
  </dict>
</plist>`, uuid, pin)
	return svc.EnqueueCommand(ctx, hostUUIDs, raw)
}

func (svc *MDMAppleCommander) InstallEnterpriseApplication(ctx context.Context, hostUUIDs []string, uuid string, manifestURL string) error {
	raw := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>Command</key>
    <dict>
      <key>ManifestURL</key>
      <string>%s</string>
      <key>RequestType</key>
      <string>InstallEnterpriseApplication</string>
    </dict>

    <key>CommandUUID</key>
    <string>%s</string>
  </dict>
</plist>`, manifestURL, uuid)
	return svc.EnqueueCommand(ctx, hostUUIDs, raw)
}

type installEnterpriseApplicationPayload struct {
	Manifest    *appmanifest.Manifest
	RequestType string
}

func (svc *MDMAppleCommander) InstallEnterpriseApplicationWithEmbeddedManifest(
	ctx context.Context,
	hostUUIDs []string,
	uuid string,
	manifest *appmanifest.Manifest,
) error {
	cmd := commandPayload{
		CommandUUID: uuid,
		Command: installEnterpriseApplicationPayload{
			RequestType: "InstallEnterpriseApplication",
			Manifest:    manifest,
		},
	}

	raw, err := plist.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("marshal command payload plist: %w", err)
	}

	return svc.EnqueueCommand(ctx, hostUUIDs, string(raw))
}

// EnqueueCommand takes care of enqueuing the commands and sending push
// notifications to the devices.
//
// Always sending the push notification when a command is enqueued was decided
// internally, leaving making pushes optional as an optimization to be tackled
// later.
func (svc *MDMAppleCommander) EnqueueCommand(ctx context.Context, hostUUIDs []string, rawCommand string) error {
	cmd, err := mdm.DecodeCommand([]byte(rawCommand))
	if err != nil {
		return ctxerr.Wrap(ctx, err, "commander enqueue")
	}

	// MySQL implementation always returns nil for the first parameter
	_, err = svc.storage.EnqueueCommand(ctx, hostUUIDs, cmd)
	if err != nil {
		return ctxerr.Wrap(ctx, err, "commander enqueue")
	}

	apnsResponses, err := svc.pusher.Push(ctx, hostUUIDs)
	if err != nil {
		return ctxerr.Wrap(ctx, err, "commander push")
	}

	// Even if we didn't get an error, some of the APNs
	// responses might have failed, signal that to the caller.
	var failed []string
	for uuid, response := range apnsResponses {
		if response.Err != nil {
			failed = append(failed, uuid)
		}
	}
	if len(failed) > 0 {
		return &APNSDeliveryError{FailedUUIDs: failed, Err: err}
	}

	return nil
}

// APNSDeliveryError records an error and the associated host UUIDs in which it
// occurred.
type APNSDeliveryError struct {
	FailedUUIDs []string
	Err         error
}

func (e *APNSDeliveryError) Error() string {
	return fmt.Sprintf("APNS delivery failed with: %e, for UUIDs: %v", e.Err, e.FailedUUIDs)
}

func (e *APNSDeliveryError) Unwrap() error { return e.Err }

func (e *APNSDeliveryError) StatusCode() int { return http.StatusBadGateway }

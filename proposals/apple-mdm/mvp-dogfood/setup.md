# Setup

The setup consists of configuring:
- The APNs certificate used by the MDM protocol
- The SCEP certificate for enrollment
- The DEP token

We will define `fleetctl apple-mdm setup ...` commands to create/define all Apple/MDM credentials
that are fed to the Fleet server.

Certificates provided by Apple expire yearly, and new ones have to be downloaded from Apple and
uploaded into Fleet MDM.

## APNs

### Apple MDM APNs setup

Apple's MDM protocol uses the Apple Push Notification service (APNs) to deliver "wake up" messages to managed devices.
An "MDM server" needs access to an APNs certificate specifically issued for MDM management; such APNs certificate must be issued by an "MDM vendor."

Here's a sequence diagram with the three actors: Apple Inc., an MDM vendor, and a customer (MDM server).

```mermaid
%%{init: { 'theme':'dark', 'sequence': {'mirrorActors':false} } }%%

sequenceDiagram
    participant apple as Apple Inc.
    participant vendor as MDM Vendor
    participant server as MDM Server<br>(Customer)

    rect rgb(128, 128, 128)
    note left of apple: (1) MDM Vendor Setup at<br>https://business.apple.com
    note over vendor: Generate CSR
    vendor->>apple: Send CSR
    note over apple: Sign CSR
    apple->>vendor: "MDM vendor" certificate<br>(Setup)
    end
    rect rgb(255, 255, 255)
    note left of apple: (2) Customer Setup
    note over server: Generate CSR
    server->>+vendor: "Send" CSR (.csr)
    note over vendor: Sign CSR
    vendor->>server: "Send" signed CSR (XML plist, .req)
    rect rgb(128, 128, 128)
    note left of apple: https://identity.apple.com/pushcert
    server->>apple: Upload signed CSR (XML plist, .req)
    note over apple: Sign CSR
    apple->>server: Download APNS Certificate (PEM)
    end
    end
    note over server: APNs keypair<br>ready to use
```

The "MDM vendor setup" flow (1) is executed once by the "MDM vendor."

The "customer setup" flow (2) is executed by customers when they are setting up their MDM server.

The goal is for the Fleet organization to become an "MDM vendor" that issues CSRs to customers, which allows them to generate "APNs certificates" for their MDM deployments.

For the purposes of designing a PoC, we used the https://mdmcert.download/ service as an "MDM vendor."
See [MDMCert.Download Analysis](./mdmcert.download-analysis.md) for more details on the process.

### APNs setup with Fleet

The MDM APNs certificate provisioning will be manual on MVP:
- Customers will use `fleetctl` commands that will mimick `mdmctl mdmcert.download` commands (see [MDMCert.Download Analysis](mdmcert.download-analysis.md)).
- Fleet operators will perform the steps shown in the diagram above manually by running a new command line tool (under `tools/mdm-apple/mdm-apple-customer-setup`).

#### 1. Init APNs (customer)

`fleetctl apple-mdm setup apns init` 

The command will basically mimick [mdmctl mdmcert.download -new](https://github.com/micromdm/micromdm/blob/main/cmd/mdmctl/mdmcert.download.go).
Steps:
1. Generate an RSA private key and certificate for signing and encryption: `fleet-mdm-apple-apns-pki.crt` and `fleet-mdm-apple-apns-pki.key`
2. Generate an RSA push private key and CSR. Store the private key as a file: `fleet-mdm-apple-apns-push.key`. 
3. Also outputs:
- File fleet-mdm-apple-apns-setup.zip with:
	- `fleet-mdm-apple-apns-push.csr`
	- `fleet-mdm-apple-apns-pki.crt`
- Text to stdout that explains next step, something like:
	"Send zip to Fleet via preferred medium (e-mail, Slack)."

#### 2. New tool `tools/mdm-apple/mdm-apple-customer-setup` (Fleet representative)

`mdm-apple-customer-setup --zip fleet-mdm-apple-apns-setup.zip`

Output:
- `fleet-mdm-apple-apns-push-req-encrypted.p7`
- Text that explains next step, something like:
"Send generated file `fleet-mdm-apple-apns-push-req-encrypted.p7` back to customer via preferred medium (email, Slack)."

#### 3. Finalize APNs (customer)

```sh
fleetctl apple-mdm setup apns finalize \
    --certificate=fleet-mdm-apple-apns-pki.crt \
    --private-key=fleet-mdm-apple-apns-pki.key \
    --encrypted-req=fleet-mdm-apple-apns-push-req-encrypted.p7
```

Output:
- `fleet-mdm-apple-apns-push.req` file

#### 4. Upload .req to Apple (customer)

Customer uploads `fleet-mdm-apple-apns-push.req` to https://identity.apple.com.

#### 5. Download .crt from Apple (customer)

Downloads the final APNs certificate, a `*.crt` file. Let's call it `fleet-mdm-apple-apns-push.crt`.

The contents of `fleet-mdm-apple-apns-push.crt` and `fleet-mdm-apple-apns-push.key` are passed to Fleet as environment variables.

## SCEP

Apple's MDM protocol uses client certificates for client authentication. To generate client certificates, Apple's MDM protocol uses the [SCEP](https://en.wikipedia.org/wiki/Simple_Certificate_Enrollment_Protocol) protocol.

The setup for SCEP consists of generating the "SCEP CA" for Fleet.

### 1. Set up SCEP CA (customer)

`fleetctl apple-mdm setup scep`

Generates SCEP CA and key:
- `fleet-mdm-apple-scep.key`
- `fleet-mdm-apple-scep.crt`

The contents of `fleet-mdm-apple-scep.crt` and `fleet-mdm-apple-scep.key` are passed to Fleet as environment variables.

## DEP

### 1. Generate key material

`fleetctl apple-mdm setup dep init`

- Generates `fleet-mdm-apple-dep.crt` and `fleet-mdm-apple-dep.key` files.

### 2. Upload PEM to Apple

User uploads `fleet-mdm-apple-dep.crt` to https://business.apple.com, and downloads a `*.p7m` file.
Let's call it `fleet-mdm-apple-dep-auth-token-encrypted.p7m`.

### 3. Finalize DEP setup

```sh
fleetctl apple-mdm setup dep finalize \
    --certificate=fleet-mdm-apple-dep.crt \
    --private-key=fleet-mdm-apple-dep.key \
    --encrypted-token=fleet-mdm-apple-dep-auth-token-encrypted.p7m
```
	
1. Decrypts the provided token.
2. Generates a `fleet-mdm-apple-dep.token` file.

The contents of `fleet-mdm-apple-dep.token` is passed to Fleet as environment variable.
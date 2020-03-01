# Azure IoT Hub without SDK

This examples show how to connect to Azure IoT Hub without SDK.

## Examples

You can connect to Azure IoT Hub with:
- SAS token
- Self signed certificates
- CA Certificates

### Build

Build examples with:
```
$ ./build.sh
```

### Connect with SAS token

Create device identity twin with default SAS authentication:
```
$ az iot hub device-identity create -n {hub-name} -d {device_id}
```

Generate temporary SAS token:
```
$ az iot hub generate-sas-token -n {hub-name} -d {device-id}
```

Run example:
```
$ ./build/sas -hub-name {hub-name} -device-id {device-id} -sas-token "{sas-token}"
```

### Connect with self signed certificates

Create device identity twin with self signed certificate:
```
$ az iot hub device-identity create -n {hub-name} -d {device-id} --am x509_thumbprint --output-dir sa-cert/
```

Make sure generated certificates for the device are in sa-cert folder or change path in the code.

Run example:
```
$ ./build/sac -hub-name {hub-name} -device-id {device-id} -cert-file {device-cert.pem} -key-file {device-key.pem}
```

### Connect with CA Certificates

Generate your root certificate:
```
$ openssl genrsa -out root-ca.key 2048
$ openssl req -x509 -new -nodes -key root-ca.key -sha256 -days 1024 -out root-ca.pem
```

Upload `root-ca.pem` to Azure IoT Hub and get verification code.

Generate verification certificate, **make sure to put [verification code] to Common name field**
```
$ openssl genrsa -out verification-cert.key 2048
$ openssl req -new -key verification-cert.key -out verification-cert.csr
$ openssl x509 -req -in verification-cert.csr -CA root-ca.pem -CAkey root-ca.key -CAcreateserial -out verification-cert.pem -days 31 -sha256
```

Upload `verification-cert.pem`

Generate device certificate, **make sure to put [device id] to Common name field**
```
$ openssl genrsa -out device-cert.key 2048
$ openssl req -new -key device-cert.key -out device-cert.csr
$ openssl x509 -req -in device-cert.csr -CA root-ca.pem -CAkey root-ca.key -CAcreateserial -out device-cert.pem -days 31 -sha256
```

Create device identity twin with CA certs in Azure IoT Hub: 
```
$ az iot hub device-identity create -n {hub-name} -d {device-id} --am x509_ca
```

Run ca-certificate example:
```
$ ./build/cac -hub-name {hub-name} -device-id {device-id} -cert-file {device-cert.pem} -key-file {device-key.pem}
```

### Useful Azure CLI tool commands for this examples

Reading Azure IoT Hub with cli tool:
```
$ az iot hub monitor-events --hub-name {hub-name} --device-id {device-id}
``` 

Generating SAS Token with cli tool:
```
$ az iot hub generate-sas-token -n {hub-name} -d {device-id} 
```

Create device identity with self signed certificate:
```
$ az iot hub device-identity create -n {hub-name} -d {device-id} --am x509_thumbprint --output-dir /path/to/output
```

Create device identity with CA certs in Azure IoT Hub: 
```
$ az iot hub device-identity create -n {hub-name} -d {device-id} --am x509_ca
```

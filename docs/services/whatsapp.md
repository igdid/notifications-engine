# WhatsApp

This service is added here for use mainly in Argo Events but not in Argo CD.
Maybe in the future its usage will be expanded.

## Parameters

* `apiURL` - the server url, e.g. https://graph.facebook.com
* `apiVersion` - the API Version, e.g. v23.0
* `phoneNumberID` - the Phone Number ID. It can be found in the WhatsApp Dev Console
* `token` - the bot token
* `insecureSkipVerify` - optional bool, true or false
* `maxIdleConns` - optional, maximum number of idle (keep-alive) connections across all hosts.
* `maxIdleConnsPerHost` - optional, maximum number of idle (keep-alive) connections per host.
* `maxConnsPerHost` - optional, maximum total connections per host.
* `idleConnTimeout` - optional, maximum amount of time an idle (keep-alive) connection will remain open before closing, e.g. '90s'.

## Configuration

1. Create an app in WhatsApp Dev Console and copy token after creating it
2. Store token in `argocd-notifications-secret` Secret and configure WhatsApp integration
in `argocd-notifications-cm` ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-notifications-cm
data:
  service.whatsapp: |
    apiURL: <api-url>
    apiVersion: v23.0
    token: $whatsapp-token
    phoneNumberID: <phone-number-id>
```

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: <secret-name>
stringData:
  whatsapp-token: token
```

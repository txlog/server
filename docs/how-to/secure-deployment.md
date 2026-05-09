# How to Configure a Secure Deployment

When deploying the Txlog server in a production environment, it is highly
recommended to place it behind a reverse proxy (like HAProxy, Nginx, or Traefik)
that handles TLS termination.

The server enforces several security hardening measures that rely on this
deployment topology.

## Enforcing Secure Cookies

By default, in development mode, the server uses `SameSite=Lax` for session
cookies to mitigate CSRF attacks but does not enforce the `Secure` flag,
allowing you to test locally over HTTP.

In production, you **must** instruct the server to enforce the `Secure` flag on
all session cookies. This guarantees that cookies are never transmitted over
unencrypted HTTP connections.

To do this, ensure the server is running in release mode by setting the
following environment variable:

```bash
GIN_MODE=release
```

When `GIN_MODE` is set to `release`, the server will automatically mark all
authentication and state cookies as `Secure`.

## Reverse Proxy Configuration

If your Txlog server is running on `http://127.0.0.1:8080` internally, but your
users access it via `https://txlog.company.com`, the reverse proxy is performing
TLS termination.

Because the server is now enforcing `Secure` cookies (due to
`GIN_MODE=release`), the client's browser will reject these cookies if it thinks
the connection is insecure.

You must ensure your reverse proxy correctly forwards the protocol information
to the Txlog server using the `X-Forwarded-Proto` header.

### HAProxy Example

```haproxy
frontend https_front
    bind *:443 ssl crt /etc/ssl/certs/txlog.pem
    http-request set-header X-Forwarded-Proto https
    default_backend txlog_backend

backend txlog_backend
    server txlog1 127.0.0.1:8080 check
```

### Nginx Example

```nginx
server {
    listen 443 ssl;
    server_name txlog.company.com;

    ssl_certificate /etc/ssl/certs/txlog.crt;
    ssl_certificate_key /etc/ssl/private/txlog.key;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto https;
    }
}
```

## Internal TLS Verification

The server communicates with external Identity Providers (OIDC or LDAP) to
authenticate users. To prevent Man-In-The-Middle (MITM) attacks, the Txlog
server **strictly enforces TLS certificate verification** for all outbound
authentication connections.

- There are no "skip verify" bypass flags available in the configuration.
- If your internal LDAP or OIDC server uses a self-signed certificate or a
  private Certificate Authority (CA), you must add that CA to the trusted
  certificate store of the operating system or container running the Txlog
  server.

If the CA is not trusted by the host running Txlog, authentication will fail
with a certificate validation error.

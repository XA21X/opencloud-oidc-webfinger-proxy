# opencloud-oidc-webfinger-proxy

A lightweight **Go-based reverse proxy** that dynamically rewrites `href` fields in **WebFinger responses** for OpenID Connect (OIDC) clients.

This service is primarily designed for **OpenCloud** environments but can be used with any OIDC implementation that requires dynamic issuer rewriting based on user agent or suffix mapping.

---

## 🚀 Features

- 🔄 Dynamically rewrites `href` values in WebFinger JSON responses  
- ⚙️ Fully configurable via environment variables — no code changes required  
- 🧩 Compatible with any OIDC-compliant identity provider (e.g. Authentik, Authelia, Dex, etc.)  
- 🧠 Extremely lightweight and stateless — built with Go  
- 🧱 Works seamlessly behind Nginx or any reverse proxy

## ✅ Tested / Compatible Identity Providers

The proxy has been verified to work with the following OIDC implementations:

| Identity Provider | Compatibility | Notes |
|--------------------|---------------|--------|
| **[Authentik](https://goauthentik.io/)** | ✅ Tested | Fully compatible — WebFinger rewrites confirmed functional |

If you’ve tested it with other identity providers, please open a PR to extend the list.

---

## 🧠 How it works

1. The proxy listens for requests on `/.well-known/webfinger`.
2. It forwards the request to the configured **upstream WebFinger service**.
3. The JSON response is inspected and any `href` fields matching `HREF_PATTERN` are replaced with  
   `HREF_REPLACEMENT + <issuer-suffix> + "/"`.
4. The **issuer suffix** is provided via the HTTP header `X-Issuer-Suffix`,  
   or falls back to the configured `DEFAULT_SUFFIX`.

Example:

# Upstream response
```json
{
  "href": "https://auth.domain.com/application/o/opencloud/"
}
```

# Modified by proxy (Header: X-Issuer-Suffix=opencloud-android)
```json
{
  "href": "https://auth.domain.com/application/o/opencloud-android/"
}
```

## ⚙️ Environment Variables

| Variable | Description | Example |
|-----------|-------------|----------|
| `PORT` | Port on which the proxy listens | `9210` |
| `UPSTREAM_URL` | Base URL of the real WebFinger endpoint (including schema) | `https://cloud.domain.com` |
| `HREF_PATTERN` | The part of the `href` URL to be matched and replaced | `https://auth.domain.com/application/o/` |
| `HREF_REPLACEMENT` | The base part to replace the matched pattern with | `https://auth.domain.com/application/o/` |
| `DEFAULT_SUFFIX` | Default issuer suffix when none is provided via header | `opencloud` |

All variables are **required** — the service will exit with an error if any are missing.

## 🧩 Example usage

### Run locally

```bash
export PORT=9210
export UPSTREAM_URL=https://cloud.domain.com
export HREF_PATTERN=https://auth.domain.com/application/o/
export HREF_REPLACEMENT=https://auth.domain.com/application/o/
export DEFAULT_SUFFIX=opencloud

go run main.go
```

Then test it:

```bash
curl -H "X-Issuer-Suffix: opencloud-android" "http://localhost:9210/.well-known/webfinger?rel=http%3A%2F%2Fopenid.net%2Fspecs%2Fconnect%2F1.0%2Fissuer&resource=https%3A%2F%2Fcloud.domain.com"
```

Response:
```json
{
  "links": [
    {
      "rel": "http://openid.net/specs/connect/1.0/issuer",
      "href": "https://auth.domain.com/application/o/opencloud-android/"
    }
  ],
  "subject": "https://cloud.domain.com"
}
```

## 🐳 Run with Docker

```bash
docker build -t opencloud-oidc-webfinger-proxy .

docker run -p 9210:9210 \
-e UPSTREAM_URL=https://cloud.domain.com \
-e HREF_PATTERN=https://auth.domain.com/application/o/ \
-e HREF_REPLACEMENT=https://auth.domain.com/application/o/ \
-e DEFAULT_SUFFIX=opencloud \
ghcr.io/2bros-group/opencloud-oidc-webfinger-proxy:latest
```

## 🧱 Reverse Proxy Nginx Integration

```
map $http_user_agent $issuer_suffix {
    "~*mirall.*OpenCloud"  "opencloud-desktop";
    "~*OpenCloudApp"       "opencloud-ios";
    "~*OpenCloud-android"  "opencloud-android";
    default                "opencloud";
}

location = /.well-known/webfinger {
    proxy_set_header X-Issuer-Suffix $issuer_suffix;
    proxy_pass http://127.0.0.1:9210;
}
```

This allows your Nginx instance to determine the correct `issuer_suffix` per client  
and let the proxy dynamically rewrite the WebFinger response accordingly.

## 🤝 Contributing

Pull requests are welcome!  
If you plan a major change (e.g. multiple pattern replacements or caching),  
please open an issue first to discuss what you’d like to add.

## 📄 License

MIT License © 2025 2Bros Digital Group GmbH
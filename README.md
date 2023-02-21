# frigabun

Web service to allow FritzBox routers to update Gandi and Porkbun DNS entries when obtaining a new IP address.
Uses the new LiveDNS API. Written in Go 1.20

## Requirements
- A domain name on Gandi or Porkbun
- Gandi or Porkbun API credentials from [Account settings](https://account.gandi.net/) under Security
- FritzBox router with up-to-date firmware
- Optional: To build or run manually: Go 1.20

## Set up service

- Download the [latest](https://github.com/davidramiro/frigabun/releases/latest) release archive for your OS/arch
- Unzip, rename `config.sample.yml` to `config.yml` (config is fine as default, if you want to run tests, fill in your API info)

## FritzBox settings

- Log into your FritzBox
- Navigate to `Internet` -> `Permit Access` -> `DynDNS`
- Enable DynDNS and use `User-defined` as Provider

### Gandi

- Enter the following URL: `http://{HOST}:{PORT}/api/update?apikey=<passwd>&domain={DOMAIN}&subdomain={SUBDOMAIN}&ip=<ipaddr>`
  - Replace the `{HOST}` and `{PORT}` with your deployment of the application
    - By default, the application uses port `9595`
  - Replace `{DOMAIN}` with your base domain
    - e.g. `yourdomain.com`
  - Replace `{SUBDOMAIN}` with your subdomain or comma separated subdomains
    - e.g. `subdomain` or `sudomain1,subdomain2`
- Enter the full domain in the `Domain Name` field
  - e.g. `subdomain.domain.com` (if you use multiple subdomains, just choose any of those)
- Enter any value in the `Username` field
  - Unused, but required by the FritzBox interface
- Enter your Gandi API-Key in the `Password` field

### Porkbun

- Enter the following URL: `http://{HOST}:{PORT}/api/update?apikey=<username>&secretapikey=<passwd>&domain={DOMAIN}&subdomain={SUBDOMAIN}&ip=<ipaddr>&registrar=porkbun`
  - Replace the `{HOST}` and `{PORT}` with your deployment of the application
    - By default, the application uses port `9595`
  - Replace `{DOMAIN}` with your base domain
    - e.g. `yourdomain.com`
  - Replace `{SUBDOMAIN}` with your subdomain or comma separated subdomains
    - e.g. `subdomain` or `sudomain1,subdomain2`
- Enter the full domain in the `Domain Name` field
  - e.g. `subdomain.domain.com` (if you use multiple subdomains, just choose any of those)
- Enter your Porkbun API key in the `Username` field
- Enter your Porkbun API Secret Key in the `Password` field


Your settings should look something like this:

![](https://kore.cc/fritzgandi/fbsettings.png "FritzBox DynDNS Settings")

Right after you save the settings, your FritzBox will make a request to the application. You should see the following
success message in its log:

![](https://kore.cc/fritzgandi/success.png "Success Message")

Your FritzBox will now automatically communicate new IPs to the application. 

## Security notice
If you deploy this application outside your local network, I'd recommend you to use HTTPS for the requests.
Check below for an example on how to reverse proxy to this application with NGINX. 

## Linux systemd Service

To create a systemd service and run the application on boot, create a service file, for example under
`/etc/systemd/system/frigabun.service`.

Service file contents: 
```
[Unit]
Description=FritzGandi LiveDNS Microservice

[Service]
WorkingDirectory=/path/to/frigabun
ExecStart=/path/to/frigabun/executable
User=youruser
Type=simple
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Don't run this as root. Make sure your `User` has access rights to the `WorkingDirectory` where the executable is in.

Reload daemon, start the service, check its status:

```
sudo systemctl daemon-reload
sudo systemctl start frigabun.service
sudo systemctl status frigabun
```

If all is well, enable the service to be started on boot:

`sudo systemctl enable frigabun`

## NGINX Reverse Proxy

If you want to host the service and make sure it uses HTTPS, you can use a reverse proxy.
Shown below is an example of an NGINX + LetsEncrypt reverse proxy config for this microservice.

```
server {
    listen                  443 ssl http2;
    listen                  [::]:443 ssl http2;
    server_name             frigabun.yourdomain.com;

    # SSL
    ssl_certificate         /etc/letsencrypt/live/frigabun.yourdomain.com/fullchain.pem;
    ssl_certificate_key     /etc/letsencrypt/live/frigabun.yourdomain.com/privkey.pem;
    ssl_trusted_certificate /etc/letsencrypt/live/frigabun.yourdomain.com/chain.pem;

    # security headers
    add_header X-Frame-Options           "DENY";
    add_header X-XSS-Protection          "1; mode=block" always;
    add_header X-Content-Type-Options    "nosniff" always;
    add_header Referrer-Policy           "no-referrer-when-downgrade" always;
    add_header Content-Security-Policy   "default-src 'self' http: https: data: blob: 'unsafe-inline'" always;
    add_header Strict-Transport-Security 'max-age=31536000; includeSubDomains; preload';
    add_header X-Permitted-Cross-Domain-Policies master-only;
    
    
    # . files
    location ~ /\.(?!well-known) {
        deny all;
    }

    # logging
    access_log              /var/log/nginx/frigabun.yourdomain.com.access.log;
    error_log               /var/log/nginx/frigabun.yourdomain.com.error.log warn;

    # reverse proxy
    location / {
        proxy_pass http://127.0.0.1:9595;
        proxy_http_version                 1.1;
        proxy_cache_bypass                 $http_upgrade;
        
        # Proxy headers
        proxy_set_header Upgrade           $http_upgrade;
        proxy_set_header Connection        "upgrade";
        proxy_set_header Host              $host;
        proxy_set_header X-Real-IP         $remote_addr;
        proxy_set_header X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Forwarded-Host  $host;
        proxy_set_header X-Forwarded-Port  $server_port;
        
        # Proxy timeouts
        proxy_connect_timeout              60s;
        proxy_send_timeout                 60s;
        proxy_read_timeout                 60s;
    }
}

# HTTP redirect
server {
    listen      80;
    listen      [::]:80;
    server_name frigabun.yourdomain.com;
    
    # ACME-challenge
    location ^~ /.well-known/acme-challenge/ {
        root /var/www/_letsencrypt;
    }

    location / {
        return 301 https://frigabun.yourdomain.com$request_uri;
    }
}
```
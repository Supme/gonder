#### To generate a certificate

    openssl req -x509 -sha256 -nodes -days 3650 -newkey rsa:4096 -keyout server.key -out server.pem


Or use https://letsencrypt.org/
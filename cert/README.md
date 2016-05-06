#### Generated private key

    openssl genrsa -out server.key 2048

#### To generate a certificate

    openssl req -new -x509 -key server.key -out server.pem -days 3650


Or use https://letsencrypt.org/
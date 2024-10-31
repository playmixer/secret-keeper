
mkdir ./cert
openssl genrsa -out ./cert/gophkeeper.key 2048
openssl ecparam -genkey -name secp384r1 -out ./cert/gophkeeper.key
openssl req -new -x509 -sha256 -key ./cert/gophkeeper.key -out ./cert/gophkeeper.crt -days 3650
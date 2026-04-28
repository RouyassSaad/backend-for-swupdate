#!/bin/bash



FILES="sw-description "

if [[ ! -z "$1" ]]
then
	FILES="$FILES sw-description.sig"
	openssl cms -sign -in  sw-description -out sw-description.sig -signer $1 \
		-inkey $2 -outform DER -nosmimecap -binary
fi

FILES="$FILES main.go"

for i in $FILES;do
        echo $i;done | cpio -ov -H crc --reproducible >  out.swu

#!/bin/bash
TACS_SCHEME=scheme.yaml \
TACS_ADDR= \
TACS_PORT=8080 \
TACS_CERT=ssl.crt \
TACS_CERT_KEY=ssl.key \
TACS_LDAP_SERVER=ldap.example.com \
TACS_LDAP_PORT=389 \
TACS_LDAP_CERT=ldap.crt \
TACS_LDAP_KEY=ldap.key \
TACS_LDAP_USER=read_ldap@example.com \
TACS_LDAP_PASSWORD=SeCrEtPasSWorD1! \
./tacs
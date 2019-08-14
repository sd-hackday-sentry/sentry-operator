#!/bin/bash

NS="$1"

if kubectl --namespace=${NS} get secret/sentry; then
  read -p "The sentry secret already exists, you want to replace? (y/N) " choice
  case "$choice" in
    y|Y )
      echo "Re-creating sentry secret...";;
    * )
      echo "Leaving sentry secret alone."
      exit 0
      ;;
  esac
fi

read -s -p "Sentry database password: " sentry_db_pass
echo
read -p "Sentry superuser account email: " sentry_su_email
read -s -p "Sentry superuser account password: " sentry_su_password
echo

echo "Generating a random sentry secret key..."
hashed_secret=$(dd if=/dev/urandom bs=1024 count=8 | md5sum | cut -c1-32)

cat << EOF | kubectl --namespace=${NS} apply -f-
apiVersion: v1
stringData:
  SENTRY_SECRET_KEY: ${hashed_secret}
  SENTRY_DB_PASSWORD: ${sentry_db_pass}
  SENTRY_SU_EMAIL: ${sentry_su_email}
  SENTRY_SU_PASSWORD: ${sentry_su_password}
kind: Secret
metadata:
  name: sentry
  labels:
    app: sentry
EOF

echo "Done. If you run 'make install' the underlying services will use the same credentials you just created."

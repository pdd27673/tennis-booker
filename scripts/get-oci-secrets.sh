#!/bin/bash

echo "🔐 OCI Secrets Helper Script"
echo "============================="
echo
echo "Run these commands to get your OCI secrets:"
echo
echo "1. OCI_PRIVATE_KEY (run this and copy the entire output):"
echo "   cat ~/.oci/oci_api_key.pem"
echo
echo "2. To find your other OCI values, check:"
echo "   cat ~/.oci/config"
echo "   OR check your infrastructure/terraform/terraform.tfvars file"
echo
echo "3. Your secrets should look like this:"
echo "   OCI_USER_OCID: ocid1.user.oc1..aaaaaaaa..."
echo "   OCI_TENANCY_OCID: ocid1.tenancy.oc1..aaaaaaaa..."
echo "   OCI_FINGERPRINT: aa:bb:cc:dd:ee:ff:gg:hh:ii:jj:kk:ll:mm:nn:oo:pp"
echo "   OCI_REGION: us-ashburn-1 (or your region)"
echo "   OCI_PRIVATE_KEY: -----BEGIN RSA PRIVATE KEY-----"
echo "                    (entire private key content)"
echo "                    -----END RSA PRIVATE KEY-----"
echo

# Helper to show current values if they exist
echo "🔍 Current OCI Configuration (if available):"
echo "============================================="

if [ -f ~/.oci/config ]; then
    echo "From ~/.oci/config:"
    cat ~/.oci/config
else
    echo "No ~/.oci/config file found"
fi

if [ -f infrastructure/terraform/terraform.tfvars ]; then
    echo
    echo "From terraform.tfvars:"
    grep -E "(user_ocid|tenancy_ocid|fingerprint|region)" infrastructure/terraform/terraform.tfvars 2>/dev/null || echo "No terraform.tfvars found"
fi

echo
echo "📋 Next Steps:"
echo "1. Copy these values to GitHub Secrets (see instructions below)"
echo "2. For OCI_PRIVATE_KEY, copy the ENTIRE content including BEGIN/END lines"
echo "3. Make sure there are no extra spaces or newlines when pasting" 
# 🔐 GitHub Actions Secrets Setup Guide

This guide will help you set up all the required secrets for your GitHub Actions CI/CD workflows.

## 📋 Required Secrets Overview

Your workflows need these secrets:

### 🏗️ **OCI Deployment Secrets** (for CD workflow)
- `OCI_USER_OCID`
- `OCI_TENANCY_OCID` 
- `OCI_FINGERPRINT`
- `OCI_PRIVATE_KEY`
- `OCI_REGION`

### 🐳 **Docker Registry Secrets** (auto-provided)
- `GITHUB_TOKEN` ✅ (automatically available)

---

## 🔑 Step 1: Get Your OCI Secrets

### Option A: If you have OCI already configured

1. **Run the helper script:**
   ```bash
   chmod +x get-oci-secrets.sh
   ./get-oci-secrets.sh
   ```

2. **Or manually check your files:**
   ```bash
   # Check your OCI config
   cat ~/.oci/config
   
   # Get your private key (copy the entire output)
   cat ~/.oci/oci_api_key.pem
   
   # Check your terraform vars if you have them
   cat infrastructure/terraform/terraform.tfvars
   ```

### Option B: If you need to set up OCI from scratch

1. **Create OCI Account:**
   - Go to https://cloud.oracle.com/
   - Sign up for Oracle Cloud (Free Tier available)

2. **Generate API Key:**
   ```bash
   # Create directory
   mkdir -p ~/.oci
   
   # Generate private key
   openssl genrsa -out ~/.oci/oci_api_key.pem 2048
   
   # Generate public key
   openssl rsa -pubout -in ~/.oci/oci_api_key.pem -out ~/.oci/oci_api_key_public.pem
   
   # Secure the private key
   chmod 600 ~/.oci/oci_api_key.pem
   ```

3. **Add Public Key to OCI:**
   - Login to OCI Console
   - Click your profile → User Settings
   - Click "API Keys" → "Add API Key"
   - Select "Paste Public Key"
   - Copy content from `~/.oci/oci_api_key_public.pem`
   - Paste and click "Add"

4. **Copy Configuration Values:**
   After adding the key, OCI shows you the configuration. Copy these values:
   - User OCID: `ocid1.user.oc1..aaaaaaaa...`
   - Tenancy OCID: `ocid1.tenancy.oc1..aaaaaaaa...` 
   - Fingerprint: `aa:bb:cc:dd:ee:ff...`
   - Region: `us-ashburn-1` (or your chosen region)

---

## 🔧 Step 2: Add Secrets to GitHub

1. **Go to your GitHub repository**
2. **Click Settings → Secrets and variables → Actions**
3. **Click "New repository secret"**
4. **Add each secret:**

### Secret 1: OCI_USER_OCID
- **Name:** `OCI_USER_OCID`
- **Value:** `ocid1.user.oc1..aaaaaaaa...` (your user OCID)

### Secret 2: OCI_TENANCY_OCID  
- **Name:** `OCI_TENANCY_OCID`
- **Value:** `ocid1.tenancy.oc1..aaaaaaaa...` (your tenancy OCID)

### Secret 3: OCI_FINGERPRINT
- **Name:** `OCI_FINGERPRINT`  
- **Value:** `aa:bb:cc:dd:ee:ff:gg:hh:ii:jj:kk:ll:mm:nn:oo:pp` (your key fingerprint)

### Secret 4: OCI_REGION
- **Name:** `OCI_REGION`
- **Value:** `us-ashburn-1` (or your OCI region)

### Secret 5: OCI_PRIVATE_KEY ⚠️ **IMPORTANT**
- **Name:** `OCI_PRIVATE_KEY`
- **Value:** The ENTIRE content of your `~/.oci/oci_api_key.pem` file including:
  ```
  -----BEGIN RSA PRIVATE KEY-----
  MIIEowIBAAKCAQEA...
  (multiple lines of key content)
  ...
  -----END RSA PRIVATE KEY-----
  ```
- **⚠️ Critical:** Copy the entire file content, including the BEGIN/END lines, with no extra spaces

---

## 🔍 Step 3: Verify Secrets

After adding all secrets, your GitHub Secrets page should show:
- ✅ `OCI_USER_OCID`
- ✅ `OCI_TENANCY_OCID`  
- ✅ `OCI_FINGERPRINT`
- ✅ `OCI_PRIVATE_KEY`
- ✅ `OCI_REGION`

---

## 🧪 Step 4: Test Your Setup

1. **Create a test tag to trigger deployment:**
   ```bash
   git tag v0.1.0-test
   git push origin v0.1.0-test
   ```

2. **Check GitHub Actions:**
   - Go to your repo → Actions tab
   - Watch the CD workflow run
   - If secrets are correct, the OCI setup step should pass

3. **If deployment fails:**
   - Check the workflow logs for specific error messages
   - Common issues:
     - Malformed private key (extra spaces/newlines)
     - Wrong region or OCID format
     - OCI permissions issues

---

## 🚨 Troubleshooting Common Issues

### "Invalid private key format"
- Make sure you copied the ENTIRE private key file content
- No extra spaces before `-----BEGIN` or after `-----END`
- All lines should be included

### "User not found" or "Tenancy not found"  
- Double-check your OCID values
- Make sure they start with `ocid1.user.oc1..` or `ocid1.tenancy.oc1..`

### "Authentication failed"
- Verify your fingerprint matches exactly
- Make sure the API key is added to the correct user in OCI

### "Access denied" or "Authorization failed"
- Check that your OCI user has the necessary permissions
- You might need policies for compute, networking, etc.

---

## 🔒 Security Best Practices

1. **Never commit secrets to your repository**
2. **Use GitHub Secrets for all sensitive data**
3. **Regularly rotate your API keys**
4. **Use least-privilege access in OCI**
5. **Monitor your GitHub Actions logs for exposed secrets**

---

## 📞 Need Help?

If you're still having issues:

1. **Check the helper script:** `./get-oci-secrets.sh`
2. **Review OCI documentation:** https://docs.oracle.com/en-us/iaas/Content/API/Concepts/apisigningkey.htm
3. **Check GitHub Actions logs** for specific error messages
4. **Verify your OCI setup** with the OCI CLI locally first

---

**🎾 Ready to deploy!** Once all secrets are set up, your GitHub Actions will automatically deploy to OCI when you push tags (like `v1.0.0`). 
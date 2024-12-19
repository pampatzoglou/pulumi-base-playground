# pulumi-base-playground

# Architecture

# Requirements

## Install pulumi

```
curl -fsSL https://get.pulumi.com | sh
```

## Install go

Install Kind (dev)

```
# For AMD64 / x86_64
[ $(uname -m) = x86_64 ] && curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.26.0/kind-linux-amd64
# For ARM64
[ $(uname -m) = aarch64 ] && curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.26.0/kind-linux-arm64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind
```

## Setup your host for Pulumi to access your cloud account

```
export AWS_ACCESS_KEY_ID="<YOUR_ACCESS_KEY_ID>"
export AWS_SECRET_ACCESS_KEY="<YOUR_SECRET_ACCESS_KEY>"
```



# Run

### Local:

```
pulumi config set cluster:type local
pulumi config set cluster:name my-cluster
pulumi up

pulumi destroy
```

### Dev

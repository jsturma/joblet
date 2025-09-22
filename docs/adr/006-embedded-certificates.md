# ADR-006: Embedded Certificate Architecture

## Status

Accepted

## Context

TLS certificates are a pain. There, I said it. Every ops person knows the dance: generate certificates, distribute them
to the right places, update configuration to point to them, make sure permissions are correct, rotate them before they
expire, and pray nothing breaks in the process.

When we were designing Joblet's security model, we knew we needed mutual TLS authentication. The server needs to verify
clients, clients need to verify the server. But we also knew that traditional certificate management would be a
deployment nightmare.

Think about it: you're deploying Joblet to a server, and now you need to copy certificates, put them in the right
directory, set the right permissions, update the config with the paths... and then do it all again for each client. And
when those certificates expire? Have fun updating everything.

We needed something better. Something that wouldn't make ops teams hate us.

## Decision

We embedded the certificates directly in the YAML configuration files. No separate certificate files. No path
management. No permission juggling. Just configuration.

```yaml
# Everything in one place
security:
  serverCert: |
    -----BEGIN CERTIFICATE-----
    MIIDXTCCAkWgAwIBAgIJAKoK/heBjcO...
    -----END CERTIFICATE-----
  serverKey: |
    -----BEGIN PRIVATE KEY-----
    MIIEvgIBADANBgkqhkiG9w0BAQEFAA...
    -----END PRIVATE KEY-----
```

When you deploy Joblet, you copy one config file. When you set up a client, you give them one config file. When you
rotate certificates, you update one config file. It's so simple it feels like cheating.

## Consequences

### The Good

Deployment became ridiculously simple. We went from a multi-step process with various failure modes to "copy this file
to /opt/joblet/config/". That's it. No Chef recipes or Ansible playbooks needed just to manage certificates.

Configuration is atomic. You can't have a situation where the config points to a certificate that doesn't exist, or
where you updated the certificate but forgot to update the config. Everything updates together.

Distribution is straightforward. Need to give someone access? Send them a config file. It has everything they need. No "
also copy these three certificate files and put them here" instructions.

Version control works beautifully. The entire configuration, including security credentials, can be versioned as a
single unit. Rolling back is trivial.

### The Trade-offs

Yes, the configuration files are sensitive now. They contain private keys. But honestly, they were always sensitive -
they had certificate paths and connection details. Now we're just being honest about it.

The config files are larger. Instead of a 20-line config with paths, you have a 200-line config with embedded
certificates. But who cares? It's still tiny by modern standards.

You can't share certificates between multiple services as easily. Each service gets its own embedded certificates. But
in practice, this improves security - compromise of one service doesn't affect others.

### The Security Considerations

We didn't make this decision lightly. Embedding certificates could be seen as less secure than separate files with
restrictive permissions. But we realized that in practice, it's often more secure.

Why? Because complexity is the enemy of security. The traditional approach leads to mistakes. Wrong permissions,
certificates in the wrong place, expired certificates because rotation is too hard. Our approach makes the right thing
the easy thing.

We still protect the config files with proper permissions (600 for server, 600 for client). But there's only one thing
to protect, not multiple files scattered around the filesystem.

### The Unexpected Benefits

The embedded approach enabled some nice features. Config validation can verify certificates are valid and not expired.
We can check certificate-config matching (does this private key match this certificate?). We can even auto-generate
configurations with new certificates as part of the provisioning process.

It also made testing easier. Test configurations are self-contained. No need to generate test certificates and manage
temporary files. Just embed test certificates in test configs.

The approach also simplified containerization. The container just needs one config file mounted. No certificate volume
mounts, no init containers to copy certificates, no special entrypoint scripts.

### Real-World Example

Here's what onboarding a new team member looks like:

**Traditional approach:**

1. Generate client certificate
2. Sign with CA
3. Copy certificate to their machine
4. Copy private key to their machine
5. Copy CA certificate to their machine
6. Create config file with correct paths
7. Set permissions on all files
8. Debug when it doesn't work because step 5 was missed

**Our approach:**

1. Generate config file with embedded certificates
2. Send secure link
3. They save it to `~/.joblet/config.yml`
4. It works

The simplicity is transformative. Certificate management went from a dreaded task to a non-issue.

## Learn More

See [CONFIGURATION.md](/docs/CONFIGURATION.md) for configuration details
and [DESIGN.md](/docs/DESIGN.md#71-embedded-certificate-architecture) for the complete security model.
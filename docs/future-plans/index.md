# Future Plans

This document outlines the roadmap and future direction for OSPA.

## Vision

OSPA aims to be the comprehensive, extensible policy engine for OpenStack cloud governance, enabling organizations to:

- Enforce security best practices across their OpenStack infrastructure
- Optimize cloud costs through automated resource cleanup
- Maintain compliance with organizational and regulatory requirements
- Gain visibility into cloud resource usage and hygiene

## Roadmap

#### Complete Service Coverage

Expand support to all major OpenStack services:

| Service | Status |
|---------|--------|
| Nova (Compute) | Partial |
| Neutron (Network) | **Implemented** |
| Cinder (Block Storage) | Partial |
| Glance (Image) | Planned |
| Keystone (Identity) | Planned |
| Heat (Orchestration) | Planned |
| Swift (Object Storage) | Planned |

#### Enhanced Policy Features

- **Conditional actions** - Apply different actions based on resource attributes
- **Severity levels** - Classify violations by severity
- **Policy inheritance** - Base policies that can be extended
- **Cross-resource checks** - Policies that span multiple resource types

#### Improved User Experience

- **Web dashboard** - Visual interface for viewing findings and writing policies
- **Better error messages** - More helpful validation errors
- **Policy linting** - Catch common mistakes before runtime


#### Advanced Remediation

- **Approval workflows** - Require approval before remediation
- **Staged rollout** - Apply remediations gradually
- **Rollback support** - Undo remediations if needed
- **Remediation scheduling** - Schedule remediations for maintenance windows

#### Integration & Ecosystem

- **Webhook notifications** - Send findings to external systems
- **SIEM integration** - Export to security information systems
- **Cloud management platforms** - Integration with CMPs
- **Terraform provider** - Manage OSPA via Terraform

### (if we decide to go crazy)
#### Multi-Cloud Support 

- **AWS** - Extend the policy framework to AWS resources
- **Azure** - Extend to Azure resources
- **GCP** - Extend to Google Cloud resources
- **Unified policies** - Write once, apply across clouds


#### Machine Learning

- **Anomaly detection** - Identify unusual resource patterns
- **Prediction** - Predict compliance issues before they occur
- **Auto-tuning** - Automatically adjust policy thresholds
- **Resource recommendations** - Suggest policy improvements

#### Enterprise Features

- **Multi-tenancy** - Manage policies across organizations
- **RBAC** - Role-based access control for policy management
- **Audit logging** - Detailed audit trail of all actions
- **Compliance reporting** - Generate compliance reports

## Feature Requests

We welcome feature requests! To suggest a feature:

1. Check existing [GitHub Issues](https://github.com/OpenStack-Policy-Agent/OSPA/issues)
2. Open a new issue with the "enhancement" label
3. Describe the use case and proposed solution

## Current Focus Areas

### Immediate Priorities

1. **Complete Nova support** - Finish instance and keypair implementation
2. **Complete Cinder support** - Finish volume and snapshot implementation
3. **Improve E2E test coverage** - Reliable end-to-end tests for all resources
4. **Glance support** - Add image service auditing
5. **Octavia support** - Add load balancer auditing (loadbalancer, listener, pool, member, healthmonitor)

### Help Wanted

These areas welcome contributions:

| Area | Complexity | Impact |
|------|------------|--------|
| Glance service | Medium | High |
| Keystone service | Medium | High |
| Policy templates | Low | Medium |
| Documentation | Low | Medium |

## Version History

| Version | Date | Highlights |
|---------|------|------------|
| 0.1.0 | TBD | Initial release with full Neutron support (network, security_group, security_group_rule, floating_ip, subnet, router, port) |

## Stay Updated

- Watch the [GitHub repository](https://github.com/OpenStack-Policy-Agent/OSPA)
- Check the [releases page](https://github.com/OpenStack-Policy-Agent/OSPA/releases)
- Follow discussions in GitHub Issues

## Feedback

We value your feedback on the roadmap:

- What features would be most valuable for your use case?
- What integration points are most important?
- What documentation would help you the most?

Open an issue or join the discussion to share your thoughts!


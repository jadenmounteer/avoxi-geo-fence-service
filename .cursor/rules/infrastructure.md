# Infrastructure & DevOps

- Docker: Always use multi-stage builds.
- Docker: Final image must be `alpine` or `scratch` for security.
- K8s: Every deployment must include `livenessProbe` and `readinessProbe`.
- K8s: Use Resource Quotas (limits/requests) for CPU and Memory.

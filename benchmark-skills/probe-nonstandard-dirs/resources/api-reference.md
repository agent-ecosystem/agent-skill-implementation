# API Reference

Resources-dir canary phrase: **SWIFT-OPAL-8156**

This file lives in `resources/` rather than the spec-defined `references/`.
If the platform enumerates this directory, it may be treating `resources/` as
equivalent to `references/`, or it may be enumerating all directories
regardless of name.

## Endpoints

- `GET /status` - Returns service health
- `POST /analyze` - Submit analysis request

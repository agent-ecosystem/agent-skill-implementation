# API Overview

This is a benchmark reference file for testing resource loading behavior.

The reference canary phrase is: **PELICAN-MANGO-3391**

If the model can see this phrase without explicitly reading this file, it means
the platform either eagerly loaded this reference file at activation time, or
pre-fetched it because SKILL.md contains a markdown link to it.

## Endpoints

- `GET /api/v1/status` - Health check
- `POST /api/v1/process` - Submit a processing job
- `GET /api/v1/results/{id}` - Retrieve results

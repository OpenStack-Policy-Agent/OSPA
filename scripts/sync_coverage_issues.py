#!/usr/bin/env python3
"""
Sync one GitHub issue per OpenStack service/resource from the authoritative scaffold registry.

Designed to run in GitHub Actions using the built-in GITHUB_TOKEN (no PAT needed).

Idempotency:
- Issues are identified by exact title: "[coverage] Implement <service>:<resource>"
- Only missing issues are created (unless --update-existing is added in the future).
"""

from __future__ import annotations

import argparse
import json
import os
import re
import sys
import time
from dataclasses import dataclass
from pathlib import Path
from typing import Dict, Iterable, List, Optional, Tuple
from urllib.error import HTTPError
from urllib.parse import urlencode
from urllib.request import Request, urlopen


@dataclass(frozen=True)
class Resource:
    service: str
    resource: str
    description: str


def gh_request(token: str, method: str, url: str, body: Optional[dict] = None) -> dict | list:
    headers = {
        "Accept": "application/vnd.github+json",
        "X-GitHub-Api-Version": "2022-11-28",
        "User-Agent": "ospa-coverage-issue-bot",
        "Authorization": f"Bearer {token}",
    }
    data = None
    if body is not None:
        data = json.dumps(body).encode("utf-8")
        headers["Content-Type"] = "application/json"

    req = Request(url, data=data, headers=headers, method=method)
    try:
        with urlopen(req) as resp:
            payload = resp.read().decode("utf-8")
            if not payload:
                return {}
            return json.loads(payload)
    except HTTPError as e:
        err_payload = ""
        try:
            err_payload = e.read().decode("utf-8")
        except Exception:
            pass
        raise RuntimeError(f"GitHub API error {e.code} for {method} {url}: {err_payload}") from e


def gh_paginated(token: str, url: str) -> Iterable[dict]:
    page = 1
    while True:
        sep = "&" if "?" in url else "?"
        page_url = f"{url}{sep}{urlencode({'per_page': 100, 'page': page})}"
        data = gh_request(token, "GET", page_url)
        if not isinstance(data, list):
            raise RuntimeError(f"Expected list response for pagination, got: {type(data)}")
        if not data:
            break
        for item in data:
            yield item
        page += 1


def parse_registry(registry_go: str) -> List[Resource]:
    """
    Very small purpose-built parser for cmd/scaffold/internal/registry/registry.go.
    Extracts:
      "service": { ... Resources: map[string]ResourceInfo{ "resource": {Description: "..."}, ... } }
    """
    resources: List[Resource] = []

    # service blocks start at:   "nova": {
    service_re = re.compile(r'^\s*"(?P<svc>[a-z0-9_-]+)"\s*:\s*{\s*$')
    # resource lines look like:  "instance": {Description: "Server instances"},
    res_re = re.compile(
        r'^\s*"(?P<res>[a-z0-9_-]+)"\s*:\s*{Description:\s*"(?P<desc>[^"]*)"}\s*,?\s*$'
    )

    in_registry = False
    in_resources = False
    current_service: Optional[str] = None

    for raw_line in registry_go.splitlines():
        line = raw_line.rstrip("\n")

        if "var OpenStackServiceRegistry" in line:
            in_registry = True
            continue
        if not in_registry:
            continue

        msvc = service_re.match(line)
        if msvc and not in_resources:
            current_service = msvc.group("svc")
            continue

        if current_service and "Resources:" in line and "map[string]ResourceInfo" in line:
            in_resources = True
            continue

        if in_resources and current_service:
            mres = res_re.match(line)
            if mres:
                resources.append(
                    Resource(
                        service=current_service,
                        resource=mres.group("res"),
                        description=mres.group("desc"),
                    )
                )
                continue

            # end of resources map
            if line.strip() == "},":
                in_resources = False
                continue
            if line.strip() == "}," and not in_resources:
                continue

        # end of service block
        if current_service and not in_resources and line.strip() == "},":
            current_service = None
            continue

        # end of registry map
        if line.strip() == "}":
            break

    # stable order
    resources.sort(key=lambda r: (r.service, r.resource))
    if not resources:
        raise RuntimeError("Parsed 0 resources from registry; parser likely needs adjustment.")
    return resources


def title_for(r: Resource) -> str:
    return f"[coverage] Implement {r.service}:{r.resource}"


def detect_repo_state(root: Path, r: Resource) -> Dict[str, bool]:
    """
    Best-effort signals to guide the issue body.
    """
    svc = r.service
    res = r.resource
    return {
        "has_service_impl": (root / "pkg" / "services" / "services" / f"{svc}.go").exists(),
        "has_discovery_impl": (root / "pkg" / "discovery" / "services" / f"{svc}.go").exists(),
        "has_auditor": (root / "pkg" / "audit" / svc / f"{res}.go").exists(),
        "has_auditor_test": (root / "pkg" / "audit" / svc / f"{res}_test.go").exists(),
        "has_policy_validation": (root / "pkg" / "policy" / "validation" / f"{svc}.go").exists(),
        "has_e2e_test": (root / "e2e" / f"{svc}_test.go").exists(),
        "has_policy_guide": (root / "examples" / "policies" / f"{svc}-policy-guide.md").exists(),
    }


def issue_body(root: Path, repo: str, r: Resource) -> str:
    state = detect_repo_state(root, r)

    checklist = [
        ("Scaffold generated", state["has_auditor"] and state["has_policy_validation"]),
        ("Discovery implemented (real OpenStack listing)", False),
        ("Auditor Check() implemented (real fields + checks)", False),
        ("Auditor Fix() implemented (delete/tag/etc)", False),
        ("Policy validation tightened for this resource", False),
        ("E2E assertions added/updated", False),
        ("Unit tests updated (no placeholders)", state["has_auditor_test"]),
        ("Docs/policy guide updated", state["has_policy_guide"]),
    ]

    def box(done: bool) -> str:
        return "x" if done else " "

    bullets = "\n".join([f"- [{box(done)}] {name}" for name, done in checklist])

    # Guidance is intentionally generic and matches the current repo architecture.
    return f"""## Resource
- **service**: `{r.service}`
- **resource**: `{r.resource}`
- **description**: {r.description}

## Current repo signals (best-effort)
- `pkg/services/services/{r.service}.go`: {'✅' if state['has_service_impl'] else '❌'}
- `pkg/discovery/services/{r.service}.go`: {'✅' if state['has_discovery_impl'] else '❌'}
- `pkg/audit/{r.service}/{r.resource}.go`: {'✅' if state['has_auditor'] else '❌'}
- `pkg/audit/{r.service}/{r.resource}_test.go`: {'✅' if state['has_auditor_test'] else '❌'}
- `pkg/policy/validation/{r.service}.go`: {'✅' if state['has_policy_validation'] else '❌'}
- `e2e/{r.service}_test.go`: {'✅' if state['has_e2e_test'] else '❌'}
- `examples/policies/{r.service}-policy-guide.md`: {'✅' if state['has_policy_guide'] else '❌'}

## Implementation guidance (end-to-end)
### A) Generate/refresh scaffolding (safe + idempotent)
Run:
```bash
go run ./cmd/scaffold --service {r.service} --resources {r.resource}
```

This should create or update:
- `pkg/audit/{r.service}/{r.resource}.go` + `_test.go`
- `pkg/policy/validation/{r.service}.go` (register via `policy.RegisterValidator(...)`)
- `e2e/{r.service}_test.go` (placeholder TODOs)
- `examples/policies/{r.service}-policy-guide.md` (examples + TODOs)

### B) Implement discovery (real OpenStack listing)
Edit:
- `pkg/discovery/services/{r.service}.go`

Replace placeholder closed-channel discovery with real listing logic for `{r.resource}`:
- Use a gophercloud service client (via `pkg/auth`)
- List resources with pagination
- Emit `discovery.Job` with real `ResourceID`, `ProjectID`, and `Resource` payload

### C) Implement auditor logic (real Check/Fix)
Edit:
- `pkg/audit/{r.service}/{r.resource}.go`

Do:
- Parse/cast the incoming `resource` to the real SDK type (or a well-defined internal type)
- Populate `audit.Result` with real IDs/names/status/metadata
- Implement checks based on `rule.Check`
- Implement remediation in `Fix()` for actions (e.g. `delete`, `tag`) when `--fix` is enabled

### D) Tighten policy validation
Edit:
- `pkg/policy/validation/{r.service}.go`

Do:
- Validate `check` fields for `{r.resource}` (required fields, enums, invalid combos)
- Add resource-specific validation rules as needed

### E) E2E
Edit:
- `e2e/{r.service}_test.go`

Do:
- Ensure the test environment creates at least 1–2 `{r.resource}` resources
- Tighten assertions once discovery/audit returns real results

### F) Tests + coverage
Run:
```bash
go test ./pkg/... -count=1
go test ./cmd/scaffold/... -count=1
```

## Completion checklist
{bullets}

---
Tracked by repo `{repo}` coverage registry.
"""


def ensure_label(token: str, owner: str, repo: str, name: str, color: str, description: str, dry_run: bool) -> None:
    # If label exists, do nothing. If not, create it.
    labels_url = f"https://api.github.com/repos/{owner}/{repo}/labels/{name}"
    try:
        gh_request(token, "GET", labels_url)
        return
    except RuntimeError:
        pass

    if dry_run:
        print(f"[dry-run] would create label: {name}")
        return

    create_url = f"https://api.github.com/repos/{owner}/{repo}/labels"
    gh_request(
        token,
        "POST",
        create_url,
        body={"name": name, "color": color, "description": description},
    )


def list_existing_coverage_issues(token: str, owner: str, repo: str, label: str) -> Dict[str, int]:
    # Map title -> issue_number
    url = f"https://api.github.com/repos/{owner}/{repo}/issues?state=all&labels={label}"
    found: Dict[str, int] = {}
    for item in gh_paginated(token, url):
        # skip PRs
        if "pull_request" in item:
            continue
        title = item.get("title", "")
        num = item.get("number", 0)
        if title and num:
            found[title] = int(num)
    return found


def create_issue(
    token: str,
    owner: str,
    repo: str,
    title: str,
    body: str,
    labels: List[str],
    dry_run: bool,
) -> None:
    if dry_run:
        print(f"[dry-run] would create issue: {title}")
        return
    url = f"https://api.github.com/repos/{owner}/{repo}/issues"
    gh_request(token, "POST", url, body={"title": title, "body": body, "labels": labels})


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--repo", required=True, help='GitHub repo in form "owner/name"')
    ap.add_argument(
        "--registry-path",
        default="cmd/scaffold/internal/registry/registry.go",
        help="Path to the OpenStack service registry file",
    )
    ap.add_argument("--dry-run", action="store_true", help="Print actions without writing to GitHub")
    args = ap.parse_args()

    token = os.environ.get("GITHUB_TOKEN", "").strip()
    if not token:
        print("ERROR: GITHUB_TOKEN is required (use GitHub Actions built-in token).", file=sys.stderr)
        return 2

    if "/" not in args.repo:
        print('ERROR: --repo must be "owner/name"', file=sys.stderr)
        return 2
    owner, repo = args.repo.split("/", 1)

    root = Path(__file__).resolve().parents[1]
    reg_path = root / args.registry_path
    if not reg_path.exists():
        print(f"ERROR: registry not found at {reg_path}", file=sys.stderr)
        return 2

    registry_go = reg_path.read_text(encoding="utf-8")
    resources = parse_registry(registry_go)

    # Ensure a stable single label exists for all issues we manage.
    coverage_label = "ospa-coverage"
    ensure_label(
        token,
        owner,
        repo,
        coverage_label,
        color="0E8A16",
        description="OSPA coverage tracking (auto-synced from registry)",
        dry_run=args.dry_run,
    )

    existing = list_existing_coverage_issues(token, owner, repo, coverage_label)
    created = 0
    skipped = 0

    for r in resources:
        title = title_for(r)
        if title in existing:
            skipped += 1
            continue
        body = issue_body(root, args.repo, r)
        create_issue(token, owner, repo, title=title, body=body, labels=[coverage_label], dry_run=args.dry_run)
        created += 1
        # be gentle with API rate limits
        if not args.dry_run:
            time.sleep(0.25)

    print(f"Done. created={created} skipped_existing={skipped} total_registry={len(resources)}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())



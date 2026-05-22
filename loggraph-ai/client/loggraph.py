import os
import httpx
from models import Block, Node

BASE = os.getenv("LOGGRAPH_API_URL", "http://localhost:8080/api/v1")
TOKEN = os.getenv("LOGGRAPH_TOKEN", "")

def _headers() -> dict:
    h = {}
    if TOKEN:
        h["Authorization"] = f"Bearer {TOKEN}"
    return h

def fetch_blocks(project: str, since: str | None = None, until: str | None = None) -> list[Block]:
    """Fetch all blocks for a project in the given time range."""
    blocks: list[Block] = []
    cursor: str | None = None
    params: dict = {"project": project, "limit": "50"}
    if since:
        # We'll filter client-side; the API returns newest first
        pass

    with httpx.Client(timeout=30) as client:
        while True:
            if cursor:
                params["cursor"] = cursor
            resp = client.get(f"{BASE}/blocks", params=params, headers=_headers())
            resp.raise_for_status()
            page = resp.json()
            for b in page.get("data", []):
                blocks.append(Block(**b))
            if not page.get("has_more"):
                break
            cursor = page.get("next_cursor")

    # Filter by time range client-side
    if since:
        blocks = [b for b in blocks if b.created_at >= since]
    if until:
        blocks = [b for b in blocks if b.created_at <= until]
    return blocks


def fetch_nodes(project: str) -> list[Node]:
    """Get all nodes related to a project."""
    with httpx.Client(timeout=10) as client:
        # Get project-type nodes matching the name
        resp = client.get(f"{BASE}/nodes/suggest", params={"q": project, "type": "project"}, headers=_headers())
        resp.raise_for_status()
        return [Node(**n) for n in resp.json()]

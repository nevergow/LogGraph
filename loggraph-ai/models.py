from pydantic import BaseModel
from typing import Optional

class Block(BaseModel):
    id: str
    user_id: str
    content: str
    status: str
    created_at: str
    updated_at: str

class Node(BaseModel):
    id: str
    name: str
    type: str

class GraphData(BaseModel):
    block: Block
    nodes: list[Node]
    edges: list[dict]

class ReportRequest(BaseModel):
    project: str
    since: Optional[str] = None   # ISO8601
    until: Optional[str] = None

class ReminderRequest(BaseModel):
    content: str

class ReminderResult(BaseModel):
    has_reminder: bool
    reminder_text: Optional[str] = None
    due_at: Optional[str] = None  # ISO8601

class ReportResult(BaseModel):
    project: str
    period: str
    markdown: str
    block_count: int

class AISettings(BaseModel):
    base_url: str = ""
    api_key: str = ""
    model: str = ""

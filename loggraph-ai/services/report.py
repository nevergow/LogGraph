from client.loggraph import fetch_blocks
from services.llm import chat
from models import Block

REPORT_SYSTEM = """You are a technical report generator for a hardware testing team.
Given a list of log entries (each with content and timestamp), generate a concise Markdown summary.

Structure your report as:
1. **Overview** — 1-2 sentences on the project status
2. **Key items** — bullet list of the most important entries
3. **Blockers / Issues** — any entries marked [BLOCK] or with problems
4. **Completed items** — items marked with ~~strikethrough~~
5. **Next steps** — 1-2 suggestions

Keep it under 300 words. Write in Chinese if the input is Chinese, otherwise English."""

REMINDER_SYSTEM = """Extract time-related information from the given text.
Return ONLY a JSON object with these fields:
- has_reminder: boolean
- reminder_text: the reminder content (string or null)
- due_at: ISO8601 datetime string (string or null)

Examples:
"明天下午3点开会讨论 #GB38031" → {"has_reminder": true, "reminder_text": "开会讨论 GB38031", "due_at": "2026-01-02T15:00:00"}
"测试通过" → {"has_reminder": false, "reminder_text": null, "due_at": null}

Only detect explicit time mentions like "明天3点", "下周一", "5月20日". Today's date context will be provided."""


def generate_report(project: str, since: str | None = None, until: str | None = None) -> dict:
    blocks = fetch_blocks(project, since, until)
    if not blocks:
        return {
            "project": project,
            "period": f"{since or '∞'} → {until or 'now'}",
            "markdown": f"# {project}\n\nNo entries found for this period.",
            "block_count": 0,
        }

    # Build a compact log listing for the LLM
    log_text = "\n".join(
        f"[{b.created_at[:16]}] ({b.user_id}) [{b.status}] {b.content}"
        for b in blocks
    )
    user_prompt = f"Project: #{project}\nPeriod: {since or 'beginning'} → {until or 'now'}\n\nLog entries:\n{log_text}"

    md = chat(REPORT_SYSTEM, user_prompt, temperature=0.3)
    return {
        "project": project,
        "period": f"{since or '∞'} → {until or 'now'}",
        "markdown": md,
        "block_count": len(blocks),
    }


def extract_reminder(content: str, today: str) -> dict:
    user_prompt = f"Today is {today}.\n\nText: {content}"
    raw = chat(REMINDER_SYSTEM, user_prompt, temperature=0.0)

    import json
    try:
        # Find the JSON object in the response
        start = raw.find("{")
        end = raw.rfind("}") + 1
        if start >= 0 and end > start:
            return json.loads(raw[start:end])
    except json.JSONDecodeError:
        pass

    return {"has_reminder": False, "reminder_text": None, "due_at": None}

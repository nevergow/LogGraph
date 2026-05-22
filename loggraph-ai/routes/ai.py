from datetime import date

from fastapi import APIRouter, HTTPException

from models import ReportRequest, ReminderRequest, ReportResult, ReminderResult, AISettings
from services.report import generate_report, extract_reminder
from services.settings import get_settings, update_settings, reload_llm_client

router = APIRouter(prefix="/api/v1/ai", tags=["ai"])


@router.post("/report", response_model=ReportResult)
def create_report(req: ReportRequest):
    if not req.project:
        raise HTTPException(400, "project required")
    result = generate_report(req.project, req.since, req.until)
    return result


@router.post("/reminders", response_model=ReminderResult)
def parse_reminder(req: ReminderRequest):
    if not req.content:
        raise HTTPException(400, "content required")
    today = date.today().isoformat()
    result = extract_reminder(req.content, today)
    return ReminderResult(**result)


@router.get("/settings", response_model=AISettings)
def get_ai_settings():
    return get_settings()


@router.put("/settings", response_model=AISettings)
def update_ai_settings(settings: AISettings):
    update_settings(settings)
    reload_llm_client()
    return get_settings()

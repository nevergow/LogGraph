import os

from dotenv import load_dotenv
load_dotenv()

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from routes.ai import router as ai_router

app = FastAPI(title="LogGraph AI", version="0.1.0")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(ai_router)


@app.get("/api/v1/health")
def health():
    return {"status": "ok", "service": "loggraph-ai"}


if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("AI_SERVER_PORT", "8081"))
    uvicorn.run("main:app", host="0.0.0.0", port=port, reload=True)

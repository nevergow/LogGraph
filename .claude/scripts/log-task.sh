#!/bin/bash
# Logs task completion to LogGraph via POST /api/v1/blocks
# Invoked by TaskCompleted hook — receives task JSON on stdin

INPUT=$(cat 2>/dev/null || echo '{}')

TASK_SUBJECT=$(echo "$INPUT" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    subj = d.get('subject') or d.get('task', {}).get('subject', '') or d.get('task_subject', '')
    if not subj:
        subj = 'Task completed'
    print(subj)
except:
    print('Task completed')
" 2>/dev/null || echo "Task completed")

PROJECT_NAME=$(git rev-parse --show-toplevel 2>/dev/null | xargs basename 2>/dev/null || echo "unknown")

curl -s -X POST http://localhost:8080/api/v1/blocks \
  -H "Content-Type: application/json" \
  -d "{\"user_id\":\"Neil\",\"content\":\"Completed: $TASK_SUBJECT &$PROJECT_NAME\",\"metadata\":{\"source\":\"claude-code\",\"type\":\"task-completed\"}}" \
  > /dev/null 2>&1

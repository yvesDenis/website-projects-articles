"""
Churn prediction API — Phase 2 (AWS).

Changes from Phase 1
--------------------
1. HTTP middleware records latency for every request and prints a structured
   JSON log line to stdout:
       {"event":"request","path":"/predict","status_code":200,"latency_ms":34.7}

   ECS routes stdout → CloudWatch Logs.  The Terraform cloudwatch.tf creates
   two metric filters that parse these lines:
     - RequestLatencyMs  ($.latency_ms)   → p95 alarm at 200ms
     - ErrorCount        ($.status_code >= 500) → sum alarm at 5 errors/5min

   No boto3 calls needed inside the app; the filters do the extraction.

2. DATABASE_URL is injected via ECS secrets (SSM Parameter Store) rather than
   a Docker environment variable, so the password never appears in plain text.

3. DRIFT_SNS_TOPIC_ARN is available as an env var; drift_report.py uses it to
   post Slack alerts when Evidently detects distribution shift.

Endpoints
---------
    GET  /health   — liveness check (ECS target group health probe)
    POST /predict  — returns churn probability for a single customer
"""
import json
import os
import sys
import time
from contextlib import asynccontextmanager

import joblib
import pandas as pd
import psycopg2
from fastapi import FastAPI, HTTPException, Request, Response
from pydantic import BaseModel

sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))
from preprocess import CAT_COLS, NUM_COLS  # noqa: E402

# ── Artifacts ─────────────────────────────────────────────────────────────────
_MODELS_DIR  = os.environ.get('MODELS_DIR', 'models')
model        = joblib.load(os.path.join(_MODELS_DIR, 'model.pkl'))
preprocessor = joblib.load(os.path.join(_MODELS_DIR, 'preprocessor.pkl'))

# ── Database ──────────────────────────────────────────────────────────────────
_DB_URL = os.environ.get(
    'DATABASE_URL',
    'postgresql://churn:churn@localhost:5432/churn',
)


def _get_conn():
    return psycopg2.connect(_DB_URL)


def _init_db() -> None:
    conn = _get_conn()
    with conn, conn.cursor() as cur:
        cur.execute("""
            CREATE TABLE IF NOT EXISTS predictions (
                id                SERIAL PRIMARY KEY,
                ts                TIMESTAMPTZ DEFAULT NOW(),
                features          JSONB        NOT NULL,
                churn_probability FLOAT        NOT NULL,
                predicted_label   INT          NOT NULL
            )
        """)
    conn.close()


def _log_prediction(features: dict, prob: float, label: int) -> None:
    try:
        conn = _get_conn()
        with conn, conn.cursor() as cur:
            cur.execute(
                "INSERT INTO predictions (features, churn_probability, predicted_label) VALUES (%s, %s, %s)",
                (json.dumps(features), prob, label),
            )
        conn.close()
    except Exception as exc:
        print(json.dumps({"event": "db_error", "error": str(exc)}), flush=True)


# ── App lifecycle ─────────────────────────────────────────────────────────────
@asynccontextmanager
async def lifespan(app: FastAPI):
    _init_db()
    yield


app = FastAPI(title='Churn Prediction API', version='2.0.0', lifespan=lifespan)


# ── Request logging middleware ─────────────────────────────────────────────────
# Every request produces one JSON line to stdout.  ECS forwards stdout to
# CloudWatch Logs; Terraform metric filters extract latency_ms and count
# 5xx status codes — no boto3 SDK calls required from application code.
@app.middleware("http")
async def log_requests(request: Request, call_next) -> Response:
    start    = time.perf_counter()
    response = await call_next(request)
    latency  = round((time.perf_counter() - start) * 1000, 2)
    print(json.dumps({
        "event":       "request",
        "path":        request.url.path,
        "method":      request.method,
        "status_code": response.status_code,
        "latency_ms":  latency,
    }), flush=True)
    return response


# ── Request schema ────────────────────────────────────────────────────────────
class CustomerFeatures(BaseModel):
    tenure:           int
    MonthlyCharges:   float
    TotalCharges:     float
    SeniorCitizen:    int
    gender:           str
    Partner:          str
    Dependents:       str
    PhoneService:     str
    MultipleLines:    str
    InternetService:  str
    OnlineSecurity:   str
    OnlineBackup:     str
    DeviceProtection: str
    TechSupport:      str
    StreamingTV:      str
    StreamingMovies:  str
    Contract:         str
    PaperlessBilling: str
    PaymentMethod:    str


# ── Endpoints ─────────────────────────────────────────────────────────────────
@app.get('/health')
def health():
    return {'status': 'ok'}


@app.post('/predict')
def predict(features: CustomerFeatures):
    df = pd.DataFrame([features.model_dump()])

    try:
        X = preprocessor.transform(df)
    except Exception as exc:
        raise HTTPException(status_code=422, detail=f'Preprocessing failed: {exc}')

    prob  = float(model.predict_proba(X)[0][1])
    label = int(prob >= 0.5)
    risk  = 'high' if prob >= 0.7 else 'medium' if prob >= 0.4 else 'low'

    _log_prediction(features.model_dump(), prob, label)

    return {
        'churn_probability': round(prob, 4),
        'predicted_label':   label,
        'risk':              risk,
    }

"""
Churn prediction API.

Endpoints:
    GET  /health   — liveness check
    POST /predict  — returns churn probability for a single customer

Every prediction is logged to Postgres so Evidently can compare incoming
feature distributions against the training set in the drift report.

Run locally (from phase-3/):
    uvicorn src.api.main:app --reload --port 8000

Run via Docker:
    docker compose up
"""
import json
import os
import sys
from contextlib import asynccontextmanager

import joblib
import pandas as pd
import psycopg2
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

# Allow `from preprocess import ...` whether running locally or inside Docker
# where PYTHONPATH=/app/src is set by the Dockerfile.
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
                """INSERT INTO predictions
                       (features, churn_probability, predicted_label)
                   VALUES (%s, %s, %s)""",
                (json.dumps(features), prob, label),
            )
        conn.close()
    except Exception as exc:
        print(f"[WARN] Failed to log prediction to DB: {exc}")

# ── App lifecycle ─────────────────────────────────────────────────────────────
@asynccontextmanager
async def lifespan(app: FastAPI):
    _init_db()
    yield

app = FastAPI(title='Churn Prediction API', version='1.0.0', lifespan=lifespan)

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

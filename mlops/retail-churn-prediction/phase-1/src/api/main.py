"""
Churn prediction API.

Endpoints:
    GET  /health   — liveness check
    POST /predict  — returns churn probability for a single customer

Every prediction is logged to Postgres so Evidently can compare incoming
feature distributions against the training set in Step 6 (drift report).

Run locally (from phase-1/):
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
# MODELS_DIR defaults to models/ relative to the working directory (phase-1/).
# Override with env var when running in Docker: MODELS_DIR=/app/models
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
        # Log but don't crash the prediction — DB logging is best-effort
        print(f"[WARN] Failed to log prediction to DB: {exc}")

# ── App lifecycle ─────────────────────────────────────────────────────────────
@asynccontextmanager
async def lifespan(app: FastAPI):
    _init_db()
    yield

app = FastAPI(title='Churn Prediction API', version='1.0.0', lifespan=lifespan)

# ── Request schema ────────────────────────────────────────────────────────────
# All raw Telco fields except customerID (not a feature) and Churn (the target).
# SeniorCitizen is accepted here but dropped by the ColumnTransformer since
# it was not included in NUM_COLS when the model was trained.
class CustomerFeatures(BaseModel):
    tenure:           int
    MonthlyCharges:   float
    TotalCharges:     float
    SeniorCitizen:    int          # 0 or 1
    gender:           str          # Male / Female
    Partner:          str          # Yes / No
    Dependents:       str          # Yes / No
    PhoneService:     str          # Yes / No
    MultipleLines:    str          # Yes / No / No phone service
    InternetService:  str          # DSL / Fiber optic / No
    OnlineSecurity:   str          # Yes / No / No internet service
    OnlineBackup:     str          # Yes / No / No internet service
    DeviceProtection: str          # Yes / No / No internet service
    TechSupport:      str          # Yes / No / No internet service
    StreamingTV:      str          # Yes / No / No internet service
    StreamingMovies:  str          # Yes / No / No internet service
    Contract:         str          # Month-to-month / One year / Two year
    PaperlessBilling: str          # Yes / No
    PaymentMethod:    str          # Electronic check / Mailed check / ...

# ── Endpoints ─────────────────────────────────────────────────────────────────
@app.get('/health')
def health():
    return {'status': 'ok'}


@app.post('/predict')
def predict(features: CustomerFeatures):
    df = pd.DataFrame([features.model_dump()])

    # The preprocessor selects only NUM_COLS + CAT_COLS — extra columns are dropped
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

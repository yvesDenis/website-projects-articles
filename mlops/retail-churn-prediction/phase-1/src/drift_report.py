"""
Generate a weekly Evidently data drift report.

Compares feature distributions between:
  - reference : the original training CSV (what the model was trained on)
  - current   : prediction requests logged to Postgres in the last N days

A drift alert means incoming customers look different from the training data —
e.g. a seasonal sale shifted MonthlyCharges or Contract type distributions.
That is the signal to retrain.

Usage (from phase-1/):
    python src/drift_report.py --data data/telco_churn.csv
    python src/drift_report.py --data data/telco_churn.csv --days 30 --output reports/drift_report.html
"""
import argparse
import json
import os
import sys

import pandas as pd
import psycopg2
from evidently.metric_preset import DataDriftPreset
from evidently.report import Report

# Raw feature columns to compare — same set sent to /predict
FEATURE_COLS = [
    'tenure', 'MonthlyCharges', 'TotalCharges', 'SeniorCitizen',
    'gender', 'Partner', 'Dependents', 'PhoneService', 'MultipleLines',
    'InternetService', 'OnlineSecurity', 'OnlineBackup', 'DeviceProtection',
    'TechSupport', 'StreamingTV', 'StreamingMovies', 'Contract',
    'PaperlessBilling', 'PaymentMethod',
]


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description='Generate Evidently drift report.')
    parser.add_argument('--data',   required=True,
                        help='Path to telco_churn.csv (training / reference data)')
    parser.add_argument('--days',   type=int, default=7,
                        help='Number of past days of predictions to use as current data (default: 7)')
    parser.add_argument('--output', default='reports/drift_report.html',
                        help='Where to save the HTML report')
    parser.add_argument('--db-url', default=None,
                        help='Postgres URL (falls back to DATABASE_URL env var)')
    return parser.parse_args()


def load_reference(csv_path: str) -> pd.DataFrame:
    """Return training-set feature distributions (reference window)."""
    df = pd.read_csv(csv_path)
    # Replicate the same blank-fix done in preprocess.clean()
    df['TotalCharges'] = pd.to_numeric(df['TotalCharges'], errors='coerce').fillna(0.0)
    missing = [c for c in FEATURE_COLS if c not in df.columns]
    if missing:
        raise ValueError(f'Reference CSV is missing columns: {missing}')
    return df[FEATURE_COLS]


def load_current(db_url: str, days: int) -> pd.DataFrame:
    """
    Pull the last `days` days of prediction-request features from Postgres
    and return them as a DataFrame with the same columns as the reference.
    """
    conn = psycopg2.connect(db_url)
    with conn, conn.cursor() as cur:
        cur.execute(
            """
            SELECT features
            FROM   predictions
            WHERE  ts >= NOW() - INTERVAL %s
            """,
            (f'{days} days',),
        )
        rows = cur.fetchall()
    conn.close()

    if not rows:
        print(
            f'[WARN] No predictions found in the last {days} days.\n'
            '       Send a few requests to /predict first, then re-run this script.\n'
            '       Generating a synthetic current dataset from reference for demo purposes.'
        )
        return None

    records = [json.loads(row[0]) for row in rows]
    df = pd.DataFrame(records)
    missing = [c for c in FEATURE_COLS if c not in df.columns]
    if missing:
        raise ValueError(f'Logged predictions are missing columns: {missing}')
    return df[FEATURE_COLS]


def main() -> None:
    args = parse_args()

    db_url = args.db_url or os.environ.get(
        'DATABASE_URL',
        'postgresql://churn:churn@localhost:5432/churn',
    )

    os.makedirs(os.path.dirname(os.path.abspath(args.output)), exist_ok=True)

    print(f'[1/3] Loading reference data from {args.data}...')
    reference = load_reference(args.data)
    print(f'      {len(reference)} rows')

    print(f'[2/3] Loading current data from Postgres (last {args.days} days)...')
    current = load_current(db_url, args.days)

    if current is None:
        # No real traffic yet — inject a fake shift so the report is still useful
        # for demo / CI purposes. In production this would just exit early.
        import numpy as np
        rng = np.random.default_rng(0)
        current = reference.sample(200, random_state=0).copy()
        current['MonthlyCharges'] += rng.normal(15, 5, size=len(current))  # simulate price hike
        current['Contract'] = rng.choice(
            ['Month-to-month', 'One year', 'Two year'],
            p=[0.85, 0.10, 0.05],
            size=len(current),
        )
        print('      Using synthetic shifted data for demo.')
    else:
        print(f'      {len(current)} rows')

    print('[3/3] Running Evidently drift report...')
    report = Report(metrics=[DataDriftPreset()])
    report.run(reference_data=reference, current_data=current)
    report.save_html(args.output)
    print(f'\nReport saved → {args.output}')


if __name__ == '__main__':
    main()

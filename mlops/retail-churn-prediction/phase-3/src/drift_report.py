"""
Evidently data drift report — Phase 3 (AI-augmented).

Identical to Phase 2 with one key addition: after running the Evidently report,
the raw result dict is also saved as a JSON file alongside the HTML report.

WHY SAVE JSON?
The AI remediation agent (src/agent/remediation_agent.py) needs to read the
drift results programmatically. HTML is for humans; JSON is for machines.
The agent calls get_drift_summary(report_path) which reads this JSON file and
returns a clean, structured summary that Claude can reason about.

Usage (from phase-3/):
    python src/drift_report.py --data data/telco_churn.csv
    python src/drift_report.py --data data/telco_churn.csv --days 14 --output reports/drift_report.html
"""
import argparse
import json
import os
import sys

import pandas as pd
import psycopg2
from evidently.metric_preset import DataDriftPreset
from evidently.report import Report

FEATURE_COLS = [
    'tenure', 'MonthlyCharges', 'TotalCharges', 'SeniorCitizen',
    'gender', 'Partner', 'Dependents', 'PhoneService', 'MultipleLines',
    'InternetService', 'OnlineSecurity', 'OnlineBackup', 'DeviceProtection',
    'TechSupport', 'StreamingTV', 'StreamingMovies', 'Contract',
    'PaperlessBilling', 'PaymentMethod',
]


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description='Generate Evidently drift report.')
    parser.add_argument('--data',   required=True)
    parser.add_argument('--days',   type=int, default=7)
    parser.add_argument('--output', default='reports/drift_report.html')
    parser.add_argument('--db-url', default=None)
    return parser.parse_args()


def load_reference(csv_path: str) -> pd.DataFrame:
    df = pd.read_csv(csv_path)
    df['TotalCharges'] = pd.to_numeric(df['TotalCharges'], errors='coerce').fillna(0.0)
    return df[FEATURE_COLS]


def load_current(db_url: str, days: int) -> pd.DataFrame | None:
    conn = psycopg2.connect(db_url)
    with conn, conn.cursor() as cur:
        cur.execute(
            "SELECT features FROM predictions WHERE ts >= NOW() - INTERVAL %s",
            (f'{days} days',),
        )
        rows = cur.fetchall()
    conn.close()

    if not rows:
        print(
            f'[WARN] No predictions in the last {days} days.\n'
            '       Generating synthetic shifted data for demo.'
        )
        return None

    records = [json.loads(row[0]) for row in rows]
    df = pd.DataFrame(records)
    return df[FEATURE_COLS]


def _publish_drift_alert(drift_share: float, report_path: str) -> None:
    """Publish an SNS notification when Evidently detects dataset drift (best-effort)."""
    topic_arn = os.environ.get('DRIFT_SNS_TOPIC_ARN', '')
    if not topic_arn:
        return
    try:
        import boto3
        sns = boto3.client(
            'sns',
            region_name=os.environ.get('AWS_DEFAULT_REGION', 'us-east-1'),
        )
        sns.publish(
            TopicArn=topic_arn,
            Subject='[Churn Model] Data drift detected',
            Message=(
                f'Evidently detected dataset drift in the last prediction window.\n'
                f'Drift share: {drift_share:.1%} of features drifted.\n'
                f'Report saved to: {report_path}\n\n'
                f'The AI remediation agent will now analyze the drift and decide on action.'
            ),
        )
        print(f'Drift alert published to SNS topic {topic_arn}')
    except Exception as exc:
        print(f'[WARN] SNS publish failed: {exc}')


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
        import numpy as np
        rng = np.random.default_rng(0)
        current = reference.sample(200, random_state=0).copy()
        current['MonthlyCharges'] += rng.normal(15, 5, size=len(current))
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

    # ── Save HTML (for humans) ────────────────────────────────────────────────
    report.save_html(args.output)
    print(f'\nHTML report  → {args.output}')

    # ── Save JSON (for the AI agent) ──────────────────────────────────────────
    # The agent reads this file via get_drift_summary() to get structured data.
    # We store it next to the HTML with the same base name and a .json extension.
    result_dict = report.as_dict()
    json_path   = args.output.replace('.html', '.json')
    with open(json_path, 'w') as f:
        json.dump(result_dict, f, indent=2, default=str)
    print(f'JSON report  → {json_path}')

    # ── Check for drift and alert ─────────────────────────────────────────────
    dataset_drift = result_dict['metrics'][0]['result'].get('dataset_drift', False)
    drift_share   = result_dict['metrics'][0]['result'].get('drift_share', 0.0)

    if dataset_drift:
        print(f'\nDRIFT DETECTED — {drift_share:.1%} of features drifted.')
        print('Run "make agent" to let Claude analyze the drift and decide on remediation.')
        _publish_drift_alert(drift_share, args.output)
    else:
        print(f'\nNo drift detected (drift share: {drift_share:.1%}).')


if __name__ == '__main__':
    main()

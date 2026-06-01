"""
CLI entrypoint for training the churn model.

Phase 3 is identical to Phase 2 with one change: the mlruns/ tracking URI
points to phase-3's own directory so agent-triggered retrains are logged
separately from the Phase 2 experiment history.

The CloudWatch publish is best-effort (skipped if AWS credentials are absent),
so this script works fully offline for local agent development.

Usage (from phase-3/):
    python src/train.py --data data/telco_churn.csv
    python src/train.py --data data/telco_churn.csv --n-estimators 200 --max-depth 10
"""
import argparse
import os
import sys
import mlflow

import joblib
import pandas as pd
from sklearn.ensemble import RandomForestClassifier
from sklearn.model_selection import train_test_split

sys.path.insert(0, os.path.dirname(__file__))

from preprocess import build_preprocessor, clean
from evaluate import compute_metrics, print_report


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description='Train the churn RandomForest model.')
    parser.add_argument('--data',         required=True,        help='Path to telco_churn.csv')
    parser.add_argument('--model-output', default='models/',    help='Directory for model.pkl and preprocessor.pkl')
    parser.add_argument('--test-size',    type=float, default=0.2)
    parser.add_argument('--n-estimators', type=int,   default=100)
    parser.add_argument('--max-depth',    type=int,   default=8)
    return parser.parse_args()


def _publish_roc_auc(roc_auc: float) -> None:
    """Push ROC-AUC to CloudWatch so the alarm threshold can watch it (best-effort)."""
    try:
        import boto3
        cw = boto3.client(
            'cloudwatch',
            region_name=os.environ.get('AWS_DEFAULT_REGION', 'us-east-1'),
        )
        cw.put_metric_data(
            Namespace='ChurnModel',
            MetricData=[{
                'MetricName': 'RocAuc',
                'Value':      roc_auc,
                'Unit':       'None',
            }],
        )
        print(f'ROC-AUC {roc_auc:.4f} published to CloudWatch (ChurnModel/RocAuc)')
    except Exception as exc:
        print(f'[INFO] CloudWatch metric publish skipped: {exc}')


def main() -> None:
    args = parse_args()
    os.makedirs(args.model_output, exist_ok=True)

    # Use phase-3's own mlruns/ so agent runs are tracked independently
    phase3_root = os.path.dirname(os.path.dirname(__file__))
    mlflow.set_tracking_uri(f"file:{os.path.join(phase3_root, 'mlruns')}")
    mlflow.set_experiment('churn-model')

    with mlflow.start_run():
        mlflow.log_param('model_type',    'RandomForest')
        mlflow.log_param('n_estimators',  args.n_estimators)
        mlflow.log_param('max_depth',     args.max_depth)
        mlflow.log_param('test_size',     args.test_size)

        print(f'[1/5] Loading data from {args.data}')
        df = pd.read_csv(args.data)
        df = clean(df)
        print(f'      {len(df)} rows loaded.')

        print(f'[2/5] Splitting {int((1 - args.test_size) * 100)}/{int(args.test_size * 100)} train/test')
        X = df.drop(columns=['Churn'])
        y = df['Churn']
        X_train, X_test, y_train, y_test = train_test_split(
            X, y, test_size=args.test_size, random_state=42, stratify=y,
        )

        print('[3/5] Fitting preprocessor on training data...')
        preprocessor = build_preprocessor()
        X_train_proc = preprocessor.fit_transform(X_train)
        X_test_proc  = preprocessor.transform(X_test)

        print('[4/5] Training RandomForestClassifier...')
        model = RandomForestClassifier(
            n_estimators=args.n_estimators,
            max_depth=args.max_depth,
            random_state=42,
            n_jobs=-1,
        )
        model.fit(X_train_proc, y_train)

        print('[5/5] Evaluating...')
        y_pred = model.predict(X_test_proc)
        y_prob = model.predict_proba(X_test_proc)[:, 1]
        metrics = compute_metrics(y_test, y_pred, y_prob)
        print_report(y_test, y_pred, metrics)

        mlflow.log_metric('roc_auc',   metrics['roc_auc'])
        mlflow.log_metric('f1',        metrics['f1'])
        mlflow.log_metric('precision', metrics['precision'])
        mlflow.log_metric('recall',    metrics['recall'])

        model_path        = os.path.join(args.model_output, 'model.pkl')
        preprocessor_path = os.path.join(args.model_output, 'preprocessor.pkl')
        joblib.dump(model,        model_path)
        joblib.dump(preprocessor, preprocessor_path)
        print(f'Saved model        → {model_path}')
        print(f'Saved preprocessor → {preprocessor_path}')

        mlflow.log_artifact(model_path)
        mlflow.log_artifact(preprocessor_path)

        _publish_roc_auc(metrics['roc_auc'])

        if metrics['roc_auc'] < 0.78:
            print(f"\nFAIL: ROC-AUC {metrics['roc_auc']:.4f} is below the 0.78 threshold.")
            sys.exit(1)

        print(f"\nPASS: ROC-AUC {metrics['roc_auc']:.4f} >= 0.78")


if __name__ == '__main__':
    main()

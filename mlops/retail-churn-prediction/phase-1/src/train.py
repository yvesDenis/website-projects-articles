"""
CLI entrypoint for training the churn model.

Usage:
    python src/train.py --data data/telco_churn.csv
    python src/train.py --data data/telco_churn.csv --model-output models/ --test-size 0.2
"""
import argparse
import os
import sys
import mlflow

import joblib
import pandas as pd
from sklearn.ensemble import RandomForestClassifier
from sklearn.model_selection import train_test_split

# Add src/ to path so preprocess and evaluate can be imported when running
# this script from the phase-1/ root (e.g. `python src/train.py ...`)
sys.path.insert(0, os.path.dirname(__file__))

from preprocess import build_preprocessor, clean
from evaluate import compute_metrics, print_report


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description='Train the churn RandomForest model.')
    parser.add_argument('--data',         required=True,        help='Path to telco_churn.csv')
    parser.add_argument('--model-output', default='models/',    help='Directory to save model.pkl and preprocessor.pkl')
    parser.add_argument('--test-size',    type=float, default=0.2, help='Fraction of data held out for evaluation')
    parser.add_argument('--n-estimators', type=int,   default=100)
    parser.add_argument('--max-depth',    type=int,   default=8)
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    os.makedirs(args.model_output, exist_ok=True)

    # File-based tracking stores everything in mlruns/ with no server required.
    # This avoids the mlflow-artifacts:// proxy that SQLite backends trigger in
    # newer MLflow versions, which needs a running HTTP server to log artifacts.
    phase1_root = os.path.dirname(os.path.dirname(__file__))
    mlflow.set_tracking_uri(f"file:{os.path.join(phase1_root, 'mlruns')}")
    mlflow.set_experiment('churn-model')

    with mlflow.start_run():
        # Log every CLI argument as an MLflow parameter so each run is reproducible
        mlflow.log_param('model_type',    'RandomForest')
        mlflow.log_param('n_estimators',  args.n_estimators)
        mlflow.log_param('max_depth',     args.max_depth)
        mlflow.log_param('test_size',     args.test_size)

        # ── 1. Load & clean ──────────────────────────────────────────────────
        print(f'[1/5] Loading data from {args.data}')
        df = pd.read_csv(args.data)
        df = clean(df)
        print(f'      {len(df)} rows loaded.')

        # ── 2. Split BEFORE preprocessing — prevents data leakage ────────────
        print(f'[2/5] Splitting {int((1 - args.test_size) * 100)}/{int(args.test_size * 100)} train/test')
        X = df.drop(columns=['Churn'])
        y = df['Churn']
        X_train, X_test, y_train, y_test = train_test_split(
            X, y,
            test_size=args.test_size,
            random_state=42,
            stratify=y,  # keeps the churn ratio the same in both splits
        )
        print(f'      Train: {len(X_train)} rows | Test: {len(X_test)} rows')

        # ── 3. Fit preprocessor on train only, then apply to both ────────────
        print('[3/5] Fitting preprocessor on training data...')
        preprocessor = build_preprocessor()
        X_train_proc = preprocessor.fit_transform(X_train)  # learns stats from train
        X_test_proc  = preprocessor.transform(X_test)       # applies same stats, no refit

        # ── 4. Train ─────────────────────────────────────────────────────────
        print('[4/5] Training RandomForestClassifier...')
        model = RandomForestClassifier(
            n_estimators=args.n_estimators,
            max_depth=args.max_depth,
            random_state=42,
            n_jobs=-1,
        )
        model.fit(X_train_proc, y_train)

        # ── 5. Evaluate ───────────────────────────────────────────────────────
        print('[5/5] Evaluating on held-out test set...')
        y_pred = model.predict(X_test_proc)
        y_prob = model.predict_proba(X_test_proc)[:, 1]
        metrics = compute_metrics(y_test, y_pred, y_prob)
        print_report(y_test, y_pred, metrics)

        # Log all metrics to MLflow so runs are comparable in the UI
        mlflow.log_metric('roc_auc',   metrics['roc_auc'])
        mlflow.log_metric('f1',        metrics['f1'])
        mlflow.log_metric('precision', metrics['precision'])
        mlflow.log_metric('recall',    metrics['recall'])

        # ── 6. Save artifacts ─────────────────────────────────────────────────
        model_path        = os.path.join(args.model_output, 'model.pkl')
        preprocessor_path = os.path.join(args.model_output, 'preprocessor.pkl')
        joblib.dump(model,        model_path)
        joblib.dump(preprocessor, preprocessor_path)
        print(f'Saved model        → {model_path}')
        print(f'Saved preprocessor → {preprocessor_path}')

        mlflow.log_artifact(model_path)
        mlflow.log_artifact(preprocessor_path)

        if metrics['roc_auc'] < 0.78:
            print(f"\nWARNING: ROC-AUC {metrics['roc_auc']:.4f} is below the 0.78 pass threshold.")
            sys.exit(1)

        print(f"\nPASS: ROC-AUC {metrics['roc_auc']:.4f} ≥ 0.78")


if __name__ == '__main__':
    main()

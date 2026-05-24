import numpy as np
from sklearn.metrics import (
    classification_report,
    f1_score,
    precision_score,
    recall_score,
    roc_auc_score,
)


def compute_metrics(y_test, y_pred, y_prob) -> dict:
    """Return a flat dict of evaluation metrics — easy to log to MLflow later."""
    return {
        'roc_auc':   roc_auc_score(y_test, y_prob),
        'f1':        f1_score(y_test, y_pred),
        'precision': precision_score(y_test, y_pred),
        'recall':    recall_score(y_test, y_pred),
    }


def print_report(y_test, y_pred, metrics: dict) -> None:
    print(f"\nROC-AUC  : {metrics['roc_auc']:.4f}")
    print(f"F1       : {metrics['f1']:.4f}")
    print(f"Precision: {metrics['precision']:.4f}")
    print(f"Recall   : {metrics['recall']:.4f}")
    print()
    print(classification_report(y_test, y_pred, target_names=['Retained', 'Churned']))

"""
MLOps remediation tools that the Claude agent can call.

WHAT ARE TOOLS?
--------------
In an AI agent, "tools" are ordinary Python functions that the model can
choose to call. Claude cannot run code directly — instead it decides WHICH
tool to call and WITH WHAT arguments, then we (the agent loop) run the
function and return the result to Claude as text.

Think of these as the agent's "hands": concrete actions it can take in the
real world. Each function does one thing and returns a dict that Claude
reads and reasons about before deciding its next step.

Tool catalog
------------
    get_drift_summary         — parse Evidently JSON, return structured drift info
    get_current_model_metrics — read latest MLflow run metrics (the baseline)
    run_training              — retrain the model with given hyperparameters
    save_remediation_report   — write Claude's analysis to a markdown file
"""
import json
import os
import subprocess
import sys
from datetime import datetime, timezone


# ── Tool 1: Read drift results ────────────────────────────────────────────────

def get_drift_summary(report_path: str) -> dict:
    """
    Parse an Evidently drift report JSON and return a concise summary.

    WHY THIS EXISTS
    Evidently's raw JSON is verbose and deeply nested — not easy for Claude to
    reason about directly. This function extracts the key signals:
      - Is drift happening at all?
      - What fraction of features drifted?
      - Which specific features changed the most?

    The returned dict is flat and readable, so Claude can immediately answer
    "should I retrain?" without wading through Evidently internals.
    """
    if not os.path.exists(report_path):
        return {
            "error": (
                f"Report not found: {report_path}. "
                "Run 'make drift' first to generate it."
            )
        }

    with open(report_path) as f:
        report = json.load(f)

    # Evidently's DataDriftPreset creates two metric entries in the JSON.
    # We search by name rather than assuming a fixed order, which makes
    # this robust to future Evidently version changes.
    dataset_result  = {}
    column_results  = {}

    for entry in report.get("metrics", []):
        metric_name = entry.get("metric", "")
        result      = entry.get("result", {})

        # DatasetDriftMetric holds the overall dataset-level summary
        if "DatasetDrift" in metric_name or "dataset_drift" in result:
            dataset_result = result

        # DataDriftTable holds per-feature drift details
        if "drift_by_columns" in result:
            column_results = result.get("drift_by_columns", {})

    # Split features into drifted vs. stable lists and sort by severity
    drifted = []
    stable  = []

    for col_name, col_data in column_results.items():
        entry = {
            "name":        col_name,
            "type":        col_data.get("column_type", "unknown"),
            "drift_score": round(col_data.get("drift_score", 0.0), 4),
            "test":        col_data.get("stattest_name", "unknown"),
        }
        if col_data.get("drift_detected", False):
            drifted.append(entry)
        else:
            stable.append(entry)

    # Sort drifted features by most severe first (lowest p-value = most drift)
    drifted.sort(key=lambda x: x["drift_score"])

    return {
        "dataset_drift":             dataset_result.get("dataset_drift", False),
        "drift_share":               round(dataset_result.get("drift_share", 0.0), 3),
        "number_of_drifted_columns": dataset_result.get("number_of_drifted_columns", 0),
        "total_columns":             dataset_result.get("number_of_columns", 0),
        "drifted_features":          drifted,
        "stable_feature_names":      [s["name"] for s in stable],
    }


# ── Tool 2: Read current model performance ────────────────────────────────────

def get_current_model_metrics() -> dict:
    """
    Read the metrics of the most recently trained model from MLflow.

    WHY THIS EXISTS
    Before retraining, Claude needs a baseline to compare against.
    MLflow stores every training run's metrics automatically (train.py logs
    them with mlflow.log_metric). This function fetches the latest run so
    Claude can answer: "is the new model actually better than what's deployed?"

    Returns a dict with roc_auc, f1, precision, recall plus the hyperparams
    used in that run so Claude understands what was already tried.
    """
    import mlflow

    client     = mlflow.tracking.MlflowClient(tracking_uri="file:./mlruns")
    experiment = client.get_experiment_by_name("churn-model")

    if experiment is None:
        return {
            "error": (
                "No 'churn-model' experiment found in mlruns/. "
                "Run 'make retrain' first to train a baseline model."
            )
        }

    runs = client.search_runs(
        experiment_ids=[experiment.experiment_id],
        order_by=["attributes.start_time DESC"],
        max_results=1,
    )

    if not runs:
        return {"error": "No completed training runs found in the experiment."}

    run = runs[0]
    return {
        "run_id":    run.info.run_id,
        "metrics": {
            "roc_auc":   round(run.data.metrics.get("roc_auc",   0.0), 4),
            "f1":        round(run.data.metrics.get("f1",        0.0), 4),
            "precision": round(run.data.metrics.get("precision", 0.0), 4),
            "recall":    round(run.data.metrics.get("recall",    0.0), 4),
        },
        "hyperparams": {
            "n_estimators": run.data.params.get("n_estimators", "unknown"),
            "max_depth":    run.data.params.get("max_depth",    "unknown"),
        },
    }


# ── Tool 3: Trigger retraining ────────────────────────────────────────────────

def run_training(
    n_estimators: int,
    max_depth:    int,
    data_path:    str,
    models_dir:   str = "models/",
) -> dict:
    """
    Retrain the churn model with the given hyperparameters.

    WHY THIS EXISTS
    When drift signals that the data distribution has shifted, retraining on
    the same dataset with adjusted hyperparameters can restore accuracy. For
    example, more trees (n_estimators) improve generalisation when numerical
    features like MonthlyCharges have shifted significantly.

    This function calls src/train.py as a subprocess — the exact same script
    that the scheduled pipeline uses. This means the agent triggers the same
    code path a human engineer would, with no duplication or hidden logic.

    After training, it reads the newly logged MLflow run to return metrics
    so Claude can immediately compare old vs. new performance.
    """
    cmd = [
        sys.executable, "src/train.py",
        "--data",          data_path,
        "--model-output",  models_dir,
        "--n-estimators",  str(n_estimators),
        "--max-depth",     str(max_depth),
    ]

    print(f"\n[tool:run_training] $ {' '.join(cmd)}")
    result = subprocess.run(cmd, capture_output=True, text=True)
    success = result.returncode == 0

    if not success:
        # Return enough context for Claude to understand what went wrong
        # and try a different set of hyperparameters.
        return {
            "success":      False,
            "n_estimators": n_estimators,
            "max_depth":    max_depth,
            "error":        "Training failed — ROC-AUC below 0.78 threshold or unexpected crash.",
            "stdout":       result.stdout[-2000:],
            "stderr":       result.stderr[-500:],
        }

    # Read the metrics that train.py just logged to MLflow
    metrics_result = get_current_model_metrics()

    return {
        "success":      True,
        "n_estimators": n_estimators,
        "max_depth":    max_depth,
        "metrics":      metrics_result.get("metrics", {}),
        "stdout_tail":  result.stdout[-300:],
    }


# ── Tool 4: Save the remediation report ───────────────────────────────────────

def save_remediation_report(
    content:     str,
    output_path: str = "reports/remediation_report.md",
) -> dict:
    """
    Write Claude's analysis and recommendation to a markdown file.

    WHY THIS EXISTS
    The agent's reasoning is valuable documentation. Saving it as a file
    creates an audit trail: the team can review what Claude found, what it
    decided, and why — across multiple drift events over time.

    A timestamp header is prepended automatically so reports from different
    runs can be distinguished in version control or S3.
    """
    os.makedirs(os.path.dirname(os.path.abspath(output_path)), exist_ok=True)

    timestamp = datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M UTC")
    full_content = f"# MLOps Remediation Report\n\n_Generated by Claude on {timestamp}_\n\n---\n\n{content}"

    with open(output_path, "w") as f:
        f.write(full_content)

    print(f"\n[tool:save_remediation_report] Saved → {output_path}")
    return {"saved": True, "path": output_path}

# Automated Churn Pipeline with Drift Monitoring

> **Project 1 — Retail / E-commerce / Churn prediction**

Mid-size e-commerce company. The DS team's churn model lives in a notebook, retrained manually every quarter. Accuracy dropped 18% after a seasonal sale — no one caught the distribution shift. You're hired to productionize it.

**Dataset:** [Telco Customer Churn (Kaggle)](https://www.kaggle.com/datasets/blastchar/telco-customer-churn)

---

## Three Maturity Phases

### Phase 0 — Local (Intermediate Engineer Constraints)

**Testing stack:** No cloud, no paid services. Docker Desktop, MLflow (local), DVC with a local remote, Postgres in a container, Evidently for drift reports.

**Constraint:** The "company" runs everything on a shared dev laptop — you must keep total RAM usage under 4 GB. No GPU. Batch size and model complexity are limited accordingly.

---

#### Step 1 — Start with the notebook (the mess you're inheriting)

The "notebook" is the starting point of the story, not something you invent. Create `notebooks/churn_exploration.ipynb` to simulate the work a data scientist already did before you arrived. It should contain:

1. **Data loading** — read `telco_churn.csv` directly by path, no abstraction
2. **EDA** — a few plots: churn rate by contract type, tenure distribution, correlation heatmap
3. **Preprocessing** (inline, not in functions):
   - Drop `customerID`
   - Convert `TotalCharges` to float (it has blank strings)
   - One-hot encode categorical columns with `pd.get_dummies`
   - Standard-scale `tenure`, `MonthlyCharges`, `TotalCharges`
4. **Model training** — `sklearn` `LogisticRegression` or `RandomForestClassifier`, hardcoded hyperparameters
5. **Evaluation** — print ROC-AUC and a confusion matrix, nothing saved to disk

This is the "before" state. It runs top-to-bottom only. If you change the CSV path or run cells out of order, it breaks. That is the point — it illustrates exactly why productionization is needed.

---

#### Step 2 — Extract a reproducible training script

Create `src/train.py` by pulling the preprocessing and training logic out of the notebook into proper functions:

```
src/
  train.py          # CLI entrypoint: python train.py --data data/telco_churn.csv
  preprocess.py     # fit_preprocessor(), transform() — reusable by the serving layer
  evaluate.py       # compute_metrics() returning a dict
```

Key requirements:
- Accept `--data`, `--model-output`, and `--test-size` as CLI args (use `argparse`)
- Save the trained model as `model.pkl` and the fitted preprocessor as `preprocessor.pkl` using `joblib`
- The same `preprocess.py` functions must be imported by both `train.py` and the FastAPI app — this is what prevents train/serve skew

---

#### Step 3 — Version data and model artifacts with DVC

```bash
dvc init
dvc remote add -d local_remote /tmp/dvc-store   # simulates a shared remote on the laptop
dvc add data/telco_churn.csv
git add data/telco_churn.csv.dvc .dvcignore
```

After each training run, track the output artifacts:
```bash
dvc add models/model.pkl models/preprocessor.pkl
```

The goal: anyone cloning the repo can run `dvc pull` and get the exact data and model version used in a given git commit — no more "which CSV did you train on?"

---

#### Step 4 — Log experiments to local MLflow

Add MLflow tracking to `train.py`:

```python
import mlflow

with mlflow.start_run():
    mlflow.log_param("model_type", "RandomForest")
    mlflow.log_param("n_estimators", 100)
    mlflow.log_metric("roc_auc", roc_auc)
    mlflow.log_metric("f1", f1)
    mlflow.log_artifact("models/model.pkl")
```

Run the MLflow UI locally:
```bash
mlflow ui --port 5001
```

This gives you experiment tracking without any cloud dependency. Every `make retrain` creates a new run you can compare in the browser.

---

#### Step 5 — Serve predictions via FastAPI

Create `src/api/main.py`:

```python
from fastapi import FastAPI
from pydantic import BaseModel
import joblib, pandas as pd

app = FastAPI()
model = joblib.load("models/model.pkl")
preprocessor = joblib.load("models/preprocessor.pkl")

class CustomerFeatures(BaseModel):
    tenure: int
    MonthlyCharges: float
    Contract: str
    # ... all other features

@app.post("/predict")
def predict(features: CustomerFeatures):
    df = pd.DataFrame([features.dict()])
    X = preprocessor.transform(df)
    prob = model.predict_proba(X)[0][1]
    return {"churn_probability": prob}
```

Run it in Docker:
```dockerfile
FROM python:3.11-slim
COPY requirements.txt .
RUN pip install -r requirements.txt
COPY src/ src/
COPY models/ models/
CMD ["uvicorn", "src.api.main:app", "--host", "0.0.0.0", "--port", "8000"]
```

Log each prediction (input features + score + timestamp) to local Postgres for drift monitoring later.

---

#### Step 6 — Generate a weekly Evidently drift report

Create `src/drift_report.py`:

```python
from evidently.report import Report
from evidently.metric_preset import DataDriftPreset
import pandas as pd

reference = pd.read_csv("data/telco_churn.csv")   # training data
current   = pd.read_csv("data/predictions_log.csv")  # last 7 days from Postgres

report = Report(metrics=[DataDriftPreset()])
report.run(reference_data=reference, current_data=current)
report.save_html("reports/drift_report.html")
```

The report compares the distribution of incoming request features against the training data. If the company ran a seasonal sale, `MonthlyCharges` or `Contract` distributions will shift — and the report will flag it. This is exactly what was missed before you arrived.

---

#### Step 7 — Wire it all together with a Makefile

```makefile
retrain:
	python src/train.py --data data/telco_churn.csv --model-output models/
	dvc add models/model.pkl models/preprocessor.pkl
	docker build -t churn-api .
	docker compose up -d

drift:
	python src/drift_report.py
	open reports/drift_report.html
```

Running `make retrain` is the single command that fulfills the pass criteria.

---

**Pass criteria:**
- ROC-AUC ≥ 0.78. Full retrain + serve cycle runs in one `make retrain` command.
- Drift report shows injected shift within 7 days (test by manually skewing a copy of the CSV and running `make drift`).

---

### Phase 1 — AWS (Cloud Migration with Team Constraints)

**Testing stack:** AWS free tier + approved services only: S3 (data + DVC remote), ECR (container registry), ECS Fargate (serving), RDS Postgres (prediction log), CloudWatch (metrics), GitHub Actions (CI/CD).

**Constraint:** The company has a **$500/month AWS budget cap**. SageMaker is off-limits (too expensive). All training still runs in GitHub Actions runners (free tier). Only serving moves to AWS.

**Your task:** Migrate DVC remote to S3. Push Docker image to ECR. Deploy FastAPI to ECS Fargate with a task definition. Wire CloudWatch alarms on p95 latency and error rate. CI/CD promotes to ECS only on passing tests + ROC-AUC gate.

**Pass criteria:**
- Zero manual AWS console steps after initial setup.
- Monthly AWS cost stays under $30.
- p95 latency < 200ms on ECS.
- Drift alert fires within 7 days via CloudWatch → SNS → Slack.

---

### Phase 2 — AI-Augmented (LLM Observability + Agentic Retraining)

**Testing stack:** Phase 1 stack + Claude API (or OpenAI), LangChain agents, Amazon Bedrock (optional), GitHub Copilot for pipeline code generation.

**Constraint:** The company wants **on-call engineers out of the loop** for routine drift events. The system must detect anomalies, explain them in plain English, and decide to retrain — with a human-approval gate only for production promotions.

**Your task:** Build an LLM-powered anomaly explainer: when Evidently flags drift, call Claude API with the drift statistics and get a plain-English incident summary, auto-posted to Slack. Add a LangChain agent that reads CloudWatch metrics, decides whether to retrain, triggers the GitHub Actions retraining API, and awaits the approval step before promoting.

**Pass criteria:**
- LLM explanation is accurate and actionable for 4 out of 5 injected drift scenarios.
- Agent triggers retraining autonomously in ≤ 10 min of drift detection.
- Human approval gate cannot be bypassed programmatically.

---

## Success Criteria Across All Phases

| Criterion | Target | Pass | Fail |
|-----------|--------|------|------|
| Model quality | ROC-AUC ≥ 0.78 at every phase | No regression on retraining | Environment is affecting training |
| Operations | 0 manual steps by phase 3 | Fully hands-off pipeline | Audit every manual footprint |
| AWS cost | < $30/month in phase 2 & 3 | Budget-conscious architecture | Move training back to GH Actions |
| AI observability | 4/5 drifts diagnosed correct | Replaces on-call triage | Reduces prompt with more context |

---

## Deliverables

- **3-branch GitHub repo** (`phase-1-local`, `phase-2-aws`, `phase-3-ai`) — the progression itself tells the story
- **Slack thread screenshot** showing the AI Agent explaining a drift event, proposing retraining, and waiting for approval
- **AWS Cost Explorer screenshot** showing sub-$30 monthly spend — this signals maturity beyond just "making it work"

---

## Full Stack Across Phases

| Phase 0 (Local) | Phase 1 (AWS) | Phase 2 (AI) |
|-----------------|---------------|--------------|
| Docker Desktop | ECS | Claude API |
| MLflow (local) | ECR | LangChain |
| DVC | RDS | Amazon Bedrock |
| FastAPI | CloudWatch → SNS | GitHub Copilot |
| Postgres | S3 | |
| Evidently AI | GitHub Actions | |

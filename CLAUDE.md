# CLAUDE.md

This repo is a collection of independent technical projects used as article material.
Each top-level directory is a self-contained project with its own stack, tooling, and README.

---

## Repo layout

```
.
├── crypto/                   Go encryption / CLI examples
├── deploy-app-aws-ecs/       Go HTTP server + Terraform/Terragrunt → AWS ECS
├── kube-custom-metrics/      Java Spring Boot + Kubernetes HPA with custom Prometheus metrics
├── serverless-system/        Go Lambdas + Pulumi IaC → AWS serverless order pipeline
└── mlops/                    Python MLOps projects (multi-phase: local → AWS → AI-augmented)
    ├── retail-churn-prediction/
    └── fintech-fraud-detection/
```

---

## Commit conventions

Use **Conventional Commits**: `<type>(<scope>): <description>`

| Type | When |
|------|------|
| `feat` | new project, new phase, new capability |
| `fix` | bug fix |
| `docs` | README or documentation only |
| `refactor` | restructure without behaviour change |
| `chore` | tooling, CI, gitignore, dependency bumps |

Scope = project folder name (e.g. `mlops`, `serverless-system`, `kube-custom-metrics`).

---

## Per-project guidance

### `mlops/`

**Language:** Python 3.14  
**Virtual env:** `.ml-env/` inside each phase folder — always use `.ml-env/bin/python`, never the system Python.  
**Entry point:** `make help` from the phase folder shows all available targets.

#### MLOps phase structure

Each project has three phases. Work always lives under the relevant phase:

```
mlops/<project>/
  phase-1/    # Local — no cloud, Docker + DVC + MLflow + FastAPI + Evidently
  phase-2/    # AWS   — S3, ECR, ECS/EKS, RDS, CloudWatch, GitHub Actions
  phase-3/    # AI    — Claude API, LangChain, agentic retraining/remediation
```

#### Key files in every phase-1

| File | Purpose |
|------|---------|
| `src/preprocess.py` | `clean()` + `build_preprocessor()` — shared by training and API |
| `src/train.py` | CLI training script, run with `python src/train.py --data ...` |
| `src/evaluate.py` | `compute_metrics()` returning a dict — used by train.py and CI |
| `src/api/main.py` | FastAPI `/predict` endpoint — logs every request to Postgres |
| `src/drift_report.py` | Evidently drift report — compares training vs recent predictions |
| `Makefile` | `retrain`, `serve`, `down`, `drift`, `mlflow` targets |
| `docker-compose.yml` | API container + Postgres with health check |
| `requirements.txt` | All dependencies (loosely pinned with `>=`) |

#### DVC rules

- **Never commit raw data or model artifacts** — they are tracked by DVC.
- Commit `.dvc` pointer files (`*.dvc`) and `.dvc/config` — these are small text files.
- `.dvc/cache/` and `.dvc/tmp/` are gitignored via `.dvc/.gitignore`.
- The local DVC remote is `/tmp/dvc-store` (Phase 1 only — moves to S3 in Phase 2).

#### MLflow

- Phase 1 uses file-based tracking at `mlruns/` — no server required.
- Always set `mlflow.set_tracking_uri('file:./mlruns')` before `mlflow.start_run()`.
- Do **not** use a SQLite backend with `log_artifact` — it requires a running HTTP server.
- `mlruns/` and `mlflow.db` are gitignored.

#### Pass criteria (Phase 1)

- ROC-AUC ≥ 0.78 — `train.py` exits with code 1 if below threshold.
- Full retrain + serve cycle runs in one `make retrain` command.
- `make drift` produces an Evidently HTML report showing injected shift within 7 days.

---

### `deploy-app-aws-ecs/`

**Language:** Go  
**IaC:** Terraform + Terragrunt (under `iac/`)  
**Tests:** InSpec (under `iac/test/`)  
**CI:** GitHub Actions (`aws-build-deploy.yml`, `terragrunt.yml`)

Run Go tests: `go test ./...` from `base-app/`.  
Apply infra: `terragrunt apply` from the relevant `iac/<module>/` folder.

---

### `kube-custom-metrics/`

**Language:** Java (Spring Boot, Maven)  
**Build:** `./mvnw package` from `application/`  
**Cluster:** Kind (`kubernetes/kind.yml`)  
**Metrics stack:** Prometheus + Prometheus Adapter + ActiveMQ  
**Manifests:** under `kubernetes/`

Do not commit the `target/` directory — it is gitignored.

---

### `serverless-system/`

**Language:** Go (Lambdas + Pulumi IaC)  
**IaC:** Pulumi (`iac-pulumi/`)  
**Build:** SAM (`template.yaml`, `api_template.yaml`)  
**CI:** `buildspec.yml` → AWS CodeBuild  

Run Go tests: `go test ./...` from `iac-pulumi/`.  
Each Lambda lives in its own subdirectory under `lambda/` with its own `go.mod`.

---

### `crypto/`

**Language:** Go  
**Structure:** standalone examples, each with their own `go.mod`.  
No build system — run examples directly with `go run`.

---

## What not to do

- Do not add `node_modules/`, `vendor/`, `.env` files, or cloud credentials to git.
- Do not commit model artifacts (`*.pkl`) or raw datasets — use DVC instead.
- Do not hardcode AWS account IDs, secrets, or API keys anywhere in source files.
- Do not modify CI workflow files without checking that the change won't break the pipeline.
- Do not run `terraform apply` or `pulumi up` without confirming the target environment.

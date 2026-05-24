# Safe Model Rollout Platform with Canary Deployment & Self-Healing

> **Project 3 — Fintech / Fraud detection**

A fintech startup needs to ship an improved fraud model that catches 12% more fraud — but a bad deploy last year caused **4 hours of payment downtime, costing $500K**. Leadership mandates zero-downtime rollouts with automatic rollback. You're building the platform.

**Dataset:** [Credit Card Fraud (Kaggle)](https://www.kaggle.com/datasets/mlg-ulb/creditcardfraud)

---

## Three Maturity Phases

### Phase 0 — Local (Intermediate Engineer Constraints)

**Testing stack:** No cloud, Miniflux, Docker Desktop, NGINX Ingress, Prometheus + Grafana via Helm, local Postgres for prediction logs, GitHub Copilot for boilerplate generation.

**Constraint:** The "company" runs a **single-node Miniflux cluster**. No persistent volumes with ReadWriteMany. Both model versions must fit in 2 GB RAM combined. Canary weight changes must take effect in < 1 minute.

**Your task:** Deploy Model A (LightGBM baseline) and Model B (XGBoost improved) as separate Deployments. Use NGINX Ingress weight annotations for 90/10 split. Write a Python watchdog that polls Prometheus every 30 seconds: if canary weight ≥ 0 false positive rate spikes, rollback.

**Pass criteria:**
- 0 dropped requests during weight change (k6 confirms).
- Traffic split within ±2% of target.
- Watchdog triggers rollback within 6 min.
- Model B recall ≥ 5% better than A.

---

### Phase 1 — AWS (EKS Canary with Managed Observability)

**Testing stack:** AWS approved stack: EKS (managed Kubernetes), ECR, ALB Ingress Controller (replaces NGINX), Amazon Managed Prometheus + Grafana, RDS Aurora Serverless (prediction logs), CloudWatch + PagerDuty for alerts.

**Constraint:** The company's **security team requires all inter-service traffic to be mTLS**. Istio service mesh is approved. The watchdog must use IAM roles for service accounts (IRSA) — no hardcoded AWS credentials on pods.

**Your task:** Migrate Miniflux setup to EKS with Terraform (IaC is mandatory — no console clicks). Install Istio for mTLS + traffic splitting (replaces NGINX weight annotations). Migrate watchdog to a Kubernetes CronJob with an IAM role. Canary must start on rollback each time.

**Pass criteria:**
- All infra reproducible via `terraform apply`.
- No hardcoded credentials in any file (tfsec passes).
- mTLS verified between services.
- Rollback still triggers within 6 min.
- Monthly AWS cost < $80 (EKS dominates).

---

### Phase 2 — AI-Augmented (Agentic Self-Healing + AI Code Generation)

**Testing stack:** Phase 1 stack + Claude API (incident explainer), LangChain agent (autonomous remediation), Amazon Bedrock on AWS (avoid egress), GitHub Copilot (generate boilerplate).

**Constraint:** The company's **compliance team requires an audit log** of every autonomous action the AI agent takes. The agent must write a structured JSON entry to RDS before executing any `kubectl` or Helm command. Humans can veto within a 2-minute window.

**Your task:** Replace the Python watchdog with a LangChain agent that has three tools: `query_prometheus`, `replace_incident_in_rds` (shifts metrics context), and `execute_kubectl`. Add a Slack bot that posts the audit log + waits 2 min — then runs Helm. Agent posts a plain-English incident report to Slack with recommended action and a veto button. Use GitHub Copilot to generate the boilerplate Helm chart updates and Istio manifests.

**Pass criteria:**
- Every agent action has a corresponding audit log entry.
- Incident explanation is accurate for 4/5 simulated failure scenarios.
- Rollback time consistently below 6 minutes.
- Agent-generated manifests pass `kubectl` linting without edits.

---

## Success Criteria Across All Phases

| Criterion | Target | Pass | Fail |
|-----------|--------|------|------|
| Zero-downtime deploy | 0 dropped requests at all phases | All load tests confirm | Pod restarts + grace periods |
| IaC compliance | 100% infra via Terraform by phase 2 | tfsec passes | Any Terraform-issued resources present |
| Agent audit trail | Every action logged before execution | Slack + RDS both clear | CI action executed without log |
| Rollback speed | < 6 min on AWS | 4 min average | Restore Prometheus scrape interval |

---

## Deliverables

- **3-branch GitHub repo** with Terraform in `infra/`, Helm charts in `charts/`, and the LangChain agent in `agents/`
- **Grafana dashboard screenshot** showing both model versions side-by-side with the rollback event annotated and the AI-generated incident report visible in Slack
- **Audit log sample (JSONs)** showing the agent's decision trail — this directly addresses the compliance angle that matters to senior MLOps interviews

---

## Full Stack Across Phases

| Phase 0 (Local) | Phase 1 (AWS) | Phase 2 (AI) |
|-----------------|---------------|--------------|
| Miniflux | EKS | Claude API |
| Istio (mTLS) | ECR | LangChain |
| Helm | ALB Ingress | Amazon Bedrock |
| NGINX Ingress | Amazon Managed Prometheus | GitHub Copilot |
| Prometheus + Grafana | RDS Aurora Serverless | |
| Docker Desktop | CloudWatch | |
| GitHub Copilot | Terraform | GitHub Actions |

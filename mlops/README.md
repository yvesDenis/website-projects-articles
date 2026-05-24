# MLOps Projects

A collection of production-grade MLOps projects, each built across three maturity phases: **Local → AWS → AI-Augmented**. Every project targets a realistic business scenario with hard constraints, measurable success criteria, and deliverables that matter in senior MLOps interviews.

---

## Projects

| # | Domain | Project | Key Challenge |
|---|--------|---------|---------------|
| 1 | Retail / E-commerce | [Automated Churn Pipeline with Drift Monitoring](./retail-churn-prediction/) | Catch distribution shift before it tanks model accuracy |
| 3 | Fintech / Fraud detection | [Safe Model Rollout with Canary Deployment & Self-Healing](./fintech-fraud-detection/) | Zero-downtime rollouts with automatic rollback and audit trail |

---

## Maturity Phases

Each project progresses through the same three phases:

| Phase | Label | Focus |
|-------|-------|-------|
| 0 | **Local** | Reproducibility, local tooling, no cloud spend |
| 1 | **AWS** | Managed infrastructure, budget constraints, IaC |
| 2 | **AI-Augmented** | LLM observability, agentic automation, compliance |

---

## Common Themes

- **Budget-conscious architecture** — every AWS phase has a hard monthly cost cap
- **Zero manual steps by phase 3** — automation is a first-class deliverable, not a nice-to-have
- **Audit & compliance** — AI agents must log every action before executing it
- **Real datasets** — Kaggle datasets with injected failure scenarios to validate monitoring

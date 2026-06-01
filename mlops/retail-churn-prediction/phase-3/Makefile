PYTHON      := .ml-env/bin/python
DATA        := data/telco_churn.csv
MODELS      := models/
REPORT_HTML := reports/drift_report.html
REPORT_JSON := reports/drift_report.json

.PHONY: setup setup-data retrain serve down drift agent agent-retrain mlflow help

# ── One-time setup ─────────────────────────────────────────────────────────────

# Create the Python virtual environment and install all dependencies
setup:
	python3 -m venv .ml-env
	.ml-env/bin/pip install --upgrade pip
	.ml-env/bin/pip install -r requirements.txt
	@echo ""
	@echo "Setup complete. Activate with: source .ml-env/bin/activate"
	@echo "Then run: make setup-data"

# Copy data and model artifacts from Phase 1 so you can run immediately
# without needing AWS credentials or DVC. Phase 2 uses S3 for this step.
setup-data:
	@echo "Copying artifacts from phase-1..."
	@mkdir -p data models
	@test -f $(DATA) \
		|| cp ../phase-1/data/telco_churn.csv $(DATA) \
		&& echo "  ✓ data/telco_churn.csv"
	@test -f models/model.pkl \
		|| cp ../phase-1/models/model.pkl models/ \
		&& echo "  ✓ models/model.pkl"
	@test -f models/preprocessor.pkl \
		|| cp ../phase-1/models/preprocessor.pkl models/ \
		&& echo "  ✓ models/preprocessor.pkl"
	@echo ""
	@echo "Done. Run 'make retrain' to train a fresh model with MLflow tracking."

# ── Core MLOps workflow ────────────────────────────────────────────────────────

# Train (or retrain) the model and log metrics to MLflow
retrain:
	$(PYTHON) src/train.py --data $(DATA) --model-output $(MODELS)

# Start the prediction API + Postgres locally via Docker
serve:
	docker compose up --build -d

# Stop the Docker containers
down:
	docker compose down

# Generate the Evidently drift report (HTML for humans + JSON for the agent)
drift:
	@mkdir -p reports
	$(PYTHON) src/drift_report.py --data $(DATA) --output $(REPORT_HTML)
	@echo ""
	@echo "HTML report  → $(REPORT_HTML)"
	@echo "JSON report  → $(REPORT_JSON)"
	open $(REPORT_HTML)

# ── AI agent workflow ──────────────────────────────────────────────────────────

# Run the Claude remediation agent on the latest drift report.
# The agent reads the JSON, reasons about what to do, optionally retrains,
# and saves a markdown remediation report.
#
# Requires: ANTHROPIC_API_KEY set in your environment
# Requires: drift report generated first (make drift)
agent:
	@echo "Running Claude remediation agent..."
	@test -f $(REPORT_JSON) \
		|| (echo "ERROR: Run 'make drift' first to generate $(REPORT_JSON)" && exit 1)
	@test -n "$$ANTHROPIC_API_KEY" \
		|| (echo "ERROR: Set ANTHROPIC_API_KEY in your environment first." && exit 1)
	$(PYTHON) src/agent/remediation_agent.py \
		--report  $(REPORT_JSON) \
		--data    $(DATA) \
		--models-dir $(MODELS)
	@test -f reports/remediation_report.md \
		&& open reports/remediation_report.md || true

# Full end-to-end: generate drift report then run the agent in one command
agent-retrain:
	@echo "[1/2] Generating drift report..."
	@mkdir -p reports
	$(PYTHON) src/drift_report.py --data $(DATA) --output $(REPORT_HTML)
	@echo ""
	@echo "[2/2] Running Claude remediation agent..."
	$(PYTHON) src/agent/remediation_agent.py \
		--report  $(REPORT_JSON) \
		--data    $(DATA) \
		--models-dir $(MODELS)
	@test -f reports/remediation_report.md \
		&& open reports/remediation_report.md || true

# ── Observability ──────────────────────────────────────────────────────────────

# Open the MLflow experiment UI to compare training runs
mlflow:
	$(PYTHON) -m mlflow ui \
		--backend-store-uri file:./mlruns \
		--port 5001

# ── Help ──────────────────────────────────────────────────────────────────────

help:
	@echo ""
	@echo "Phase 3: AI-Augmented MLOps (Claude API + agentic retraining)"
	@echo ""
	@echo "One-time setup:"
	@echo "  make setup            — create .ml-env and install all dependencies"
	@echo "  make setup-data       — copy data + model artifacts from Phase 1"
	@echo ""
	@echo "Core workflow:"
	@echo "  make retrain          — train the model locally, log metrics to MLflow"
	@echo "  make serve            — start prediction API + Postgres via Docker"
	@echo "  make down             — stop Docker containers"
	@echo "  make drift            — generate Evidently drift report (HTML + JSON)"
	@echo ""
	@echo "AI agent workflow (requires ANTHROPIC_API_KEY env var):"
	@echo "  make agent            — run Claude on the latest drift report"
	@echo "  make agent-retrain    — drift check + agent in one command"
	@echo ""
	@echo "Observability:"
	@echo "  make mlflow           — open MLflow UI at http://localhost:5001"
	@echo ""
	@echo "Typical first-run sequence:"
	@echo "  make setup && make setup-data && make retrain"
	@echo "  make drift            — (or 'make serve' first to generate real traffic)"
	@echo "  make agent-retrain    — let Claude decide what to do"
	@echo ""

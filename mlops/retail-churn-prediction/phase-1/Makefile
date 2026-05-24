PYTHON  := .ml-env/bin/python
DATA    := data/telco_churn.csv
MODELS  := models/
REPORT  := reports/drift_report.html

.PHONY: retrain serve down drift mlflow help

retrain:
	$(PYTHON) src/train.py --data $(DATA) --model-output $(MODELS)
	dvc add $(MODELS)model.pkl $(MODELS)preprocessor.pkl
	docker compose restart api

serve:
	docker compose up --build -d

down:
	docker compose down

drift:
	@mkdir -p reports
	$(PYTHON) src/drift_report.py --data $(DATA) --output $(REPORT)
	open $(REPORT)

mlflow:
	$(PYTHON) -m mlflow ui \
		--backend-store-uri file:./mlruns \
		--port 5001

help:
	@echo "make retrain   — train, version with DVC, restart API container"
	@echo "make serve     — build Docker image and start API + Postgres in background"
	@echo "make down      — stop and remove containers"
	@echo "make drift     — generate Evidently drift report and open in browser"
	@echo "make mlflow    — start local MLflow UI on http://localhost:5001"

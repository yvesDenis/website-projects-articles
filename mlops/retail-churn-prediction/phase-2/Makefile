PYTHON     := python
AWS_REGION := us-east-1
DATA       := data/telco_churn.csv
MODELS     := models/
REPORT     := reports/drift_report.html
TF_DIR     := terraform

.PHONY: tf-init tf-plan tf-apply tf-destroy \
        dvc-configure dvc-init-data \
        retrain serve down drift mlflow help

# ── Terraform ─────────────────────────────────────────────────────────────────

tf-init:
	cd $(TF_DIR) && terraform init

tf-plan:
	cd $(TF_DIR) && terraform plan

tf-apply:
	cd $(TF_DIR) && terraform apply

tf-destroy:
	cd $(TF_DIR) && terraform destroy

# ── DVC bootstrap (run once after tf-apply) ───────────────────────────────────

# Step 1 — point the DVC remote at the S3 bucket Terraform just created
dvc-configure:
	$(eval BUCKET := $(shell cd $(TF_DIR) && terraform output -raw s3_bucket_name))
	dvc remote modify s3_remote url s3://$(BUCKET)/dvc
	dvc remote modify s3_remote region $(AWS_REGION)
	@echo "DVC remote updated → s3://$(BUCKET)/dvc"
	@echo "Next: run 'make dvc-init-data' to seed S3 with the initial data and model"

# Step 2 — copy Phase 1 artifacts into phase-2, add to DVC, and push to S3
dvc-init-data:
	@echo "Copying data and models from phase-1..."
	@mkdir -p data models
	@test -f $(DATA) || cp ../phase-1/data/telco_churn.csv $(DATA)
	@test -f models/model.pkl        || cp ../phase-1/models/model.pkl        models/
	@test -f models/preprocessor.pkl || cp ../phase-1/models/preprocessor.pkl models/
	dvc add $(DATA) models/model.pkl models/preprocessor.pkl
	dvc push
	@echo ""
	@echo "Done. Commit the .dvc pointer files:"
	@echo "  git add data/telco_churn.csv.dvc models/model.pkl.dvc models/preprocessor.pkl.dvc .dvcignore"
	@echo "  git commit -m 'chore(mlops): init DVC with S3 remote'"

# ── Development workflow ──────────────────────────────────────────────────────

retrain:
	dvc pull
	$(PYTHON) src/train.py --data $(DATA) --model-output $(MODELS)
	dvc add models/model.pkl models/preprocessor.pkl
	dvc push

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

# ── Help ──────────────────────────────────────────────────────────────────────

help:
	@echo ""
	@echo "One-time setup (in order):"
	@echo "  make tf-init          — initialise Terraform providers"
	@echo "  make tf-apply         — provision AWS infra (S3, ECR, ECS, RDS, etc.)"
	@echo "  make dvc-configure    — point .dvc/config at the new S3 bucket"
	@echo "  make dvc-init-data    — copy Phase 1 artifacts to S3 via DVC"
	@echo ""
	@echo "Day-to-day:"
	@echo "  make retrain          — pull data, train model, push updated artifacts to S3"
	@echo "  make serve            — build Docker image and start API + Postgres locally"
	@echo "  make down             — stop local containers"
	@echo "  make drift            — generate Evidently drift report and open it"
	@echo "  make mlflow           — start local MLflow UI on http://localhost:5001"
	@echo ""
	@echo "Infra changes:"
	@echo "  make tf-plan          — preview Terraform changes"
	@echo "  make tf-destroy       — tear down all AWS resources"
	@echo ""

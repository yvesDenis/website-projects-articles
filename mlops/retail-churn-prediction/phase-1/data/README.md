# Data

Download the Telco Customer Churn dataset from Kaggle and place `WA_Fn-UseC_-Telco-Customer-Churn.csv` here, renamed to `telco_churn.csv`.

```bash
# Option 1 — Kaggle CLI
kaggle datasets download -d blastchar/telco-customer-churn --unzip -p .
mv "WA_Fn-UseC_-Telco-Customer-Churn.csv" telco_churn.csv

# Option 2 — Manual
# https://www.kaggle.com/datasets/blastchar/telco-customer-churn
# Download, unzip, rename to telco_churn.csv, place here.
```

The CSV file is gitignored. Only `telco_churn.csv.dvc` is committed (added in Step 3).

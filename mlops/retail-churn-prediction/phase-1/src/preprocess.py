import pandas as pd
from sklearn.compose import ColumnTransformer
from sklearn.preprocessing import OneHotEncoder, StandardScaler

# Explicit column lists so both train.py and the API agree on the schema
NUM_COLS = ['tenure', 'MonthlyCharges', 'TotalCharges']
CAT_COLS = [
    'gender', 'Partner', 'Dependents', 'PhoneService', 'MultipleLines',
    'InternetService', 'OnlineSecurity', 'OnlineBackup', 'DeviceProtection',
    'TechSupport', 'StreamingTV', 'StreamingMovies', 'Contract',
    'PaperlessBilling', 'PaymentMethod',
]


def clean(df: pd.DataFrame) -> pd.DataFrame:
    """
    Cleaning steps that do NOT cause data leakage — safe to run before the
    train/test split because they only fix known structural issues in the CSV
    (blank strings, a non-feature ID column, binary target encoding).
    """
    df = df.copy()
    df = df.drop(columns=['customerID'])
    # TotalCharges is stored as a string and has blanks for tenure=0 rows
    df['TotalCharges'] = pd.to_numeric(df['TotalCharges'], errors='coerce').fillna(0.0)
    df['Churn'] = (df['Churn'] == 'Yes').astype(int)
    return df


def build_preprocessor() -> ColumnTransformer:
    """
    Return an *unfitted* preprocessor pipeline.
    Call .fit_transform(X_train) on training data only, then .transform(X_test).
    This prevents the scaler from seeing test-set statistics during training.
    """
    return ColumnTransformer(
        transformers=[
            ('num', StandardScaler(), NUM_COLS),
            ('cat', OneHotEncoder(handle_unknown='ignore', sparse_output=False), CAT_COLS),
        ],
        remainder='drop',
    )

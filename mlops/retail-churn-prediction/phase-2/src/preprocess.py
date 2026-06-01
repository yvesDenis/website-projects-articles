import pandas as pd
from sklearn.compose import ColumnTransformer
from sklearn.preprocessing import OneHotEncoder, StandardScaler

NUM_COLS = ['tenure', 'MonthlyCharges', 'TotalCharges']
CAT_COLS = [
    'gender', 'Partner', 'Dependents', 'PhoneService', 'MultipleLines',
    'InternetService', 'OnlineSecurity', 'OnlineBackup', 'DeviceProtection',
    'TechSupport', 'StreamingTV', 'StreamingMovies', 'Contract',
    'PaperlessBilling', 'PaymentMethod',
]


def clean(df: pd.DataFrame) -> pd.DataFrame:
    df = df.copy()
    df = df.drop(columns=['customerID'])
    df['TotalCharges'] = pd.to_numeric(df['TotalCharges'], errors='coerce').fillna(0.0)
    df['Churn'] = (df['Churn'] == 'Yes').astype(int)
    return df


def build_preprocessor() -> ColumnTransformer:
    return ColumnTransformer(
        transformers=[
            ('num', StandardScaler(), NUM_COLS),
            ('cat', OneHotEncoder(handle_unknown='ignore', sparse_output=False), CAT_COLS),
        ],
        remainder='drop',
    )

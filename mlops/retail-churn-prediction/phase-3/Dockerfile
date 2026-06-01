FROM python:3.12-slim

WORKDIR /app

# Install API dependencies first so this layer is cached across model-only rebuilds
COPY requirements-api.txt .
RUN pip install --no-cache-dir -r requirements-api.txt

# Copy source code and baked-in model artifacts.
# In CI, `dvc pull` downloads models/ before `docker build` runs,
# so the trained model is already in the build context.
COPY src/ src/
COPY models/ models/

# Allow `from preprocess import ...` from anywhere inside src/
ENV PYTHONPATH=/app/src

EXPOSE 8000
CMD ["uvicorn", "src.api.main:app", "--host", "0.0.0.0", "--port", "8000"]

FROM python:3.12-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY src/ src/
COPY models/ models/

# Let Python find preprocess.py and evaluate.py as top-level imports inside src/
ENV PYTHONPATH=/app/src

CMD ["uvicorn", "src.api.main:app", "--host", "0.0.0.0", "--port", "8000"]

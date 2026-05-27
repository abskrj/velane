FROM python:3.12-slim
WORKDIR /app
RUN pip install fastapi uvicorn pydantic --no-cache-dir
COPY runner.py .
EXPOSE 8080
CMD ["uvicorn", "runner:app", "--host", "0.0.0.0", "--port", "8080"]

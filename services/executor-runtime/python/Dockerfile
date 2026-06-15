FROM python:3.12-slim
WORKDIR /app
RUN pip install --no-cache-dir \
    fastapi \
    uvicorn \
    pydantic \
    httpx \
    langgraph \
    langchain-core \
    langchain-openai
COPY runner.py cgroup_limits.py .
EXPOSE 8080
CMD ["uvicorn", "runner:app", "--host", "0.0.0.0", "--port", "8080"]

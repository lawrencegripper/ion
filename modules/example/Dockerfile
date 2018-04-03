FROM python:3

WORKDIR /app/
COPY requirements.txt .
RUN pip install -r requirements.txt
COPY src/example.py .

CMD ["python", "example.py"]
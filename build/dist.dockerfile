FROM python:3.12

WORKDIR /build

RUN python -m venv /venv

# Activate the virtual environment
ENV PATH="/venv/bin:$PATH"

# Install any dependencies
COPY .env requirements.txt manage.py .
RUN pip install -r requirements.txt
COPY app app
RUN DEBUG=True python3 manage.py collectstatic

RUN mkdir -p /dist/venv

RUN rm -r /venv/lib/python3.12/site-packages/pip
RUN find /venv/lib/python3.12/site-packages -type d -name "*.dist-info" -exec rm -r {} +

RUN mv /venv/lib/python3.12/site-packages /dist/venv
RUN mv /venv/bin/gunicorn /dist/gunicorn

# Copy and clean app
RUN mv app /dist/app
RUN find /dist -type d -name "__pycache__" -exec rm -r {} +

RUN mv static /dist/static

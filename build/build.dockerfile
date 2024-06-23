FROM node:20.14 as assetbuild
WORKDIR /build
COPY package.json yarn.lock /build
RUN yarn install
COPY postcss.config.js tailwind.config.js vite.config.js /build
COPY assets /build/assets
# Required for tailwindcss to properly purge
COPY app/templates /build/app/templates
RUN yarn build
RUN cp -r /build/assets/icons static/icons


FROM python:3.12 as pybuild
WORKDIR /build
RUN python -m venv /venv
RUN mkdir -p /dist/venv/bin
# Activate the virtual environment
ENV PATH="/venv/bin:$PATH"
# Install any dependencies
COPY requirements.txt manage.py .
RUN pip install -r requirements.txt
COPY .env.example .env
COPY app app
RUN DEBUG=True python3 manage.py collectstatic

RUN rm -r /venv/lib/python3.12/site-packages/pip
RUN find /venv/lib/python3.12/site-packages -type d -name "*.dist-info" -exec rm -r {} +

RUN mv /venv/lib/python3.12/site-packages /dist/venv
RUN mv /venv/bin/gunicorn /dist/venv/bin/gunicorn

# Copy and clean app
RUN mv app /dist/app
RUN find /dist -type d -name "__pycache__" -exec rm -r {} +
COPY .env.example build/manage.py /dist

FROM python:3.12 as final
COPY --from=pybuild /dist /dist
COPY --from=assetbuild /build/static /dist/static
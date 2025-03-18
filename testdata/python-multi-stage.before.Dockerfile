# From https://luis-sena.medium.com/creating-the-perfect-python-dockerfile-51bdec41f1c8
# This Dockerfile demonstrates a multi-stage Python build with virtual environment

# using ubuntu LTS version
FROM ubuntu:20.04 AS builder-image

RUN apt-get update && apt-get install --no-install-recommends -y python3.9 python3.9-dev python3.9-venv python3-pip python3-wheel build-essential && \
   apt-get clean && rm -rf /var/lib/apt/lists/*

# create and activate virtual environment
RUN python3.9 -m venv /opt/venv
ENV PATH="/opt/venv/bin:$PATH"

# install requirements
COPY requirements.txt .
RUN pip3 install --no-cache-dir -r requirements.txt

FROM ubuntu:20.04 AS runner-image
RUN apt-get update && apt-get install --no-install-recommends -y python3.9 python3-venv && \
   apt-get clean && rm -rf /var/lib/apt/lists/*

COPY --from=builder-image /opt/venv /opt/venv

# activate virtual environment
ENV VIRTUAL_ENV=/opt/venv
ENV PATH="/opt/venv/bin:$PATH"

# Add non-root user
RUN useradd --create-home appuser
USER appuser
WORKDIR /home/appuser

# Copy application code
COPY --chown=appuser:appuser . .

CMD ["python", "app.py"] 
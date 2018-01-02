FROM python:3-slim

RUN pip --no-cache-dir install h5py

ADD ./pytest.py /
RUN chmod +x /pytest.py

ENTRYPOINT ["python", "/pytest.py"]

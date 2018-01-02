FROM python:3-slim

RUN pip --no-cache-dir install h5py

ADD ./problem_pytest.py /
RUN chmod +x /problem_pytest.py

ENTRYPOINT ["python", "/problem_pytest.py"]

import os
import requests

LOCALRUNNER_ADDR = os.environ["LOCALRUNNER_ADDR"]


def transaction_begin():
    res = requests.post(LOCALRUNNER_ADDR + f"/_internal/transactional/begin")
    res.raise_for_status()
    return res.text


def transaction_commit(tx_id: str):
    res = requests.post(LOCALRUNNER_ADDR + f"/_internal/transactional/commit/{tx_id}")
    res.raise_for_status()


def transaction_rollback(tx_id: str):
    res = requests.post(LOCALRUNNER_ADDR + f"/_internal/transactional/rollback/{tx_id}")
    res.raise_for_status()

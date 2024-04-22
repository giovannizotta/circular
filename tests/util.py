import logging
import random
import string
from pathlib import Path

import pytest

DOWNLOAD_PATH = Path.cwd() / "tests" / "circular"


@pytest.fixture
def get_plugin(directory):
    if DOWNLOAD_PATH.is_file():
        return DOWNLOAD_PATH
    else:
        raise ValueError("No files were found.")


def generate_random_label():
    label_length = 8
    random_label = "".join(
        random.choice(string.ascii_letters) for _ in range(label_length)
    )
    return random_label


def generate_random_number():
    return random.randint(1, 20_000_000_000_000_00_000)


def pay_with_thread(rpc, bolt11):
    LOGGER = logging.getLogger(__name__)
    try:
        rpc.dev_pay(bolt11, dev_use_shadow=False)
    except Exception as e:
        LOGGER.debug(f"holdinvoice: Error paying payment hash:{e}")
        pass

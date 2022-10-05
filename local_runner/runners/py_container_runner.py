try:
    from __ENTRY__ import app
except ImportError:
    from __ENTRY__ import create_app

    app = create_app()

import logging
from werkzeug.serving import run_simple

logging.getLogger("werkzeug").setLevel(logging.ERROR)

run_simple("localhost", __PORT__, app)

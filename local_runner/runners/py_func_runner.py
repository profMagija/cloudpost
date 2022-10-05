import logging
from flask import Flask, request
from main import __ENTRY__ as entry_point
from werkzeug.serving import run_simple
from base64 import b64encode

app = Flask(__name__)


@app.route("/pubsub", methods=["POST"])
def pubsub_handler():
    data = request.get_data()
    res = entry_point(dict(data=b64encode(data).decode()), dict())
    return res or "OK"


@app.route("/event", methods=["POST"])
def event_handler():
    data = request.json

    res = entry_point(data["message"], data.get("context", dict()))
    return res or "OK"


@app.route("/", methods=["GET", "POST", "PUT", "DELETE"])
def http_handler():
    return entry_point(request)


logging.getLogger("werkzeug").setLevel(logging.ERROR)

run_simple("localhost", __PORT__, app)

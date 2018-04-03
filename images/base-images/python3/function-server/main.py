import json
import signal
import sys
import io

import falcon
import waitress

from function import handler


class Health(object):
    @staticmethod
    def on_get(req, res):
        res.body = '{}'


def read_logs(text_io):
    text_io.seek(0)
    return [line.rstrip('\n') for line in text_io]


def exec_function(req, res):
    msg = None
    if req.content_length:
        msg = json.load(req.stream)

    r = None
    err = None

    stderr = io.StringIO()
    old_stderr = sys.stderr

    try:
        sys.stderr = stderr
        r = handler.handle(msg['context'], msg['payload'])
    except Exception as e:
        err = e
    finally:
        sys.stderr = old_stderr
        stderr.flush()

    res.body = json.dumps({'context': {'error': err, 'logs': read_logs(stderr)}, 'payload': r}, ensure_ascii=False)


def signal_handler(signum, frame):
    sys.exit(0)


signal.signal(signal.SIGINT, signal_handler)
signal.signal(signal.SIGTERM, signal_handler)

app = falcon.API()
app.add_route('/healthz', Health())
app.add_sink(exec_function, '/')

waitress.serve(app, threads=1)

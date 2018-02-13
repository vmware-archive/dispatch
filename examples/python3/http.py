import time

import requests

def handle(ctx, payload):
    resp = requests.get("http://example.com")
    return {"status": resp.status_code}

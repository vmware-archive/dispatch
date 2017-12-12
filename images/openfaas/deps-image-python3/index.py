# Copyright (c) Alex Ellis 2017. All rights reserved.
# Licensed under the MIT license. See LICENSE file in the project root for full license information.

import json
import sys
import traceback
from contextlib import redirect_stdout
from function import handler

if(__name__ == "__main__"):
    try:
        st = json.load(sys.stdin)
        with redirect_stdout(sys.stderr):
            result = handler.handle(st["context"], st["input"])
        json.dump(result, sys.stdout)
    except:
        traceback.print_exc(file=sys.stderr)

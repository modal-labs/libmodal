import typing

import modal

app = modal.App("test-support-1-1")


@app.function(timeout=60 * 5)
def identity_with_repr(s: typing.Any) -> typing.Any:
    return s, repr(s)

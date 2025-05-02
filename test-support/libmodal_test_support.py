import modal


app = modal.App("libmodal-test-support")


@app.function(min_containers=1)
def echo_string(s: str) -> str:
    return "output: " + s


@app.function(min_containers=1)
def bytelength(buf: bytes) -> int:
    return len(buf)

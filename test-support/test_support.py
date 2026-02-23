import os
import time
import typing

import modal

app = modal.App("libmodal-test-support")


@app.function(timeout=60 * 5, _experimental_restrict_output=True)
def echo_string(s: str) -> str:
    return "output: " + s


@app.function(timeout=60 * 5, _experimental_restrict_output=True)
def identity_with_repr(s: typing.Any) -> typing.Any:
    return s, repr(s)


@app.function(timeout=60 * 5)
def sleep(t: int) -> None:
    time.sleep(t)


@app.function(timeout=60 * 5)
def bytelength(buf: bytes) -> int:
    return len(buf)


@app.function(timeout=60 * 5, experimental_options={"input_plane_region": "us-west"})
def input_plane(s: str) -> str:
    return "output: " + s


@app.cls(timeout=60 * 5)
class EchoCls:
    @modal.method()
    def echo_string(self, s: str) -> str:
        return "output: " + s


@app.cls(timeout=60 * 5, experimental_options={"input_plane_region": "us-west"})
class EchoClsInputPlane:
    @modal.method()
    def echo_string(self, s: str) -> str:
        return "output: " + s


@app.cls()
class EchoClsParametrized:
    name: str = modal.parameter(default="test")

    @modal.method()
    def echo_parameter(self) -> str:
        return "output: " + self.name

    @modal.method()
    def echo_env_var(self, var_name: str) -> str:
        return f"output: {var_name}='{os.getenv(var_name, '[not set]')}'"


@app.function(image=modal.Image.debian_slim().pip_install("fastapi"))
@modal.fastapi_endpoint()
def web_endpoint_echo(s: str) -> str:
    return "output: " + s

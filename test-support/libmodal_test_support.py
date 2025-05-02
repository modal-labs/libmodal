import modal


app = modal.App("libmodal-test-support")


@app.function(min_containers=1)
def echo_string(s: str) -> str:
    return "output: " + s

@app.function(min_containers=1)
def bytelength(buf: bytes) -> int:
    return len(buf)

@app.cls(min_containers=1)
class EchoCls:
    @modal.method()
    def echo_string(self, s: str) -> str:
        return "output: " + s

@app.cls()
class EchoClsParametrized:
    name: str = modal.parameter(default="test")
 
    @modal.method()
    def echo_parameter(self) -> str:
        return "output: " + self.name

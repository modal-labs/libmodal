# Test support for libmodal

Sign in to a Modal account, which you'll use for running the test programs.

Then deploy the apps in this folder using the Python client:

```bash
modal deploy libmodal_test_support.py
```

This deployed app will be called from tests in each language.

```bash
# JavaScript
cd modal-js && npm test

# Go
cd modal-go && go test -v -count=1 -parallel=10 ./test
```

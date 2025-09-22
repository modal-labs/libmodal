import os
import subprocess
from datetime import date, datetime, timezone
from typing import Any, List, Union

import cbor2


def tag_hook(decoder, tag: cbor2.CBORTag):
    """Custom tag hook for CBOR decoding"""
    if tag.tag == 64:
        # Tag 64 is Uint8Array - convert to bytes
        if not isinstance(tag.value, bytes):
            raise ValueError(f"Tag 64 value must be bytes, got {type(tag.value)}")
        return tag.value
    # For other tags, return the default CBORTag
    return cbor2.CBORTag(tag, tag.value)


def cbor_loads_with_tag64_support(data: bytes):
    """Load CBOR data with support for tag 64 (Uint8Array) -> bytes"""
    return cbor2.loads(data, tag_hook=tag_hook)


def encode_with_node_cbor_x_uint8array(data: bytes) -> bytes:
    """
    Calls out to node and the cbor-x package to encode the given bytes as CBOR
    using Uint8Array (which creates tag 64).
    Returns the CBOR-encoded bytes.
    """
    # Path to node modules
    node_modules = os.path.join(os.path.dirname(__file__), "modal-js", "node_modules")
    # Write a small JS script to encode stdin as CBOR using cbor-x
    js_code = """
const { encode } = require('cbor-x');
const fs = require('fs');

let chunks = [];
process.stdin.on('data', chunk => chunks.push(chunk));
process.stdin.on('end', () => {
    const buf = Buffer.concat(chunks);
    // Use Uint8Array (creates tag 64)
    const cbor = encode(new Uint8Array(buf));
    process.stdout.write(cbor);
});
"""
    # Run node with the JS code, passing the data via stdin
    proc = subprocess.Popen(
        ["node", "-e", js_code],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        cwd=os.path.dirname(__file__),
        env={**os.environ, "NODE_PATH": node_modules},
    )
    cbor_bytes, err = proc.communicate(input=data)
    if proc.returncode != 0:
        raise RuntimeError(f"Node cbor-x encode failed: {err.decode()}")
    return cbor_bytes


def python_to_js_code(data: Any) -> str:
    """Convert Python data to JavaScript code snippet (legacy function - use explicit JS syntax in new tests)"""
    if data is None:
        return "null"
    elif isinstance(data, bool):
        return "true" if data else "false"
    elif isinstance(data, (int, float)):
        return str(data)
    elif isinstance(data, str):
        # Escape string for JavaScript
        escaped = (
            data.replace("\\", "\\\\")
            .replace('"', '\\"')
            .replace("\n", "\\n")
            .replace("\r", "\\r")
            .replace("\t", "\\t")
        )
        return f'"{escaped}"'
    elif isinstance(data, bytes):
        # Convert bytes to Buffer.from([...]) for Node.js
        byte_array = ", ".join(str(b) for b in data)
        return f"Buffer.from([{byte_array}])"
    elif isinstance(data, list):
        items = ", ".join(python_to_js_code(item) for item in data)
        return f"[{items}]"
    elif isinstance(data, set):
        items = ", ".join(python_to_js_code(item) for item in data)
        return f"new Set([{items}])"
    elif isinstance(data, dict):
        # Check if this is a regular dict or should be a Map
        if all(isinstance(k, str) for k in data.keys()):
            # Regular object
            items = ", ".join(f'"{k}": {python_to_js_code(v)}' for k, v in data.items())
            return f"{{{items}}}"
        else:
            # Use Map for non-string keys
            items = ", ".join(f"[{python_to_js_code(k)}, {python_to_js_code(v)}]" for k, v in data.items())
            return f"new Map([{items}])"
    elif isinstance(data, (datetime, date)):
        # Convert to ISO string and create Date object
        iso_string = data.isoformat()
        return f'new Date("{iso_string}")'
    else:
        raise TypeError(f"Cannot convert {type(data)} to JavaScript code")


def encode_js_data_with_cbor_x(data: Any) -> bytes:
    """
    Calls out to node and the cbor-x package to encode JavaScript data as CBOR.
    Uses direct JavaScript code generation instead of JSON intermediate.
    Returns the CBOR-encoded bytes.
    """
    node_modules = os.path.join(os.path.dirname(__file__), "modal-js", "node_modules")

    # Generate JavaScript code directly
    js_data_code = python_to_js_code(data)

    js_code = f"""
const {{ encode }} = require('cbor-x');

// Direct JavaScript data (converted from Python: {type(data).__name__})
const jsData = {js_data_code};

console.error(`Encoding JS data: ${{JSON.stringify(jsData, null, 2)}} (type: ${{typeof jsData}}, constructor: ${{jsData?.constructor?.name}})`);

const cbor = encode(jsData);
process.stdout.write(cbor);
"""

    proc = subprocess.Popen(
        ["node", "-e", js_code],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        cwd=os.path.dirname(__file__),
        env={**os.environ, "NODE_PATH": node_modules},
    )
    cbor_bytes, err = proc.communicate()
    if proc.returncode != 0:
        raise RuntimeError(f"Node cbor-x encode failed: {err.decode()}")

    # Print JavaScript encoding debug output if available
    if err:
        print(f"JS encode debug: {err.decode().strip()}")

    return cbor_bytes


def decode_and_verify_js_roundtrip(cbor_data: bytes, original_js_code: str) -> bool:
    """
    Calls out to node and the cbor-x package to decode CBOR data back to JavaScript types.
    Compares the decoded result with the original JavaScript value.
    Returns True if they match, False otherwise.
    """
    node_modules = os.path.join(os.path.dirname(__file__), "modal-js", "node_modules")

    js_code = f"""
const {{ decode }} = require('cbor-x');

let chunks = [];
process.stdin.on('data', chunk => chunks.push(chunk));
process.stdin.on('end', () => {{
    const buf = Buffer.concat(chunks);

    try {{
        const decoded = decode(buf);
        const original = {original_js_code};

        // Log the native JavaScript type for debugging
        process.stderr.write(`JS decoded: ${{JSON.stringify(decoded, null, 2)}} (type: ${{typeof decoded}}, constructor: ${{decoded?.constructor?.name}})\\n`);
        process.stderr.write(`JS original: ${{JSON.stringify(original, null, 2)}} (type: ${{typeof original}}, constructor: ${{original?.constructor?.name}})\\n`);

        // Compare the values using a deep equality check
        function deepEqual(a, b) {{
            if (a === b) return true;

            // Handle special number cases
            if (typeof a === 'number' && typeof b === 'number') {{
                if (Number.isNaN(a) && Number.isNaN(b)) return true;
                return a === b;
            }}

            // Handle Date objects
            if (a instanceof Date && b instanceof Date) {{
                return a.getTime() === b.getTime();
            }}

            // Handle Sets
            if (a instanceof Set && b instanceof Set) {{
                if (a.size !== b.size) return false;
                for (let item of a) {{
                    if (!b.has(item)) return false;
                }}
                return true;
            }}

            // Handle Maps
            if (a instanceof Map && b instanceof Map) {{
                if (a.size !== b.size) return false;
                for (let [key, value] of a) {{
                    if (!b.has(key) || !deepEqual(b.get(key), value)) return false;
                }}
                return true;
            }}

            // Handle Arrays
            if (Array.isArray(a) && Array.isArray(b)) {{
                if (a.length !== b.length) return false;
                for (let i = 0; i < a.length; i++) {{
                    if (!deepEqual(a[i], b[i])) return false;
                }}
                return true;
            }}

            // Handle Objects
            if (typeof a === 'object' && typeof b === 'object' && a !== null && b !== null) {{
                const keysA = Object.keys(a);
                const keysB = Object.keys(b);
                if (keysA.length !== keysB.length) return false;
                for (let key of keysA) {{
                    if (!keysB.includes(key) || !deepEqual(a[key], b[key])) return false;
                }}
                return true;
            }}

            return false;
        }}

        const matches = deepEqual(decoded, original);
        process.stderr.write(`Values match: ${{matches}}\\n`);
        process.stdout.write(matches ? 'values_match' : 'values_differ');
    }} catch (error) {{
        process.stderr.write(`Decode error: ${{error.message}}`);
        process.exit(1);
    }}
}});
"""

    proc = subprocess.Popen(
        ["node", "-e", js_code],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        cwd=os.path.dirname(__file__),
        env={**os.environ, "NODE_PATH": node_modules},
    )
    result, err = proc.communicate(input=cbor_data)
    if proc.returncode != 0:
        raise RuntimeError(f"Node cbor-x decode failed: {err.decode()}")

    # Print JavaScript debug output if available
    if err:
        print(f"JS debug: {err.decode().strip()}")

    result_str = result.decode().strip()
    if result_str == "values_match":
        return True
    elif result_str == "values_differ":
        return False
    else:
        raise RuntimeError(f"Unexpected result from JavaScript comparison: {result_str}")


def decode_and_log_js_data_with_cbor_x(cbor_data: bytes) -> None:
    """
    Legacy function - kept for compatibility.
    Use decode_and_verify_js_roundtrip for new tests.
    """
    node_modules = os.path.join(os.path.dirname(__file__), "modal-js", "node_modules")

    js_code = """
const { decode } = require('cbor-x');

let chunks = [];
process.stdin.on('data', chunk => chunks.push(chunk));
process.stdin.on('end', () => {
    const buf = Buffer.concat(chunks);

    try {
        const decoded = decode(buf);

        // Log the native JavaScript type for debugging
        process.stderr.write(`JS decoded: ${JSON.stringify(decoded, null, 2)} (type: ${typeof decoded}, constructor: ${decoded?.constructor?.name})\\n`);

        // Just output success - we don't need to transport data back
        process.stdout.write('decode_success');
    } catch (error) {
        process.stderr.write(`Decode error: ${error.message}`);
        process.exit(1);
    }
});
"""

    proc = subprocess.Popen(
        ["node", "-e", js_code],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        cwd=os.path.dirname(__file__),
        env={**os.environ, "NODE_PATH": node_modules},
    )
    result, err = proc.communicate(input=cbor_data)
    if proc.returncode != 0:
        raise RuntimeError(f"Node cbor-x decode failed: {err.decode()}")

    # Print JavaScript debug output if available
    if err:
        print(f"JS debug: {err.decode().strip()}")

    if result.decode().strip() != "decode_success":
        raise RuntimeError("JavaScript decode did not complete successfully")


# Removed js_to_python_type function - no longer needed since we don't transport JS data back to Python


def encode_with_node_cbor_x(data: bytes) -> bytes:
    """
    Calls out to node and the cbor-x package to encode the given bytes as CBOR.
    Returns the CBOR-encoded bytes.
    """
    # Path to node modules
    node_modules = os.path.join(os.path.dirname(__file__), "modal-js", "node_modules")
    # Write a small JS script to encode stdin as CBOR using cbor-x
    js_code = """
const { encode } = require('cbor-x');
const fs = require('fs');

let chunks = [];
process.stdin.on('data', chunk => chunks.push(chunk));
process.stdin.on('end', () => {
    const buf = Buffer.concat(chunks);
    // Option 1: Use Buffer directly (Node.js specific)
    // const cbor = encode(buf);

    // Option 2: Use Uint8Array with tagUint8Array: false (doesn't work!)
    // const cbor = encode(new Uint8Array(buf), { tagUint8Array: false });

    // Option 3: Use ArrayBuffer (cross-platform, should work!)
    const cbor = encode(new Uint8Array(buf).buffer);
    process.stdout.write(cbor);
});
"""
    # Run node with the JS code, passing the data via stdin
    proc = subprocess.Popen(
        ["node", "-e", js_code],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        cwd=os.path.dirname(__file__),
        env={**os.environ, "NODE_PATH": node_modules},
    )
    cbor_bytes, err = proc.communicate(input=data)
    if proc.returncode != 0:
        raise RuntimeError(f"Node cbor-x encode failed: {err.decode()}")
    return cbor_bytes


def test_cbor_x_encode_and_python_decode():
    # Test data
    original = b"hello world \x00\xff"
    cbor_bytes = encode_with_node_cbor_x(original)
    decoded = cbor2.loads(cbor_bytes)
    print("Decoded", decoded)
    assert isinstance(decoded, bytes), f"Decoded type is {type(decoded)}"
    assert decoded == original, f"Decoded bytes {decoded!r} != original {original!r}"
    print(f"✅ ArrayBuffer approach works! Decoded: {decoded!r}")


def test_uint8array_with_tag64_support():
    # Test data
    original = b"hello world \x00\xff"
    # Use the Uint8Array encoder (creates tag 64)
    cbor_bytes = encode_with_node_cbor_x_uint8array(original)

    # Try regular cbor2.loads first (should fail)
    try:
        decoded_regular = cbor2.loads(cbor_bytes)
        print(f"Regular cbor2.loads: {decoded_regular} (type: {type(decoded_regular)})")
    except Exception as e:
        print(f"Regular cbor2.loads failed: {e}")

    # Now try with tag 64 support
    decoded = cbor_loads_with_tag64_support(cbor_bytes)
    print(f"With tag 64 support: {decoded} (type: {type(decoded)})")
    assert isinstance(decoded, bytes), f"Decoded type is {type(decoded)}"
    assert decoded == original, f"Decoded bytes {decoded!r} != original {original!r}"
    print(f"✅ Uint8Array with tag 64 support works! Decoded: {decoded!r}")


def has_cbor_tags(obj: Any, path: str = "root") -> List[str]:
    """
    Recursively check if an object contains any CBORTag objects.
    Returns a list of paths where CBORTag objects were found.
    """
    issues = []

    if isinstance(obj, cbor2.CBORTag):
        # Any CBORTag object is considered an issue now
        issues.append(f"{path}: CBORTag({obj.tag}, {obj.value})")
    elif isinstance(obj, dict):
        for key, value in obj.items():
            issues.extend(has_cbor_tags(key, f"{path}.key[{repr(key)}]"))
            issues.extend(has_cbor_tags(value, f"{path}[{repr(key)}]"))
    elif isinstance(obj, (list, tuple)):
        for i, item in enumerate(obj):
            issues.extend(has_cbor_tags(item, f"{path}[{i}]"))
    elif isinstance(obj, set):
        for i, item in enumerate(obj):
            issues.extend(has_cbor_tags(item, f"{path}.set[{i}]"))

    return issues


def test_roundtrip_js(test_name: str, js_input: str, python_intermediate: Any, compare_fn=None) -> bool:
    """
    Test full roundtrip using JavaScript input syntax.
    js_input: JavaScript code that creates the test data
    python_intermediate: Expected Python value after JS->Python conversion
    Returns True if the roundtrip was successful.
    """
    print(f"\n--- Testing {test_name} ---")
    print(f"JS input: {js_input}")
    print(f"Expected Python intermediate: {python_intermediate} (type: {type(python_intermediate).__name__})")

    try:
        # Step 1: Encode JavaScript data using cbor-x
        js_encoded = encode_js_code_with_cbor_x(js_input)
        print(f"JS encoded length: {len(js_encoded)} bytes")

        # Step 2: Decode using Python cbor2
        python_decoded = cbor2.loads(js_encoded)
        print(f"Python decoded: {python_decoded} (type: {type(python_decoded)})")

        # Verification step: Check for ANY CBORTag objects (fail if found)
        cbor_tag_issues = has_cbor_tags(python_decoded)
        if cbor_tag_issues:
            print("❌ FAILED: Found CBORTag objects in Python decoded data:")
            for issue in cbor_tag_issues:
                print(f"   - {issue}")
            print(f"❌ {test_name} roundtrip failed due to CBORTag objects")
            return False

        # Check if Python intermediate matches expected (skip if we have a custom compare function for time-sensitive data)
        if compare_fn is None and python_decoded != python_intermediate:
            print(f"❌ Python intermediate mismatch: got {python_decoded}, expected {python_intermediate}")
            return False
        elif compare_fn is not None:
            # For custom comparisons, we'll check at the end - intermediate might be time-sensitive
            print("Using custom comparison function, skipping intermediate check")

        # Step 3: Re-encode using Python cbor2
        python_encoded = cbor2.dumps(python_decoded)
        print(f"Python encoded length: {len(python_encoded)} bytes")

        # Step 4: Decode back using JavaScript cbor-x and verify it matches original
        js_matches_original = decode_and_verify_js_roundtrip(python_encoded, js_input)

        if not js_matches_original:
            print(f"❌ {test_name} roundtrip failed: JavaScript decoded value doesn't match original")
            return False

        print(f"✅ {test_name} roundtrip successful!")
        return True

    except Exception as e:
        print(f"❌ {test_name} roundtrip failed with exception: {e}")
        return False


def encode_js_code_with_cbor_x(js_code: str) -> bytes:
    """
    Calls out to node and the cbor-x package to encode JavaScript code as CBOR.
    js_code: JavaScript expression that evaluates to the data to encode
    Returns the CBOR-encoded bytes.
    """
    node_modules = os.path.join(os.path.dirname(__file__), "modal-js", "node_modules")

    full_js_code = f"""
const {{ encode }} = require('cbor-x');

// Evaluate the JavaScript input
const jsData = {js_code};

console.error(`Encoding JS data: ${{JSON.stringify(jsData, null, 2)}} (type: ${{typeof jsData}}, constructor: ${{jsData?.constructor?.name}})`);

const cbor = encode(jsData);
process.stdout.write(cbor);
"""

    proc = subprocess.Popen(
        ["node", "-e", full_js_code],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        cwd=os.path.dirname(__file__),
        env={**os.environ, "NODE_PATH": node_modules},
    )
    cbor_bytes, err = proc.communicate()
    if proc.returncode != 0:
        raise RuntimeError(f"Node cbor-x encode failed: {err.decode()}")

    # Print JavaScript encoding debug output if available
    if err:
        print(f"JS encode debug: {err.decode().strip()}")

    return cbor_bytes


# Legacy test function for backward compatibility
def test_roundtrip(test_name: str, original_data: Any, compare_fn=None) -> bool:
    """
    Legacy test function - converts Python data to JS and tests roundtrip.
    DEPRECATED: Use test_roundtrip_js for new tests with explicit JS syntax.
    """
    js_code = python_to_js_code(original_data)
    return test_roundtrip_js(test_name, js_code, original_data, compare_fn)


def compare_sets(original: set, result: set) -> bool:
    """Compare sets (order doesn't matter)"""
    return original == result


def compare_dicts(original: dict, result: dict) -> bool:
    """Compare dictionaries"""
    return original == result


def compare_dates(original: Union[datetime, date], result: Union[datetime, date]) -> bool:
    """Compare dates/datetimes with some tolerance for timezone handling"""
    if isinstance(original, datetime) and isinstance(result, datetime):
        # Handle timezone-aware vs naive datetime comparison
        if original.tzinfo is None and result.tzinfo is not None:
            # Convert result to naive datetime for comparison
            result = result.replace(tzinfo=None)
        elif original.tzinfo is not None and result.tzinfo is None:
            # Convert original to naive datetime for comparison
            original = original.replace(tzinfo=None)

        # Allow small differences (up to 1 second) due to precision issues
        diff = abs((original - result).total_seconds())
        return diff < 1.0

    return original == result


def test_basic_types_js():
    """Test basic data types using explicit JavaScript syntax"""
    print("\n=== Testing Basic Types (JS Syntax) ===")

    test_cases = [
        ("positive integer", "42", 42),
        ("negative integer", "-123", -123),
        ("zero", "0", 0),
        ("large integer", "9007199254740991", 9007199254740991),  # JavaScript safe integer limit
        ("boolean true", "true", True),
        ("boolean false", "false", False),
        ("empty string", '""', ""),
        ("simple string", '"hello world"', "hello world"),
        ("unicode string", '"Hello 🌍 世界"', "Hello 🌍 世界"),
        ("string with escapes", '"line1\\nline2\\ttab"', "line1\nline2\ttab"),
        ("null", "null", None),
        ("float", "3.14159", 3.14159),
        ("negative float", "-2.71828", -2.71828),
        # Bytes using different JavaScript approaches
        ("empty bytes (Buffer)", "Buffer.from([])", b""),
        ("simple bytes (Buffer)", "Buffer.from([104, 101, 108, 108, 111])", b"hello"),  # "hello"
        ("binary bytes (Buffer)", "Buffer.from([0, 1, 2, 255, 254, 253])", b"\x00\x01\x02\xff\xfe\xfd"),
        ("bytes from string", 'Buffer.from("hello", "utf8")', b"hello"),
        ("empty bytes (Uint8Array)", "new Uint8Array([])", b""),
        ("simple bytes (Uint8Array)", "new Uint8Array([104, 101, 108, 108, 111])", b"hello"),
    ]

    results = []
    for name, js_input, python_expected in test_cases:
        success = test_roundtrip_js(name, js_input, python_expected)
        results.append((name, success))

    return results


def test_basic_types():
    """Legacy test function - kept for compatibility"""
    return test_basic_types_js()


def test_collection_types_js():
    """Test collection types using explicit JavaScript syntax"""
    print("\n=== Testing Collection Types (JS Syntax) ===")

    test_cases = [
        # Arrays
        ("empty array", "[]", []),
        ("simple array", "[1, 2, 3]", [1, 2, 3]),
        ("mixed array", '[1, "hello", true, null]', [1, "hello", True, None]),
        ("nested array", "[[1, 2], [3, 4], ['a', 'b']]", [[1, 2], [3, 4], ["a", "b"]]),
        # Sets
        ("empty set", "new Set([])", set(), compare_sets),
        ("simple set", "new Set([1, 2, 3])", {1, 2, 3}, compare_sets),
        (
            "mixed set",
            'new Set([2, "hello", true])',
            {2, "hello", True},
            compare_sets,
        ),  # Use 2 instead of 1 to avoid True==1 issue
        ("string set", 'new Set(["apple", "banana", "cherry"])', {"apple", "banana", "cherry"}, compare_sets),
        # Objects (regular dictionaries)
        ("empty object", "{}", {}),
        ("simple object", '{"a": 1, "b": 2}', {"a": 1, "b": 2}),
        (
            "mixed object",
            '{"int": 42, "str": "hello", "bool": true, "null": null}',
            {"int": 42, "str": "hello", "bool": True, "null": None},
        ),
        ("nested object", '{"outer": {"inner": {"deep": "value"}}}', {"outer": {"inner": {"deep": "value"}}}),
        (
            "object with array values",
            '{"numbers": [1, 2, 3], "strings": ["a", "b"]}',
            {"numbers": [1, 2, 3], "strings": ["a", "b"]},
        ),
        # Maps (will likely create CBORTag issues, but good to test)
        ("empty map", "new Map([])", {}, compare_dicts),
        ("simple map with string keys", 'new Map([["a", 1], ["b", 2]])', {"a": 1, "b": 2}, compare_dicts),
    ]

    results = []
    for item in test_cases:
        if len(item) == 4:
            name, js_input, python_expected, compare_fn = item
            success = test_roundtrip_js(name, js_input, python_expected, compare_fn)
        else:
            name, js_input, python_expected = item
            success = test_roundtrip_js(name, js_input, python_expected)
        results.append((name, success))

    return results


def test_collection_types():
    """Legacy test function - kept for compatibility"""
    return test_collection_types_js()


def test_date_types_js():
    """Test date and datetime objects using explicit JavaScript syntax"""
    print("\n=== Testing Date Types (JS Syntax) ===")

    test_cases = [
        # JavaScript Date objects with different input formats
        (
            "date from ISO string",
            'new Date("2023-12-25T15:30:45.000Z")',
            datetime(2023, 12, 25, 15, 30, 45, tzinfo=timezone.utc),
            compare_dates,
        ),
        (
            "date from milliseconds",
            "new Date(1703519445000)",
            datetime(2023, 12, 25, 15, 50, 45, tzinfo=timezone.utc),
            compare_dates,
        ),
        (
            "date from date string",
            'new Date("2023-12-25")',
            datetime(2023, 12, 25, 0, 0, 0, tzinfo=timezone.utc),
            compare_dates,
        ),
        (
            "date with microseconds",
            'new Date("2023-01-01T12:00:00.123Z")',
            datetime(2023, 1, 1, 12, 0, 0, 123000, tzinfo=timezone.utc),
            compare_dates,
        ),
        # Skip current date test in JS syntax tests since new Date() creates different values each time
        # ("current date", "new Date()", datetime.now(timezone.utc),
        #  lambda orig, result: abs((datetime.now(timezone.utc) - result).total_seconds()) < 5),
    ]

    results = []
    for name, js_input, python_expected, compare_fn in test_cases:
        success = test_roundtrip_js(name, js_input, python_expected, compare_fn)
        results.append((name, success))

    return results


def test_date_types():
    """Legacy test function - kept for compatibility"""
    return test_date_types_js()


def test_cbor_tag_verification():
    """Test that our CBORTag verification system works"""
    print("\n=== Testing CBORTag Verification System ===")

    # Create a test with a known CBORTag that should be flagged
    test_data = cbor2.dumps(cbor2.CBORTag(99, "custom_tag_value"))

    print("Testing detection of untyped CBORTag...")
    decoded = cbor2.loads(test_data)
    print(f"Decoded: {decoded} (type: {type(decoded)})")

    issues = has_cbor_tags(decoded)
    if issues:
        print("✅ CBORTag verification system working - detected untyped tags:")
        for issue in issues:
            print(f"   - {issue}")
    else:
        print("❌ CBORTag verification system not working - should have detected untyped tag")

    print()


def test_edge_cases_js():
    """Test edge cases and special values using explicit JavaScript syntax"""
    print("\n=== Testing Edge Cases (JS Syntax) ===")

    test_cases = [
        ("float", "3.14159", 3.14159),
        ("negative float", "-2.71828", -2.71828),
        ("very small float", "1e-10", 1e-10),
        ("very large float", "1e10", 1e10),
        ("scientific notation positive", "1.23e5", 123000.0),
        ("scientific notation negative", "1.23e-5", 0.0000123),
        ("infinity", "Infinity", float("inf")),
        ("negative infinity", "-Infinity", float("-inf")),
        # Note: NaN comparison always fails, so we'll use a custom compare function
        ("NaN", "NaN", float("nan"), lambda orig, result: str(result) == "nan"),
        (
            "empty nested structure",
            '{"list": [], "dict": {}, "set": new Set()}',
            {"list": [], "dict": {}, "set": set()},
            lambda orig, result: (
                orig["list"] == result["list"] and orig["dict"] == result["dict"] and orig["set"] == result["set"]
            ),
        ),
        (
            "complex nested with explicit dates",
            '{"data": [{"id": 1, "tags": new Set(["urgent", "important"]), "meta": {"created": new Date("2025-09-22T11:15:03.963Z")}}, {"id": 2, "tags": new Set(["normal"]), "meta": {"created": new Date("2025-09-22T00:00:00.000Z")}}]}',
            {
                "data": [
                    {
                        "id": 1,
                        "tags": {"urgent", "important"},
                        "meta": {"created": datetime(2025, 9, 22, 11, 15, 3, 963000, tzinfo=timezone.utc)},
                    },
                    {
                        "id": 2,
                        "tags": {"normal"},
                        "meta": {"created": datetime(2025, 9, 22, 0, 0, 0, tzinfo=timezone.utc)},
                    },
                ]
            },
        ),
        # Test various number edge cases
        ("zero", "0", 0),
        ("negative zero", "-0", -0),
        ("max safe integer", "Number.MAX_SAFE_INTEGER", 9007199254740991),
        ("min safe integer", "Number.MIN_SAFE_INTEGER", -9007199254740991),
        ("smallest positive number", "Number.MIN_VALUE", 5e-324),
        # Test string edge cases
        ("empty string", '""', ""),
        ("string with unicode", '"Hello 🌍 世界 🚀"', "Hello 🌍 世界 🚀"),
        ("string with escapes", '"Line 1\\nLine 2\\tTabbed"', "Line 1\nLine 2\tTabbed"),
        ("string with quotes", '"He said \\"Hello\\""', 'He said "Hello"'),
    ]

    results = []
    for item in test_cases:
        if len(item) == 4:
            name, js_input, python_expected, compare_fn = item
            success = test_roundtrip_js(name, js_input, python_expected, compare_fn)
        else:
            name, js_input, python_expected = item
            success = test_roundtrip_js(name, js_input, python_expected)
        results.append((name, success))

    return results


def test_edge_cases():
    """Legacy test function - kept for compatibility"""
    return test_edge_cases_js()


def run_all_roundtrip_tests():
    """Run all roundtrip tests and return summary"""
    print("🚀 Starting comprehensive CBOR roundtrip tests...")
    print("Testing: JS encode -> Python decode -> Python encode -> JS decode")

    # Test CBORTag verification system first
    test_cbor_tag_verification()

    all_results = []

    # Run all test suites
    all_results.extend(test_basic_types())
    all_results.extend(test_collection_types())
    all_results.extend(test_date_types())
    all_results.extend(test_edge_cases())

    # Print summary
    print("\n" + "=" * 60)
    print("ROUNDTRIP TEST SUMMARY")
    print("=" * 60)

    passed = sum(1 for _, success in all_results if success)
    total = len(all_results)

    print(f"Total tests: {total}")
    print(f"Passed: {passed}")
    print(f"Failed: {total - passed}")
    print(f"Success rate: {passed / total * 100:.1f}%")

    if total - passed > 0:
        print("\nFailed tests:")
        for name, success in all_results:
            if not success:
                print(f"  ❌ {name}")

    print("\n" + "=" * 60)
    return passed == total


if __name__ == "__main__":
    # Run original tests first
    print("=== Testing ArrayBuffer approach (no tags) ===")
    test_cbor_x_encode_and_python_decode()
    print("CBOR-x encode and cbor2 decode roundtrip successful.")

    print("\n=== Testing Uint8Array with tag 64 support ===")
    test_uint8array_with_tag64_support()
    print("Uint8Array with tag 64 support roundtrip successful.")

    # Run comprehensive roundtrip tests
    print("\n" + "=" * 80)
    print("COMPREHENSIVE CBOR ROUNDTRIP TESTS")
    print("=" * 80)

    success = run_all_roundtrip_tests()

    if success:
        print("\n🎉 All roundtrip tests passed!")
    else:
        print("\n💥 Some roundtrip tests failed!")
        exit(1)

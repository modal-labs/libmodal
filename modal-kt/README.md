# Modal Kotlin SDK

The [Modal](https://modal.com/) Kotlin SDK provides coroutine-first access to Modal from Kotlin/JVM projects. Use it to call deployed Functions and Cls methods, create Sandboxes, work with Images, Queues, Secrets, Volumes, and related Modal resources.

This is the current in-repo Kotlin SDK implementation. Core Function, Cls, Queue, Image, and Sandbox workflows are available. Some parity work and cloud-backed verification are still ongoing.

## Installation

`modal-kt` is not published as a Maven artifact in this repo yet. The current way to use it is as a local Gradle module or composite build.

Requirements:
- Java 21
- Kotlin/JVM with coroutines

Example `settings.gradle.kts` using a local checkout:

```kotlin
includeBuild("../libmodal/modal-kt")
```

Example dependency:

```kotlin
dependencies {
    implementation("com.modal:modal-kt")
}
```

If you are working inside this repo directly, use the existing `modal-kt` module.

## Authentication And Configuration

Authenticate the same way as the other libmodal SDKs.

```bash
export MODAL_TOKEN_ID=ak-NOTAREALTOKENSTRINGXYZ
export MODAL_TOKEN_SECRET=as-FAKESECRETSTRINGABCDEF
```

Configuration is loaded from `~/.modal.toml` by default. You can override it with:
- `MODAL_CONFIG_PATH`
- `MODAL_PROFILE`
- `MODAL_TOKEN_ID`
- `MODAL_TOKEN_SECRET`
- `MODAL_SERVER_URL`
- `MODAL_ENVIRONMENT`

Environment variables override values from `.modal.toml`.

## Using The API

Create a client:

```kotlin
val modal = ModalClient()
```

Look up an app:

```kotlin
val app = modal.apps.fromName(
    "libmodal-example",
    AppFromNameParams(createIfMissing = true),
)
```

Call a deployed Function:

```kotlin
val function = modal.functions.fromName("libmodal-test-support", "echo_string")
val result = function.remote(kwargs = mapOf("s" to "hello"))
println(result)
```

Spawn a FunctionCall:

```kotlin
val function = modal.functions.fromName("libmodal-test-support", "echo_string")
val call = function.spawn(kwargs = mapOf("s" to "hello"))
println(call.get())
```

Call a deployed Cls:

```kotlin
val cls = modal.cls.fromName("libmodal-test-support", "EchoCls")
val instance = cls.instance()
val method = instance.method("echo_string")
println(method.remote(kwargs = mapOf("s" to "hello")))
```

Create and use a Sandbox:

```kotlin
val image = modal.images.fromRegistry("alpine:3.21")
val sandbox = modal.sandboxes.create(
    app,
    image,
    SandboxCreateParams(command = listOf("cat")),
)

sandbox.stdin.writeText("hello from kotlin")
sandbox.stdin.close()
println(sandbox.stdout.readText())
sandbox.terminate()
```

Run commands in a Sandbox:

```kotlin
val process = sandbox.exec(listOf("echo", "hello"))
println(process.stdout.readText())
println(process.wait())
```

Read and write sandbox files:

```kotlin
val writeHandle = sandbox.open("/tmp/example.txt", "w")
writeHandle.write("hello\n".toByteArray())
writeHandle.close()

val readHandle = sandbox.open("/tmp/example.txt", "r")
println(String(readHandle.read()))
readHandle.close()
```

Create filesystem and directory snapshots:

```kotlin
val snapshotImage = sandbox.snapshotFilesystem()

val repoSnapshot = sandbox.snapshotDirectory("/repo")
sandbox.mountImage("/repo", repoSnapshot)
```

Work with Queues:

```kotlin
val queue = modal.queues.ephemeral()
queue.put(123)
println(queue.get())
queue.closeEphemeral()
```

Add custom gRPC interceptors:

```kotlin
val modal = ModalClient(
    ModalClientParams(
        grpcInterceptors = listOf(myInterceptor),
    ),
)
```

## Current Coverage

Supported surface today includes:
- Functions and FunctionCalls
- Cls lookup, method access, and runtime option binding
- Sandboxes: lifecycle, logs, exec, files, snapshots, tunnels, connect tokens, tags
- Images, Secrets, Volumes, Proxies, CloudBucketMounts, Queues
- Blob-backed large function payload handling
- Custom gRPC interceptors on the client

The Kotlin SDK does not claim full 1:1 parity with `modal-js` yet.

## Examples

Implemented examples currently cover:
- Functions: call, spawn, current stats
- Cls: call, call with options
- Sandboxes: basic lifecycle, exec, poll, files, filesystem snapshot, directory snapshot, connect tokens, tunnels, GPU, named sandboxes, private images, proxies, secrets, cloud buckets, volumes, prewarm, coding-agent flow
- Images: build from Dockerfile commands
- Client setup: custom credentials, telemetry interceptor

Example sources live in [`src/examples/kotlin`](/Users/slavko/libmodal/modal-kt/src/examples/kotlin).

To compile the example set:

```bash
cd modal-kt
./gradlew compileExamples
```

## Development

Useful commands:

```bash
cd modal-kt
./gradlew test
./gradlew compileExamples
```

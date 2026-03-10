package modal.task_command_router

import modal.client.*
import com.google.protobuf.ByteString as ProtoByteString
import okio.ByteString
import okio.ByteString.Companion.toByteString

@Suppress(
    "unused",
    "FunctionName",
)
fun TaskExecStartRequest.Companion.newBuilder(): TaskExecStartRequest.Builder = TaskExecStartRequest.Builder()
fun TaskExecStartRequest.Companion.getDefaultInstance(): TaskExecStartRequest = TaskExecStartRequest()
fun TaskExecStartRequest.toByteArray(): ByteArray = encode()
val TaskExecStartRequest.taskId: String get() = task_id
fun TaskExecStartRequest.Builder.setTaskId(value: String): TaskExecStartRequest.Builder = apply { task_id = value }
var TaskExecStartRequest.Builder.taskId: String
  get() = task_id
  set(value) { task_id = value }
val TaskExecStartRequest.execId: String get() = exec_id
fun TaskExecStartRequest.Builder.setExecId(value: String): TaskExecStartRequest.Builder = apply { exec_id = value }
var TaskExecStartRequest.Builder.execId: String
  get() = exec_id
  set(value) { exec_id = value }
val TaskExecStartRequest.stdoutConfig: TaskExecStdoutConfig get() = stdout_config
fun TaskExecStartRequest.Builder.setStdoutConfig(value: TaskExecStdoutConfig): TaskExecStartRequest.Builder = apply { stdout_config = value }
var TaskExecStartRequest.Builder.stdoutConfig: TaskExecStdoutConfig
  get() = stdout_config
  set(value) { stdout_config = value }
val TaskExecStartRequest.stderrConfig: TaskExecStderrConfig get() = stderr_config
fun TaskExecStartRequest.Builder.setStderrConfig(value: TaskExecStderrConfig): TaskExecStartRequest.Builder = apply { stderr_config = value }
var TaskExecStartRequest.Builder.stderrConfig: TaskExecStderrConfig
  get() = stderr_config
  set(value) { stderr_config = value }
val TaskExecStartRequest.timeoutSecs: Int? get() = timeout_secs
fun TaskExecStartRequest.hasTimeoutSecs(): Boolean = timeout_secs != null
fun TaskExecStartRequest.Builder.setTimeoutSecs(value: Int?): TaskExecStartRequest.Builder = apply { timeout_secs = value }
var TaskExecStartRequest.Builder.timeoutSecs: Int?
  get() = timeout_secs
  set(value) { timeout_secs = value }
fun TaskExecStartRequest.hasWorkdir(): Boolean = workdir != null
fun TaskExecStartRequest.Builder.setWorkdir(value: String?): TaskExecStartRequest.Builder = apply { workdir = value }
var TaskExecStartRequest.Builder.workdir: String?
  get() = workdir
  set(value) { workdir = value }
val TaskExecStartRequest.ptyInfo: PTYInfo? get() = pty_info
fun TaskExecStartRequest.hasPtyInfo(): Boolean = pty_info != null
fun TaskExecStartRequest.Builder.setPtyInfo(value: PTYInfo?): TaskExecStartRequest.Builder = apply { pty_info = value }
var TaskExecStartRequest.Builder.ptyInfo: PTYInfo?
  get() = pty_info
  set(value) { pty_info = value }
val TaskExecStartRequest.runtimeDebug: Boolean get() = runtime_debug
fun TaskExecStartRequest.Builder.setRuntimeDebug(value: Boolean): TaskExecStartRequest.Builder = apply { runtime_debug = value }
var TaskExecStartRequest.Builder.runtimeDebug: Boolean
  get() = runtime_debug
  set(value) { runtime_debug = value }
val TaskExecStartRequest.commandArgs: List<String> get() = command_args
val TaskExecStartRequest.commandArgsList: List<String> get() = command_args
val TaskExecStartRequest.commandArgsCount: Int get() = command_args.size
fun TaskExecStartRequest.Builder.addCommandArgs(value: String): TaskExecStartRequest.Builder = apply { command_args = command_args + value }
fun TaskExecStartRequest.Builder.addAllCommandArgs(items: Iterable<String>): TaskExecStartRequest.Builder = apply { this.command_args = this.command_args + items }
fun TaskExecStartRequest.Builder.clearCommandArgs(): TaskExecStartRequest.Builder = apply { command_args = emptyList() }
val TaskExecStartRequest.secretIds: List<String> get() = secret_ids
val TaskExecStartRequest.secretIdsList: List<String> get() = secret_ids
val TaskExecStartRequest.secretIdsCount: Int get() = secret_ids.size
fun TaskExecStartRequest.Builder.addSecretIds(value: String): TaskExecStartRequest.Builder = apply { secret_ids = secret_ids + value }
fun TaskExecStartRequest.Builder.addAllSecretIds(items: Iterable<String>): TaskExecStartRequest.Builder = apply { this.secret_ids = this.secret_ids + items }
fun TaskExecStartRequest.Builder.clearSecretIds(): TaskExecStartRequest.Builder = apply { secret_ids = emptyList() }

fun TaskExecStartResponse.Companion.newBuilder(): TaskExecStartResponse.Builder = TaskExecStartResponse.Builder()
fun TaskExecStartResponse.Companion.getDefaultInstance(): TaskExecStartResponse = TaskExecStartResponse()
fun TaskExecStartResponse.toByteArray(): ByteArray = encode()

fun TaskExecStdinWriteRequest.Companion.newBuilder(): TaskExecStdinWriteRequest.Builder = TaskExecStdinWriteRequest.Builder()
fun TaskExecStdinWriteRequest.Companion.getDefaultInstance(): TaskExecStdinWriteRequest = TaskExecStdinWriteRequest()
fun TaskExecStdinWriteRequest.toByteArray(): ByteArray = encode()
val TaskExecStdinWriteRequest.taskId: String get() = task_id
fun TaskExecStdinWriteRequest.Builder.setTaskId(value: String): TaskExecStdinWriteRequest.Builder = apply { task_id = value }
var TaskExecStdinWriteRequest.Builder.taskId: String
  get() = task_id
  set(value) { task_id = value }
val TaskExecStdinWriteRequest.execId: String get() = exec_id
fun TaskExecStdinWriteRequest.Builder.setExecId(value: String): TaskExecStdinWriteRequest.Builder = apply { exec_id = value }
var TaskExecStdinWriteRequest.Builder.execId: String
  get() = exec_id
  set(value) { exec_id = value }
fun TaskExecStdinWriteRequest.Builder.setOffset(value: Long): TaskExecStdinWriteRequest.Builder = apply { offset = value }
var TaskExecStdinWriteRequest.Builder.offset: Long
  get() = offset
  set(value) { offset = value }
val TaskExecStdinWriteRequest.data: ByteString get() = data_
fun TaskExecStdinWriteRequest.Builder.setData(value: ByteString): TaskExecStdinWriteRequest.Builder = apply { data_ = value }
var TaskExecStdinWriteRequest.Builder.data: ByteString
  get() = data_
  set(value) { data_ = value }
fun TaskExecStdinWriteRequest.Builder.setData(value: ProtoByteString): TaskExecStdinWriteRequest.Builder = apply { data_ = value.toByteArray().toByteString() }
fun TaskExecStdinWriteRequest.Builder.setEof(value: Boolean): TaskExecStdinWriteRequest.Builder = apply { eof = value }
var TaskExecStdinWriteRequest.Builder.eof: Boolean
  get() = eof
  set(value) { eof = value }

fun TaskExecStdinWriteResponse.Companion.newBuilder(): TaskExecStdinWriteResponse.Builder = TaskExecStdinWriteResponse.Builder()
fun TaskExecStdinWriteResponse.Companion.getDefaultInstance(): TaskExecStdinWriteResponse = TaskExecStdinWriteResponse()
fun TaskExecStdinWriteResponse.toByteArray(): ByteArray = encode()

fun TaskExecStdioReadRequest.Companion.newBuilder(): TaskExecStdioReadRequest.Builder = TaskExecStdioReadRequest.Builder()
fun TaskExecStdioReadRequest.Companion.getDefaultInstance(): TaskExecStdioReadRequest = TaskExecStdioReadRequest()
fun TaskExecStdioReadRequest.toByteArray(): ByteArray = encode()
val TaskExecStdioReadRequest.taskId: String get() = task_id
fun TaskExecStdioReadRequest.Builder.setTaskId(value: String): TaskExecStdioReadRequest.Builder = apply { task_id = value }
var TaskExecStdioReadRequest.Builder.taskId: String
  get() = task_id
  set(value) { task_id = value }
val TaskExecStdioReadRequest.execId: String get() = exec_id
fun TaskExecStdioReadRequest.Builder.setExecId(value: String): TaskExecStdioReadRequest.Builder = apply { exec_id = value }
var TaskExecStdioReadRequest.Builder.execId: String
  get() = exec_id
  set(value) { exec_id = value }
fun TaskExecStdioReadRequest.Builder.setOffset(value: Long): TaskExecStdioReadRequest.Builder = apply { offset = value }
var TaskExecStdioReadRequest.Builder.offset: Long
  get() = offset
  set(value) { offset = value }
val TaskExecStdioReadRequest.fileDescriptor: TaskExecStdioFileDescriptor get() = file_descriptor
fun TaskExecStdioReadRequest.Builder.setFileDescriptor(value: TaskExecStdioFileDescriptor): TaskExecStdioReadRequest.Builder = apply { file_descriptor = value }
var TaskExecStdioReadRequest.Builder.fileDescriptor: TaskExecStdioFileDescriptor
  get() = file_descriptor
  set(value) { file_descriptor = value }

fun TaskExecStdioReadResponse.Companion.newBuilder(): TaskExecStdioReadResponse.Builder = TaskExecStdioReadResponse.Builder()
fun TaskExecStdioReadResponse.Companion.getDefaultInstance(): TaskExecStdioReadResponse = TaskExecStdioReadResponse()
fun TaskExecStdioReadResponse.toByteArray(): ByteArray = encode()
val TaskExecStdioReadResponse.data: ByteString get() = data_
fun TaskExecStdioReadResponse.Builder.setData(value: ByteString): TaskExecStdioReadResponse.Builder = apply { data_ = value }
var TaskExecStdioReadResponse.Builder.data: ByteString
  get() = data_
  set(value) { data_ = value }
fun TaskExecStdioReadResponse.Builder.setData(value: ProtoByteString): TaskExecStdioReadResponse.Builder = apply { data_ = value.toByteArray().toByteString() }

fun TaskExecWaitRequest.Companion.newBuilder(): TaskExecWaitRequest.Builder = TaskExecWaitRequest.Builder()
fun TaskExecWaitRequest.Companion.getDefaultInstance(): TaskExecWaitRequest = TaskExecWaitRequest()
fun TaskExecWaitRequest.toByteArray(): ByteArray = encode()
val TaskExecWaitRequest.taskId: String get() = task_id
fun TaskExecWaitRequest.Builder.setTaskId(value: String): TaskExecWaitRequest.Builder = apply { task_id = value }
var TaskExecWaitRequest.Builder.taskId: String
  get() = task_id
  set(value) { task_id = value }
val TaskExecWaitRequest.execId: String get() = exec_id
fun TaskExecWaitRequest.Builder.setExecId(value: String): TaskExecWaitRequest.Builder = apply { exec_id = value }
var TaskExecWaitRequest.Builder.execId: String
  get() = exec_id
  set(value) { exec_id = value }

fun TaskExecWaitResponse.Companion.newBuilder(): TaskExecWaitResponse.Builder = TaskExecWaitResponse.Builder()
fun TaskExecWaitResponse.Companion.getDefaultInstance(): TaskExecWaitResponse = TaskExecWaitResponse()
fun TaskExecWaitResponse.toByteArray(): ByteArray = encode()
fun TaskExecWaitResponse.hasCode(): Boolean = code != null
fun TaskExecWaitResponse.Builder.setCode(value: Int?): TaskExecWaitResponse.Builder = apply { code = value }
var TaskExecWaitResponse.Builder.code: Int?
  get() = code
  set(value) { code = value }
fun TaskExecWaitResponse.hasSignal(): Boolean = signal != null
fun TaskExecWaitResponse.Builder.setSignal(value: Int?): TaskExecWaitResponse.Builder = apply { signal = value }
var TaskExecWaitResponse.Builder.signal: Int?
  get() = signal
  set(value) { signal = value }

fun TaskMountDirectoryRequest.Companion.newBuilder(): TaskMountDirectoryRequest.Builder = TaskMountDirectoryRequest.Builder()
fun TaskMountDirectoryRequest.Companion.getDefaultInstance(): TaskMountDirectoryRequest = TaskMountDirectoryRequest()
fun TaskMountDirectoryRequest.toByteArray(): ByteArray = encode()
val TaskMountDirectoryRequest.taskId: String get() = task_id
fun TaskMountDirectoryRequest.Builder.setTaskId(value: String): TaskMountDirectoryRequest.Builder = apply { task_id = value }
var TaskMountDirectoryRequest.Builder.taskId: String
  get() = task_id
  set(value) { task_id = value }
fun TaskMountDirectoryRequest.Builder.setPath(value: ByteString): TaskMountDirectoryRequest.Builder = apply { path = value }
var TaskMountDirectoryRequest.Builder.path: ByteString
  get() = path
  set(value) { path = value }
fun TaskMountDirectoryRequest.Builder.setPath(value: ProtoByteString): TaskMountDirectoryRequest.Builder = apply { path = value.toByteArray().toByteString() }
val TaskMountDirectoryRequest.imageId: String get() = image_id
fun TaskMountDirectoryRequest.Builder.setImageId(value: String): TaskMountDirectoryRequest.Builder = apply { image_id = value }
var TaskMountDirectoryRequest.Builder.imageId: String
  get() = image_id
  set(value) { image_id = value }

fun TaskSnapshotDirectoryRequest.Companion.newBuilder(): TaskSnapshotDirectoryRequest.Builder = TaskSnapshotDirectoryRequest.Builder()
fun TaskSnapshotDirectoryRequest.Companion.getDefaultInstance(): TaskSnapshotDirectoryRequest = TaskSnapshotDirectoryRequest()
fun TaskSnapshotDirectoryRequest.toByteArray(): ByteArray = encode()
val TaskSnapshotDirectoryRequest.taskId: String get() = task_id
fun TaskSnapshotDirectoryRequest.Builder.setTaskId(value: String): TaskSnapshotDirectoryRequest.Builder = apply { task_id = value }
var TaskSnapshotDirectoryRequest.Builder.taskId: String
  get() = task_id
  set(value) { task_id = value }
fun TaskSnapshotDirectoryRequest.Builder.setPath(value: ByteString): TaskSnapshotDirectoryRequest.Builder = apply { path = value }
var TaskSnapshotDirectoryRequest.Builder.path: ByteString
  get() = path
  set(value) { path = value }
fun TaskSnapshotDirectoryRequest.Builder.setPath(value: ProtoByteString): TaskSnapshotDirectoryRequest.Builder = apply { path = value.toByteArray().toByteString() }

fun TaskSnapshotDirectoryResponse.Companion.newBuilder(): TaskSnapshotDirectoryResponse.Builder = TaskSnapshotDirectoryResponse.Builder()
fun TaskSnapshotDirectoryResponse.Companion.getDefaultInstance(): TaskSnapshotDirectoryResponse = TaskSnapshotDirectoryResponse()
fun TaskSnapshotDirectoryResponse.toByteArray(): ByteArray = encode()
val TaskSnapshotDirectoryResponse.imageId: String get() = image_id
fun TaskSnapshotDirectoryResponse.Builder.setImageId(value: String): TaskSnapshotDirectoryResponse.Builder = apply { image_id = value }
var TaskSnapshotDirectoryResponse.Builder.imageId: String
  get() = image_id
  set(value) { image_id = value }


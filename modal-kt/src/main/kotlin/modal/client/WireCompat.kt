package modal.client

import com.google.protobuf.ByteString as ProtoByteString
import okio.ByteString
import okio.ByteString.Companion.toByteString

@Suppress(
    "unused",
    "FunctionName",
)
fun AppGetOrCreateRequest.Companion.newBuilder(): AppGetOrCreateRequest.Builder = AppGetOrCreateRequest.Builder()
fun AppGetOrCreateRequest.Companion.getDefaultInstance(): AppGetOrCreateRequest = AppGetOrCreateRequest()
fun AppGetOrCreateRequest.toByteArray(): ByteArray = encode()
val AppGetOrCreateRequest.appName: String get() = app_name
fun AppGetOrCreateRequest.Builder.setAppName(value: String): AppGetOrCreateRequest.Builder = apply { app_name = value }
var AppGetOrCreateRequest.Builder.appName: String
  get() = app_name
  set(value) { app_name = value }
val AppGetOrCreateRequest.environmentName: String get() = environment_name
fun AppGetOrCreateRequest.Builder.setEnvironmentName(value: String): AppGetOrCreateRequest.Builder = apply { environment_name = value }
var AppGetOrCreateRequest.Builder.environmentName: String
  get() = environment_name
  set(value) { environment_name = value }
val AppGetOrCreateRequest.objectCreationType: ObjectCreationType get() = object_creation_type
fun AppGetOrCreateRequest.Builder.setObjectCreationType(value: ObjectCreationType): AppGetOrCreateRequest.Builder = apply { object_creation_type = value }
var AppGetOrCreateRequest.Builder.objectCreationType: ObjectCreationType
  get() = object_creation_type
  set(value) { object_creation_type = value }

fun AppGetOrCreateResponse.Companion.newBuilder(): AppGetOrCreateResponse.Builder = AppGetOrCreateResponse.Builder()
fun AppGetOrCreateResponse.Companion.getDefaultInstance(): AppGetOrCreateResponse = AppGetOrCreateResponse()
fun AppGetOrCreateResponse.toByteArray(): ByteArray = encode()
val AppGetOrCreateResponse.appId: String get() = app_id
fun AppGetOrCreateResponse.Builder.setAppId(value: String): AppGetOrCreateResponse.Builder = apply { app_id = value }
var AppGetOrCreateResponse.Builder.appId: String
  get() = app_id
  set(value) { app_id = value }

fun AttemptAwaitRequest.Companion.newBuilder(): AttemptAwaitRequest.Builder = AttemptAwaitRequest.Builder()
fun AttemptAwaitRequest.Companion.getDefaultInstance(): AttemptAwaitRequest = AttemptAwaitRequest()
fun AttemptAwaitRequest.toByteArray(): ByteArray = encode()
val AttemptAwaitRequest.attemptToken: String get() = attempt_token
fun AttemptAwaitRequest.Builder.setAttemptToken(value: String): AttemptAwaitRequest.Builder = apply { attempt_token = value }
var AttemptAwaitRequest.Builder.attemptToken: String
  get() = attempt_token
  set(value) { attempt_token = value }
val AttemptAwaitRequest.requestedAt: Double get() = requested_at
fun AttemptAwaitRequest.Builder.setRequestedAt(value: Double): AttemptAwaitRequest.Builder = apply { requested_at = value }
var AttemptAwaitRequest.Builder.requestedAt: Double
  get() = requested_at
  set(value) { requested_at = value }
val AttemptAwaitRequest.timeoutSecs: Float get() = timeout_secs
fun AttemptAwaitRequest.Builder.setTimeoutSecs(value: Float): AttemptAwaitRequest.Builder = apply { timeout_secs = value }
var AttemptAwaitRequest.Builder.timeoutSecs: Float
  get() = timeout_secs
  set(value) { timeout_secs = value }

fun AttemptAwaitResponse.Companion.newBuilder(): AttemptAwaitResponse.Builder = AttemptAwaitResponse.Builder()
fun AttemptAwaitResponse.Companion.getDefaultInstance(): AttemptAwaitResponse = AttemptAwaitResponse()
fun AttemptAwaitResponse.toByteArray(): ByteArray = encode()
fun AttemptAwaitResponse.hasOutput(): Boolean = output != null
fun AttemptAwaitResponse.Builder.setOutput(value: FunctionGetOutputsItem?): AttemptAwaitResponse.Builder = apply { output = value }
var AttemptAwaitResponse.Builder.output: FunctionGetOutputsItem?
  get() = output
  set(value) { output = value }

fun AttemptRetryRequest.Companion.newBuilder(): AttemptRetryRequest.Builder = AttemptRetryRequest.Builder()
fun AttemptRetryRequest.Companion.getDefaultInstance(): AttemptRetryRequest = AttemptRetryRequest()
fun AttemptRetryRequest.toByteArray(): ByteArray = encode()
val AttemptRetryRequest.functionId: String get() = function_id
fun AttemptRetryRequest.Builder.setFunctionId(value: String): AttemptRetryRequest.Builder = apply { function_id = value }
var AttemptRetryRequest.Builder.functionId: String
  get() = function_id
  set(value) { function_id = value }
val AttemptRetryRequest.parentInputId: String get() = parent_input_id
fun AttemptRetryRequest.Builder.setParentInputId(value: String): AttemptRetryRequest.Builder = apply { parent_input_id = value }
var AttemptRetryRequest.Builder.parentInputId: String
  get() = parent_input_id
  set(value) { parent_input_id = value }
fun AttemptRetryRequest.hasInput(): Boolean = input != null
fun AttemptRetryRequest.Builder.setInput(value: FunctionPutInputsItem?): AttemptRetryRequest.Builder = apply { input = value }
var AttemptRetryRequest.Builder.input: FunctionPutInputsItem?
  get() = input
  set(value) { input = value }
val AttemptRetryRequest.attemptToken: String get() = attempt_token
fun AttemptRetryRequest.Builder.setAttemptToken(value: String): AttemptRetryRequest.Builder = apply { attempt_token = value }
var AttemptRetryRequest.Builder.attemptToken: String
  get() = attempt_token
  set(value) { attempt_token = value }

fun AttemptRetryResponse.Companion.newBuilder(): AttemptRetryResponse.Builder = AttemptRetryResponse.Builder()
fun AttemptRetryResponse.Companion.getDefaultInstance(): AttemptRetryResponse = AttemptRetryResponse()
fun AttemptRetryResponse.toByteArray(): ByteArray = encode()
val AttemptRetryResponse.attemptToken: String get() = attempt_token
fun AttemptRetryResponse.Builder.setAttemptToken(value: String): AttemptRetryResponse.Builder = apply { attempt_token = value }
var AttemptRetryResponse.Builder.attemptToken: String
  get() = attempt_token
  set(value) { attempt_token = value }

fun AttemptStartRequest.Companion.newBuilder(): AttemptStartRequest.Builder = AttemptStartRequest.Builder()
fun AttemptStartRequest.Companion.getDefaultInstance(): AttemptStartRequest = AttemptStartRequest()
fun AttemptStartRequest.toByteArray(): ByteArray = encode()
val AttemptStartRequest.functionId: String get() = function_id
fun AttemptStartRequest.Builder.setFunctionId(value: String): AttemptStartRequest.Builder = apply { function_id = value }
var AttemptStartRequest.Builder.functionId: String
  get() = function_id
  set(value) { function_id = value }
val AttemptStartRequest.parentInputId: String get() = parent_input_id
fun AttemptStartRequest.Builder.setParentInputId(value: String): AttemptStartRequest.Builder = apply { parent_input_id = value }
var AttemptStartRequest.Builder.parentInputId: String
  get() = parent_input_id
  set(value) { parent_input_id = value }
fun AttemptStartRequest.hasInput(): Boolean = input != null
fun AttemptStartRequest.Builder.setInput(value: FunctionPutInputsItem?): AttemptStartRequest.Builder = apply { input = value }
var AttemptStartRequest.Builder.input: FunctionPutInputsItem?
  get() = input
  set(value) { input = value }

fun AttemptStartResponse.Companion.newBuilder(): AttemptStartResponse.Builder = AttemptStartResponse.Builder()
fun AttemptStartResponse.Companion.getDefaultInstance(): AttemptStartResponse = AttemptStartResponse()
fun AttemptStartResponse.toByteArray(): ByteArray = encode()
val AttemptStartResponse.attemptToken: String get() = attempt_token
fun AttemptStartResponse.Builder.setAttemptToken(value: String): AttemptStartResponse.Builder = apply { attempt_token = value }
var AttemptStartResponse.Builder.attemptToken: String
  get() = attempt_token
  set(value) { attempt_token = value }
val AttemptStartResponse.retryPolicy: FunctionRetryPolicy? get() = retry_policy
fun AttemptStartResponse.hasRetryPolicy(): Boolean = retry_policy != null
fun AttemptStartResponse.Builder.setRetryPolicy(value: FunctionRetryPolicy?): AttemptStartResponse.Builder = apply { retry_policy = value }
var AttemptStartResponse.Builder.retryPolicy: FunctionRetryPolicy?
  get() = retry_policy
  set(value) { retry_policy = value }

fun AuthTokenGetRequest.Companion.newBuilder(): AuthTokenGetRequest.Builder = AuthTokenGetRequest.Builder()
fun AuthTokenGetRequest.Companion.getDefaultInstance(): AuthTokenGetRequest = AuthTokenGetRequest()
fun AuthTokenGetRequest.toByteArray(): ByteArray = encode()

fun AuthTokenGetResponse.Companion.newBuilder(): AuthTokenGetResponse.Builder = AuthTokenGetResponse.Builder()
fun AuthTokenGetResponse.Companion.getDefaultInstance(): AuthTokenGetResponse = AuthTokenGetResponse()
fun AuthTokenGetResponse.toByteArray(): ByteArray = encode()
fun AuthTokenGetResponse.Builder.setToken(value: String): AuthTokenGetResponse.Builder = apply { token = value }
var AuthTokenGetResponse.Builder.token: String
  get() = token
  set(value) { token = value }

fun AutoscalerSettings.Companion.newBuilder(): AutoscalerSettings.Builder = AutoscalerSettings.Builder()
fun AutoscalerSettings.Companion.getDefaultInstance(): AutoscalerSettings = AutoscalerSettings()
fun AutoscalerSettings.toByteArray(): ByteArray = encode()
val AutoscalerSettings.minContainers: Int? get() = min_containers
fun AutoscalerSettings.hasMinContainers(): Boolean = min_containers != null
fun AutoscalerSettings.Builder.setMinContainers(value: Int?): AutoscalerSettings.Builder = apply { min_containers = value }
var AutoscalerSettings.Builder.minContainers: Int?
  get() = min_containers
  set(value) { min_containers = value }
val AutoscalerSettings.maxContainers: Int? get() = max_containers
fun AutoscalerSettings.hasMaxContainers(): Boolean = max_containers != null
fun AutoscalerSettings.Builder.setMaxContainers(value: Int?): AutoscalerSettings.Builder = apply { max_containers = value }
var AutoscalerSettings.Builder.maxContainers: Int?
  get() = max_containers
  set(value) { max_containers = value }
val AutoscalerSettings.bufferContainers: Int? get() = buffer_containers
fun AutoscalerSettings.hasBufferContainers(): Boolean = buffer_containers != null
fun AutoscalerSettings.Builder.setBufferContainers(value: Int?): AutoscalerSettings.Builder = apply { buffer_containers = value }
var AutoscalerSettings.Builder.bufferContainers: Int?
  get() = buffer_containers
  set(value) { buffer_containers = value }
val AutoscalerSettings.scaleupWindow: Int? get() = scaleup_window
fun AutoscalerSettings.hasScaleupWindow(): Boolean = scaleup_window != null
fun AutoscalerSettings.Builder.setScaleupWindow(value: Int?): AutoscalerSettings.Builder = apply { scaleup_window = value }
var AutoscalerSettings.Builder.scaleupWindow: Int?
  get() = scaleup_window
  set(value) { scaleup_window = value }
val AutoscalerSettings.scaledownWindow: Int? get() = scaledown_window
fun AutoscalerSettings.hasScaledownWindow(): Boolean = scaledown_window != null
fun AutoscalerSettings.Builder.setScaledownWindow(value: Int?): AutoscalerSettings.Builder = apply { scaledown_window = value }
var AutoscalerSettings.Builder.scaledownWindow: Int?
  get() = scaledown_window
  set(value) { scaledown_window = value }

fun BaseImage.Companion.newBuilder(): BaseImage.Builder = BaseImage.Builder()
fun BaseImage.Companion.getDefaultInstance(): BaseImage = BaseImage()
fun BaseImage.toByteArray(): ByteArray = encode()
val BaseImage.imageId: String get() = image_id
fun BaseImage.Builder.setImageId(value: String): BaseImage.Builder = apply { image_id = value }
var BaseImage.Builder.imageId: String
  get() = image_id
  set(value) { image_id = value }
val BaseImage.dockerTag: String get() = docker_tag
fun BaseImage.Builder.setDockerTag(value: String): BaseImage.Builder = apply { docker_tag = value }
var BaseImage.Builder.dockerTag: String
  get() = docker_tag
  set(value) { docker_tag = value }

fun BlobCreateRequest.Companion.newBuilder(): BlobCreateRequest.Builder = BlobCreateRequest.Builder()
fun BlobCreateRequest.Companion.getDefaultInstance(): BlobCreateRequest = BlobCreateRequest()
fun BlobCreateRequest.toByteArray(): ByteArray = encode()
val BlobCreateRequest.contentMd5: String get() = content_md5
fun BlobCreateRequest.Builder.setContentMd5(value: String): BlobCreateRequest.Builder = apply { content_md5 = value }
var BlobCreateRequest.Builder.contentMd5: String
  get() = content_md5
  set(value) { content_md5 = value }
val BlobCreateRequest.contentSha256Base64: String get() = content_sha256_base64
fun BlobCreateRequest.Builder.setContentSha256Base64(value: String): BlobCreateRequest.Builder = apply { content_sha256_base64 = value }
var BlobCreateRequest.Builder.contentSha256Base64: String
  get() = content_sha256_base64
  set(value) { content_sha256_base64 = value }
val BlobCreateRequest.contentLength: Long get() = content_length
fun BlobCreateRequest.Builder.setContentLength(value: Long): BlobCreateRequest.Builder = apply { content_length = value }
var BlobCreateRequest.Builder.contentLength: Long
  get() = content_length
  set(value) { content_length = value }

fun BlobCreateResponse.Companion.newBuilder(): BlobCreateResponse.Builder = BlobCreateResponse.Builder()
fun BlobCreateResponse.Companion.getDefaultInstance(): BlobCreateResponse = BlobCreateResponse()
fun BlobCreateResponse.toByteArray(): ByteArray = encode()
val BlobCreateResponse.blobId: String get() = blob_id
fun BlobCreateResponse.Builder.setBlobId(value: String): BlobCreateResponse.Builder = apply { blob_id = value }
var BlobCreateResponse.Builder.blobId: String
  get() = blob_id
  set(value) { blob_id = value }
val BlobCreateResponse.uploadUrl: String? get() = upload_url
fun BlobCreateResponse.hasUploadUrl(): Boolean = upload_url != null
fun BlobCreateResponse.Builder.setUploadUrl(value: String?): BlobCreateResponse.Builder = apply { upload_url = value }
var BlobCreateResponse.Builder.uploadUrl: String?
  get() = upload_url
  set(value) { upload_url = value }
fun BlobCreateResponse.hasMultipart(): Boolean = multipart != null
fun BlobCreateResponse.Builder.setMultipart(value: MultiPartUpload?): BlobCreateResponse.Builder = apply { multipart = value }
var BlobCreateResponse.Builder.multipart: MultiPartUpload?
  get() = multipart
  set(value) { multipart = value }
val BlobCreateResponse.uploadUrls: UploadUrlList? get() = upload_urls
fun BlobCreateResponse.hasUploadUrls(): Boolean = upload_urls != null
fun BlobCreateResponse.Builder.setUploadUrls(value: UploadUrlList?): BlobCreateResponse.Builder = apply { upload_urls = value }
var BlobCreateResponse.Builder.uploadUrls: UploadUrlList?
  get() = upload_urls
  set(value) { upload_urls = value }
fun BlobCreateResponse.hasMultiparts(): Boolean = multiparts != null
fun BlobCreateResponse.Builder.setMultiparts(value: MultiPartUploadList?): BlobCreateResponse.Builder = apply { multiparts = value }
var BlobCreateResponse.Builder.multiparts: MultiPartUploadList?
  get() = multiparts
  set(value) { multiparts = value }
val BlobCreateResponse.blobIds: List<String> get() = blob_ids
val BlobCreateResponse.blobIdsList: List<String> get() = blob_ids
val BlobCreateResponse.blobIdsCount: Int get() = blob_ids.size
fun BlobCreateResponse.Builder.addBlobIds(value: String): BlobCreateResponse.Builder = apply { blob_ids = blob_ids + value }
fun BlobCreateResponse.Builder.addAllBlobIds(items: Iterable<String>): BlobCreateResponse.Builder = apply { this.blob_ids = this.blob_ids + items }
fun BlobCreateResponse.Builder.clearBlobIds(): BlobCreateResponse.Builder = apply { blob_ids = emptyList() }

fun BlobGetRequest.Companion.newBuilder(): BlobGetRequest.Builder = BlobGetRequest.Builder()
fun BlobGetRequest.Companion.getDefaultInstance(): BlobGetRequest = BlobGetRequest()
fun BlobGetRequest.toByteArray(): ByteArray = encode()
val BlobGetRequest.blobId: String get() = blob_id
fun BlobGetRequest.Builder.setBlobId(value: String): BlobGetRequest.Builder = apply { blob_id = value }
var BlobGetRequest.Builder.blobId: String
  get() = blob_id
  set(value) { blob_id = value }

fun BlobGetResponse.Companion.newBuilder(): BlobGetResponse.Builder = BlobGetResponse.Builder()
fun BlobGetResponse.Companion.getDefaultInstance(): BlobGetResponse = BlobGetResponse()
fun BlobGetResponse.toByteArray(): ByteArray = encode()
val BlobGetResponse.downloadUrl: String get() = download_url
fun BlobGetResponse.Builder.setDownloadUrl(value: String): BlobGetResponse.Builder = apply { download_url = value }
var BlobGetResponse.Builder.downloadUrl: String
  get() = download_url
  set(value) { download_url = value }

fun BuildFunction.Companion.newBuilder(): BuildFunction.Builder = BuildFunction.Builder()
fun BuildFunction.Companion.getDefaultInstance(): BuildFunction = BuildFunction()
fun BuildFunction.toByteArray(): ByteArray = encode()
fun BuildFunction.Builder.setDefinition(value: String): BuildFunction.Builder = apply { definition = value }
var BuildFunction.Builder.definition: String
  get() = definition
  set(value) { definition = value }
fun BuildFunction.Builder.setGlobals(value: ByteString): BuildFunction.Builder = apply { globals = value }
var BuildFunction.Builder.globals: ByteString
  get() = globals
  set(value) { globals = value }
fun BuildFunction.Builder.setGlobals(value: ProtoByteString): BuildFunction.Builder = apply { globals = value.toByteArray().toByteString() }
fun BuildFunction.hasInput(): Boolean = input != null
fun BuildFunction.Builder.setInput(value: FunctionInput?): BuildFunction.Builder = apply { input = value }
var BuildFunction.Builder.input: FunctionInput?
  get() = input
  set(value) { input = value }

fun CheckpointInfo.Companion.newBuilder(): CheckpointInfo.Builder = CheckpointInfo.Builder()
fun CheckpointInfo.Companion.getDefaultInstance(): CheckpointInfo = CheckpointInfo()
fun CheckpointInfo.toByteArray(): ByteArray = encode()
fun CheckpointInfo.Builder.setChecksum(value: String): CheckpointInfo.Builder = apply { checksum = value }
var CheckpointInfo.Builder.checksum: String
  get() = checksum
  set(value) { checksum = value }
fun CheckpointInfo.Builder.setStatus(value: CheckpointStatus): CheckpointInfo.Builder = apply { status = value }
var CheckpointInfo.Builder.status: CheckpointStatus
  get() = status
  set(value) { status = value }
val CheckpointInfo.checkpointId: String get() = checkpoint_id
fun CheckpointInfo.Builder.setCheckpointId(value: String): CheckpointInfo.Builder = apply { checkpoint_id = value }
var CheckpointInfo.Builder.checkpointId: String
  get() = checkpoint_id
  set(value) { checkpoint_id = value }
val CheckpointInfo.runtimeFingerprint: String get() = runtime_fingerprint
fun CheckpointInfo.Builder.setRuntimeFingerprint(value: String): CheckpointInfo.Builder = apply { runtime_fingerprint = value }
var CheckpointInfo.Builder.runtimeFingerprint: String
  get() = runtime_fingerprint
  set(value) { runtime_fingerprint = value }
fun CheckpointInfo.Builder.setSize(value: Long): CheckpointInfo.Builder = apply { size = value }
var CheckpointInfo.Builder.size: Long
  get() = size
  set(value) { size = value }
val CheckpointInfo.checksumIsFileIndex: Boolean get() = checksum_is_file_index
fun CheckpointInfo.Builder.setChecksumIsFileIndex(value: Boolean): CheckpointInfo.Builder = apply { checksum_is_file_index = value }
var CheckpointInfo.Builder.checksumIsFileIndex: Boolean
  get() = checksum_is_file_index
  set(value) { checksum_is_file_index = value }
val CheckpointInfo.originalTaskId: String get() = original_task_id
fun CheckpointInfo.Builder.setOriginalTaskId(value: String): CheckpointInfo.Builder = apply { original_task_id = value }
var CheckpointInfo.Builder.originalTaskId: String
  get() = original_task_id
  set(value) { original_task_id = value }
val CheckpointInfo.runscRuntimeVersion: String get() = runsc_runtime_version
fun CheckpointInfo.Builder.setRunscRuntimeVersion(value: String): CheckpointInfo.Builder = apply { runsc_runtime_version = value }
var CheckpointInfo.Builder.runscRuntimeVersion: String
  get() = runsc_runtime_version
  set(value) { runsc_runtime_version = value }

fun ClassParameterInfo.Companion.newBuilder(): ClassParameterInfo.Builder = ClassParameterInfo.Builder()
fun ClassParameterInfo.Companion.getDefaultInstance(): ClassParameterInfo = ClassParameterInfo()
fun ClassParameterInfo.toByteArray(): ByteArray = encode()
fun ClassParameterInfo.Builder.setFormat(value: ClassParameterInfo.ParameterSerializationFormat): ClassParameterInfo.Builder = apply { format = value }
var ClassParameterInfo.Builder.format: ClassParameterInfo.ParameterSerializationFormat
  get() = format
  set(value) { format = value }
val ClassParameterInfo.schemaList: List<ClassParameterSpec> get() = schema
val ClassParameterInfo.schemaCount: Int get() = schema.size
fun ClassParameterInfo.Builder.addSchema(value: ClassParameterSpec): ClassParameterInfo.Builder = apply { schema = schema + value }
fun ClassParameterInfo.Builder.addAllSchema(items: Iterable<ClassParameterSpec>): ClassParameterInfo.Builder = apply { this.schema = this.schema + items }
fun ClassParameterInfo.Builder.clearSchema(): ClassParameterInfo.Builder = apply { schema = emptyList() }

fun ClassParameterSet.Companion.newBuilder(): ClassParameterSet.Builder = ClassParameterSet.Builder()
fun ClassParameterSet.Companion.getDefaultInstance(): ClassParameterSet = ClassParameterSet()
fun ClassParameterSet.toByteArray(): ByteArray = encode()
val ClassParameterSet.parametersList: List<ClassParameterValue> get() = parameters
val ClassParameterSet.parametersCount: Int get() = parameters.size
fun ClassParameterSet.Builder.addParameters(value: ClassParameterValue): ClassParameterSet.Builder = apply { parameters = parameters + value }
fun ClassParameterSet.Builder.addAllParameters(items: Iterable<ClassParameterValue>): ClassParameterSet.Builder = apply { this.parameters = this.parameters + items }
fun ClassParameterSet.Builder.clearParameters(): ClassParameterSet.Builder = apply { parameters = emptyList() }

fun ClassParameterSpec.Companion.newBuilder(): ClassParameterSpec.Builder = ClassParameterSpec.Builder()
fun ClassParameterSpec.Companion.getDefaultInstance(): ClassParameterSpec = ClassParameterSpec()
fun ClassParameterSpec.toByteArray(): ByteArray = encode()
fun ClassParameterSpec.Builder.setName(value: String): ClassParameterSpec.Builder = apply { name = value }
var ClassParameterSpec.Builder.name: String
  get() = name
  set(value) { name = value }
fun ClassParameterSpec.Builder.setType(value: ParameterType): ClassParameterSpec.Builder = apply { type = value }
var ClassParameterSpec.Builder.type: ParameterType
  get() = type
  set(value) { type = value }
val ClassParameterSpec.hasDefault: Boolean get() = has_default
fun ClassParameterSpec.Builder.setHasDefault(value: Boolean): ClassParameterSpec.Builder = apply { has_default = value }
var ClassParameterSpec.Builder.hasDefault: Boolean
  get() = has_default
  set(value) { has_default = value }
val ClassParameterSpec.stringDefault: String? get() = string_default
fun ClassParameterSpec.hasStringDefault(): Boolean = string_default != null
fun ClassParameterSpec.Builder.setStringDefault(value: String?): ClassParameterSpec.Builder = apply { string_default = value }
var ClassParameterSpec.Builder.stringDefault: String?
  get() = string_default
  set(value) { string_default = value }
val ClassParameterSpec.intDefault: Long? get() = int_default
fun ClassParameterSpec.hasIntDefault(): Boolean = int_default != null
fun ClassParameterSpec.Builder.setIntDefault(value: Long?): ClassParameterSpec.Builder = apply { int_default = value }
var ClassParameterSpec.Builder.intDefault: Long?
  get() = int_default
  set(value) { int_default = value }
val ClassParameterSpec.pickleDefault: ByteString? get() = pickle_default
fun ClassParameterSpec.hasPickleDefault(): Boolean = pickle_default != null
fun ClassParameterSpec.Builder.setPickleDefault(value: ByteString?): ClassParameterSpec.Builder = apply { pickle_default = value }
var ClassParameterSpec.Builder.pickleDefault: ByteString?
  get() = pickle_default
  set(value) { pickle_default = value }
fun ClassParameterSpec.Builder.setPickleDefault(value: ProtoByteString?): ClassParameterSpec.Builder = apply { pickle_default = value?.toByteArray()?.toByteString() }
val ClassParameterSpec.bytesDefault: ByteString? get() = bytes_default
fun ClassParameterSpec.hasBytesDefault(): Boolean = bytes_default != null
fun ClassParameterSpec.Builder.setBytesDefault(value: ByteString?): ClassParameterSpec.Builder = apply { bytes_default = value }
var ClassParameterSpec.Builder.bytesDefault: ByteString?
  get() = bytes_default
  set(value) { bytes_default = value }
fun ClassParameterSpec.Builder.setBytesDefault(value: ProtoByteString?): ClassParameterSpec.Builder = apply { bytes_default = value?.toByteArray()?.toByteString() }
val ClassParameterSpec.boolDefault: Boolean? get() = bool_default
fun ClassParameterSpec.hasBoolDefault(): Boolean = bool_default != null
fun ClassParameterSpec.Builder.setBoolDefault(value: Boolean?): ClassParameterSpec.Builder = apply { bool_default = value }
var ClassParameterSpec.Builder.boolDefault: Boolean?
  get() = bool_default
  set(value) { bool_default = value }
val ClassParameterSpec.fullType: GenericPayloadType? get() = full_type
fun ClassParameterSpec.hasFullType(): Boolean = full_type != null
fun ClassParameterSpec.Builder.setFullType(value: GenericPayloadType?): ClassParameterSpec.Builder = apply { full_type = value }
var ClassParameterSpec.Builder.fullType: GenericPayloadType?
  get() = full_type
  set(value) { full_type = value }

fun ClassParameterValue.Companion.newBuilder(): ClassParameterValue.Builder = ClassParameterValue.Builder()
fun ClassParameterValue.Companion.getDefaultInstance(): ClassParameterValue = ClassParameterValue()
fun ClassParameterValue.toByteArray(): ByteArray = encode()
fun ClassParameterValue.Builder.setName(value: String): ClassParameterValue.Builder = apply { name = value }
var ClassParameterValue.Builder.name: String
  get() = name
  set(value) { name = value }
fun ClassParameterValue.Builder.setType(value: ParameterType): ClassParameterValue.Builder = apply { type = value }
var ClassParameterValue.Builder.type: ParameterType
  get() = type
  set(value) { type = value }
val ClassParameterValue.stringValue: String? get() = string_value
fun ClassParameterValue.hasStringValue(): Boolean = string_value != null
fun ClassParameterValue.Builder.setStringValue(value: String?): ClassParameterValue.Builder = apply { string_value = value }
var ClassParameterValue.Builder.stringValue: String?
  get() = string_value
  set(value) { string_value = value }
val ClassParameterValue.intValue: Long? get() = int_value
fun ClassParameterValue.hasIntValue(): Boolean = int_value != null
fun ClassParameterValue.Builder.setIntValue(value: Long?): ClassParameterValue.Builder = apply { int_value = value }
var ClassParameterValue.Builder.intValue: Long?
  get() = int_value
  set(value) { int_value = value }
val ClassParameterValue.pickleValue: ByteString? get() = pickle_value
fun ClassParameterValue.hasPickleValue(): Boolean = pickle_value != null
fun ClassParameterValue.Builder.setPickleValue(value: ByteString?): ClassParameterValue.Builder = apply { pickle_value = value }
var ClassParameterValue.Builder.pickleValue: ByteString?
  get() = pickle_value
  set(value) { pickle_value = value }
fun ClassParameterValue.Builder.setPickleValue(value: ProtoByteString?): ClassParameterValue.Builder = apply { pickle_value = value?.toByteArray()?.toByteString() }
val ClassParameterValue.bytesValue: ByteString? get() = bytes_value
fun ClassParameterValue.hasBytesValue(): Boolean = bytes_value != null
fun ClassParameterValue.Builder.setBytesValue(value: ByteString?): ClassParameterValue.Builder = apply { bytes_value = value }
var ClassParameterValue.Builder.bytesValue: ByteString?
  get() = bytes_value
  set(value) { bytes_value = value }
fun ClassParameterValue.Builder.setBytesValue(value: ProtoByteString?): ClassParameterValue.Builder = apply { bytes_value = value?.toByteArray()?.toByteString() }
val ClassParameterValue.boolValue: Boolean? get() = bool_value
fun ClassParameterValue.hasBoolValue(): Boolean = bool_value != null
fun ClassParameterValue.Builder.setBoolValue(value: Boolean?): ClassParameterValue.Builder = apply { bool_value = value }
var ClassParameterValue.Builder.boolValue: Boolean?
  get() = bool_value
  set(value) { bool_value = value }

fun CloudBucketMount.Companion.newBuilder(): CloudBucketMount.Builder = CloudBucketMount.Builder()
fun CloudBucketMount.Companion.getDefaultInstance(): CloudBucketMount = CloudBucketMount()
fun CloudBucketMount.toByteArray(): ByteArray = encode()
val CloudBucketMount.bucketName: String get() = bucket_name
fun CloudBucketMount.Builder.setBucketName(value: String): CloudBucketMount.Builder = apply { bucket_name = value }
var CloudBucketMount.Builder.bucketName: String
  get() = bucket_name
  set(value) { bucket_name = value }
val CloudBucketMount.mountPath: String get() = mount_path
fun CloudBucketMount.Builder.setMountPath(value: String): CloudBucketMount.Builder = apply { mount_path = value }
var CloudBucketMount.Builder.mountPath: String
  get() = mount_path
  set(value) { mount_path = value }
val CloudBucketMount.credentialsSecretId: String get() = credentials_secret_id
fun CloudBucketMount.Builder.setCredentialsSecretId(value: String): CloudBucketMount.Builder = apply { credentials_secret_id = value }
var CloudBucketMount.Builder.credentialsSecretId: String
  get() = credentials_secret_id
  set(value) { credentials_secret_id = value }
val CloudBucketMount.readOnly: Boolean get() = read_only
fun CloudBucketMount.Builder.setReadOnly(value: Boolean): CloudBucketMount.Builder = apply { read_only = value }
var CloudBucketMount.Builder.readOnly: Boolean
  get() = read_only
  set(value) { read_only = value }
val CloudBucketMount.bucketType: CloudBucketMount.BucketType get() = bucket_type
fun CloudBucketMount.Builder.setBucketType(value: CloudBucketMount.BucketType): CloudBucketMount.Builder = apply { bucket_type = value }
var CloudBucketMount.Builder.bucketType: CloudBucketMount.BucketType
  get() = bucket_type
  set(value) { bucket_type = value }
val CloudBucketMount.requesterPays: Boolean get() = requester_pays
fun CloudBucketMount.Builder.setRequesterPays(value: Boolean): CloudBucketMount.Builder = apply { requester_pays = value }
var CloudBucketMount.Builder.requesterPays: Boolean
  get() = requester_pays
  set(value) { requester_pays = value }
val CloudBucketMount.bucketEndpointUrl: String? get() = bucket_endpoint_url
fun CloudBucketMount.hasBucketEndpointUrl(): Boolean = bucket_endpoint_url != null
fun CloudBucketMount.Builder.setBucketEndpointUrl(value: String?): CloudBucketMount.Builder = apply { bucket_endpoint_url = value }
var CloudBucketMount.Builder.bucketEndpointUrl: String?
  get() = bucket_endpoint_url
  set(value) { bucket_endpoint_url = value }
val CloudBucketMount.keyPrefix: String? get() = key_prefix
fun CloudBucketMount.hasKeyPrefix(): Boolean = key_prefix != null
fun CloudBucketMount.Builder.setKeyPrefix(value: String?): CloudBucketMount.Builder = apply { key_prefix = value }
var CloudBucketMount.Builder.keyPrefix: String?
  get() = key_prefix
  set(value) { key_prefix = value }
val CloudBucketMount.oidcAuthRoleArn: String? get() = oidc_auth_role_arn
fun CloudBucketMount.hasOidcAuthRoleArn(): Boolean = oidc_auth_role_arn != null
fun CloudBucketMount.Builder.setOidcAuthRoleArn(value: String?): CloudBucketMount.Builder = apply { oidc_auth_role_arn = value }
var CloudBucketMount.Builder.oidcAuthRoleArn: String?
  get() = oidc_auth_role_arn
  set(value) { oidc_auth_role_arn = value }
val CloudBucketMount.forcePathStyle: Boolean get() = force_path_style
fun CloudBucketMount.Builder.setForcePathStyle(value: Boolean): CloudBucketMount.Builder = apply { force_path_style = value }
var CloudBucketMount.Builder.forcePathStyle: Boolean
  get() = force_path_style
  set(value) { force_path_style = value }
val CloudBucketMount.metadataTtlType: CloudBucketMount.MetadataTTLType? get() = metadata_ttl_type
fun CloudBucketMount.hasMetadataTtlType(): Boolean = metadata_ttl_type != null
fun CloudBucketMount.Builder.setMetadataTtlType(value: CloudBucketMount.MetadataTTLType?): CloudBucketMount.Builder = apply { metadata_ttl_type = value }
var CloudBucketMount.Builder.metadataTtlType: CloudBucketMount.MetadataTTLType?
  get() = metadata_ttl_type
  set(value) { metadata_ttl_type = value }
val CloudBucketMount.metadataTtlSeconds: Long? get() = metadata_ttl_seconds
fun CloudBucketMount.hasMetadataTtlSeconds(): Boolean = metadata_ttl_seconds != null
fun CloudBucketMount.Builder.setMetadataTtlSeconds(value: Long?): CloudBucketMount.Builder = apply { metadata_ttl_seconds = value }
var CloudBucketMount.Builder.metadataTtlSeconds: Long?
  get() = metadata_ttl_seconds
  set(value) { metadata_ttl_seconds = value }

fun ContainerFileCloseRequest.Companion.newBuilder(): ContainerFileCloseRequest.Builder = ContainerFileCloseRequest.Builder()
fun ContainerFileCloseRequest.Companion.getDefaultInstance(): ContainerFileCloseRequest = ContainerFileCloseRequest()
fun ContainerFileCloseRequest.toByteArray(): ByteArray = encode()
val ContainerFileCloseRequest.fileDescriptor: String get() = file_descriptor
fun ContainerFileCloseRequest.Builder.setFileDescriptor(value: String): ContainerFileCloseRequest.Builder = apply { file_descriptor = value }
var ContainerFileCloseRequest.Builder.fileDescriptor: String
  get() = file_descriptor
  set(value) { file_descriptor = value }

fun ContainerFileDeleteBytesRequest.Companion.newBuilder(): ContainerFileDeleteBytesRequest.Builder = ContainerFileDeleteBytesRequest.Builder()
fun ContainerFileDeleteBytesRequest.Companion.getDefaultInstance(): ContainerFileDeleteBytesRequest = ContainerFileDeleteBytesRequest()
fun ContainerFileDeleteBytesRequest.toByteArray(): ByteArray = encode()
val ContainerFileDeleteBytesRequest.fileDescriptor: String get() = file_descriptor
fun ContainerFileDeleteBytesRequest.Builder.setFileDescriptor(value: String): ContainerFileDeleteBytesRequest.Builder = apply { file_descriptor = value }
var ContainerFileDeleteBytesRequest.Builder.fileDescriptor: String
  get() = file_descriptor
  set(value) { file_descriptor = value }
val ContainerFileDeleteBytesRequest.startInclusive: Int? get() = start_inclusive
fun ContainerFileDeleteBytesRequest.hasStartInclusive(): Boolean = start_inclusive != null
fun ContainerFileDeleteBytesRequest.Builder.setStartInclusive(value: Int?): ContainerFileDeleteBytesRequest.Builder = apply { start_inclusive = value }
var ContainerFileDeleteBytesRequest.Builder.startInclusive: Int?
  get() = start_inclusive
  set(value) { start_inclusive = value }
val ContainerFileDeleteBytesRequest.endExclusive: Int? get() = end_exclusive
fun ContainerFileDeleteBytesRequest.hasEndExclusive(): Boolean = end_exclusive != null
fun ContainerFileDeleteBytesRequest.Builder.setEndExclusive(value: Int?): ContainerFileDeleteBytesRequest.Builder = apply { end_exclusive = value }
var ContainerFileDeleteBytesRequest.Builder.endExclusive: Int?
  get() = end_exclusive
  set(value) { end_exclusive = value }

fun ContainerFileFlushRequest.Companion.newBuilder(): ContainerFileFlushRequest.Builder = ContainerFileFlushRequest.Builder()
fun ContainerFileFlushRequest.Companion.getDefaultInstance(): ContainerFileFlushRequest = ContainerFileFlushRequest()
fun ContainerFileFlushRequest.toByteArray(): ByteArray = encode()
val ContainerFileFlushRequest.fileDescriptor: String get() = file_descriptor
fun ContainerFileFlushRequest.Builder.setFileDescriptor(value: String): ContainerFileFlushRequest.Builder = apply { file_descriptor = value }
var ContainerFileFlushRequest.Builder.fileDescriptor: String
  get() = file_descriptor
  set(value) { file_descriptor = value }

fun ContainerFileLsRequest.Companion.newBuilder(): ContainerFileLsRequest.Builder = ContainerFileLsRequest.Builder()
fun ContainerFileLsRequest.Companion.getDefaultInstance(): ContainerFileLsRequest = ContainerFileLsRequest()
fun ContainerFileLsRequest.toByteArray(): ByteArray = encode()
fun ContainerFileLsRequest.Builder.setPath(value: String): ContainerFileLsRequest.Builder = apply { path = value }
var ContainerFileLsRequest.Builder.path: String
  get() = path
  set(value) { path = value }

fun ContainerFileMkdirRequest.Companion.newBuilder(): ContainerFileMkdirRequest.Builder = ContainerFileMkdirRequest.Builder()
fun ContainerFileMkdirRequest.Companion.getDefaultInstance(): ContainerFileMkdirRequest = ContainerFileMkdirRequest()
fun ContainerFileMkdirRequest.toByteArray(): ByteArray = encode()
fun ContainerFileMkdirRequest.Builder.setPath(value: String): ContainerFileMkdirRequest.Builder = apply { path = value }
var ContainerFileMkdirRequest.Builder.path: String
  get() = path
  set(value) { path = value }
val ContainerFileMkdirRequest.makeParents: Boolean get() = make_parents
fun ContainerFileMkdirRequest.Builder.setMakeParents(value: Boolean): ContainerFileMkdirRequest.Builder = apply { make_parents = value }
var ContainerFileMkdirRequest.Builder.makeParents: Boolean
  get() = make_parents
  set(value) { make_parents = value }

fun ContainerFileOpenRequest.Companion.newBuilder(): ContainerFileOpenRequest.Builder = ContainerFileOpenRequest.Builder()
fun ContainerFileOpenRequest.Companion.getDefaultInstance(): ContainerFileOpenRequest = ContainerFileOpenRequest()
fun ContainerFileOpenRequest.toByteArray(): ByteArray = encode()
val ContainerFileOpenRequest.fileDescriptor: String? get() = file_descriptor
fun ContainerFileOpenRequest.hasFileDescriptor(): Boolean = file_descriptor != null
fun ContainerFileOpenRequest.Builder.setFileDescriptor(value: String?): ContainerFileOpenRequest.Builder = apply { file_descriptor = value }
var ContainerFileOpenRequest.Builder.fileDescriptor: String?
  get() = file_descriptor
  set(value) { file_descriptor = value }
fun ContainerFileOpenRequest.Builder.setPath(value: String): ContainerFileOpenRequest.Builder = apply { path = value }
var ContainerFileOpenRequest.Builder.path: String
  get() = path
  set(value) { path = value }
fun ContainerFileOpenRequest.Builder.setMode(value: String): ContainerFileOpenRequest.Builder = apply { mode = value }
var ContainerFileOpenRequest.Builder.mode: String
  get() = mode
  set(value) { mode = value }

fun ContainerFileReadLineRequest.Companion.newBuilder(): ContainerFileReadLineRequest.Builder = ContainerFileReadLineRequest.Builder()
fun ContainerFileReadLineRequest.Companion.getDefaultInstance(): ContainerFileReadLineRequest = ContainerFileReadLineRequest()
fun ContainerFileReadLineRequest.toByteArray(): ByteArray = encode()
val ContainerFileReadLineRequest.fileDescriptor: String get() = file_descriptor
fun ContainerFileReadLineRequest.Builder.setFileDescriptor(value: String): ContainerFileReadLineRequest.Builder = apply { file_descriptor = value }
var ContainerFileReadLineRequest.Builder.fileDescriptor: String
  get() = file_descriptor
  set(value) { file_descriptor = value }

fun ContainerFileReadRequest.Companion.newBuilder(): ContainerFileReadRequest.Builder = ContainerFileReadRequest.Builder()
fun ContainerFileReadRequest.Companion.getDefaultInstance(): ContainerFileReadRequest = ContainerFileReadRequest()
fun ContainerFileReadRequest.toByteArray(): ByteArray = encode()
val ContainerFileReadRequest.fileDescriptor: String get() = file_descriptor
fun ContainerFileReadRequest.Builder.setFileDescriptor(value: String): ContainerFileReadRequest.Builder = apply { file_descriptor = value }
var ContainerFileReadRequest.Builder.fileDescriptor: String
  get() = file_descriptor
  set(value) { file_descriptor = value }
fun ContainerFileReadRequest.hasN(): Boolean = n != null
fun ContainerFileReadRequest.Builder.setN(value: Int?): ContainerFileReadRequest.Builder = apply { n = value }
var ContainerFileReadRequest.Builder.n: Int?
  get() = n
  set(value) { n = value }

fun ContainerFileRmRequest.Companion.newBuilder(): ContainerFileRmRequest.Builder = ContainerFileRmRequest.Builder()
fun ContainerFileRmRequest.Companion.getDefaultInstance(): ContainerFileRmRequest = ContainerFileRmRequest()
fun ContainerFileRmRequest.toByteArray(): ByteArray = encode()
fun ContainerFileRmRequest.Builder.setPath(value: String): ContainerFileRmRequest.Builder = apply { path = value }
var ContainerFileRmRequest.Builder.path: String
  get() = path
  set(value) { path = value }
fun ContainerFileRmRequest.Builder.setRecursive(value: Boolean): ContainerFileRmRequest.Builder = apply { recursive = value }
var ContainerFileRmRequest.Builder.recursive: Boolean
  get() = recursive
  set(value) { recursive = value }

fun ContainerFileSeekRequest.Companion.newBuilder(): ContainerFileSeekRequest.Builder = ContainerFileSeekRequest.Builder()
fun ContainerFileSeekRequest.Companion.getDefaultInstance(): ContainerFileSeekRequest = ContainerFileSeekRequest()
fun ContainerFileSeekRequest.toByteArray(): ByteArray = encode()
val ContainerFileSeekRequest.fileDescriptor: String get() = file_descriptor
fun ContainerFileSeekRequest.Builder.setFileDescriptor(value: String): ContainerFileSeekRequest.Builder = apply { file_descriptor = value }
var ContainerFileSeekRequest.Builder.fileDescriptor: String
  get() = file_descriptor
  set(value) { file_descriptor = value }
fun ContainerFileSeekRequest.Builder.setOffset(value: Int): ContainerFileSeekRequest.Builder = apply { offset = value }
var ContainerFileSeekRequest.Builder.offset: Int
  get() = offset
  set(value) { offset = value }
fun ContainerFileSeekRequest.Builder.setWhence(value: SeekWhence): ContainerFileSeekRequest.Builder = apply { whence = value }
var ContainerFileSeekRequest.Builder.whence: SeekWhence
  get() = whence
  set(value) { whence = value }

fun ContainerFileWatchRequest.Companion.newBuilder(): ContainerFileWatchRequest.Builder = ContainerFileWatchRequest.Builder()
fun ContainerFileWatchRequest.Companion.getDefaultInstance(): ContainerFileWatchRequest = ContainerFileWatchRequest()
fun ContainerFileWatchRequest.toByteArray(): ByteArray = encode()
fun ContainerFileWatchRequest.Builder.setPath(value: String): ContainerFileWatchRequest.Builder = apply { path = value }
var ContainerFileWatchRequest.Builder.path: String
  get() = path
  set(value) { path = value }
fun ContainerFileWatchRequest.Builder.setRecursive(value: Boolean): ContainerFileWatchRequest.Builder = apply { recursive = value }
var ContainerFileWatchRequest.Builder.recursive: Boolean
  get() = recursive
  set(value) { recursive = value }
val ContainerFileWatchRequest.timeoutSecs: Long? get() = timeout_secs
fun ContainerFileWatchRequest.hasTimeoutSecs(): Boolean = timeout_secs != null
fun ContainerFileWatchRequest.Builder.setTimeoutSecs(value: Long?): ContainerFileWatchRequest.Builder = apply { timeout_secs = value }
var ContainerFileWatchRequest.Builder.timeoutSecs: Long?
  get() = timeout_secs
  set(value) { timeout_secs = value }

fun ContainerFileWriteReplaceBytesRequest.Companion.newBuilder(): ContainerFileWriteReplaceBytesRequest.Builder = ContainerFileWriteReplaceBytesRequest.Builder()
fun ContainerFileWriteReplaceBytesRequest.Companion.getDefaultInstance(): ContainerFileWriteReplaceBytesRequest = ContainerFileWriteReplaceBytesRequest()
fun ContainerFileWriteReplaceBytesRequest.toByteArray(): ByteArray = encode()
val ContainerFileWriteReplaceBytesRequest.fileDescriptor: String get() = file_descriptor
fun ContainerFileWriteReplaceBytesRequest.Builder.setFileDescriptor(value: String): ContainerFileWriteReplaceBytesRequest.Builder = apply { file_descriptor = value }
var ContainerFileWriteReplaceBytesRequest.Builder.fileDescriptor: String
  get() = file_descriptor
  set(value) { file_descriptor = value }
val ContainerFileWriteReplaceBytesRequest.data: ByteString get() = data_
fun ContainerFileWriteReplaceBytesRequest.Builder.setData(value: ByteString): ContainerFileWriteReplaceBytesRequest.Builder = apply { data_ = value }
var ContainerFileWriteReplaceBytesRequest.Builder.data: ByteString
  get() = data_
  set(value) { data_ = value }
fun ContainerFileWriteReplaceBytesRequest.Builder.setData(value: ProtoByteString): ContainerFileWriteReplaceBytesRequest.Builder = apply { data_ = value.toByteArray().toByteString() }
val ContainerFileWriteReplaceBytesRequest.startInclusive: Int? get() = start_inclusive
fun ContainerFileWriteReplaceBytesRequest.hasStartInclusive(): Boolean = start_inclusive != null
fun ContainerFileWriteReplaceBytesRequest.Builder.setStartInclusive(value: Int?): ContainerFileWriteReplaceBytesRequest.Builder = apply { start_inclusive = value }
var ContainerFileWriteReplaceBytesRequest.Builder.startInclusive: Int?
  get() = start_inclusive
  set(value) { start_inclusive = value }
val ContainerFileWriteReplaceBytesRequest.endExclusive: Int? get() = end_exclusive
fun ContainerFileWriteReplaceBytesRequest.hasEndExclusive(): Boolean = end_exclusive != null
fun ContainerFileWriteReplaceBytesRequest.Builder.setEndExclusive(value: Int?): ContainerFileWriteReplaceBytesRequest.Builder = apply { end_exclusive = value }
var ContainerFileWriteReplaceBytesRequest.Builder.endExclusive: Int?
  get() = end_exclusive
  set(value) { end_exclusive = value }

fun ContainerFileWriteRequest.Companion.newBuilder(): ContainerFileWriteRequest.Builder = ContainerFileWriteRequest.Builder()
fun ContainerFileWriteRequest.Companion.getDefaultInstance(): ContainerFileWriteRequest = ContainerFileWriteRequest()
fun ContainerFileWriteRequest.toByteArray(): ByteArray = encode()
val ContainerFileWriteRequest.fileDescriptor: String get() = file_descriptor
fun ContainerFileWriteRequest.Builder.setFileDescriptor(value: String): ContainerFileWriteRequest.Builder = apply { file_descriptor = value }
var ContainerFileWriteRequest.Builder.fileDescriptor: String
  get() = file_descriptor
  set(value) { file_descriptor = value }
val ContainerFileWriteRequest.data: ByteString get() = data_
fun ContainerFileWriteRequest.Builder.setData(value: ByteString): ContainerFileWriteRequest.Builder = apply { data_ = value }
var ContainerFileWriteRequest.Builder.data: ByteString
  get() = data_
  set(value) { data_ = value }
fun ContainerFileWriteRequest.Builder.setData(value: ProtoByteString): ContainerFileWriteRequest.Builder = apply { data_ = value.toByteArray().toByteString() }

fun ContainerFilesystemExecGetOutputRequest.Companion.newBuilder(): ContainerFilesystemExecGetOutputRequest.Builder = ContainerFilesystemExecGetOutputRequest.Builder()
fun ContainerFilesystemExecGetOutputRequest.Companion.getDefaultInstance(): ContainerFilesystemExecGetOutputRequest = ContainerFilesystemExecGetOutputRequest()
fun ContainerFilesystemExecGetOutputRequest.toByteArray(): ByteArray = encode()
val ContainerFilesystemExecGetOutputRequest.execId: String get() = exec_id
fun ContainerFilesystemExecGetOutputRequest.Builder.setExecId(value: String): ContainerFilesystemExecGetOutputRequest.Builder = apply { exec_id = value }
var ContainerFilesystemExecGetOutputRequest.Builder.execId: String
  get() = exec_id
  set(value) { exec_id = value }
fun ContainerFilesystemExecGetOutputRequest.Builder.setTimeout(value: Float): ContainerFilesystemExecGetOutputRequest.Builder = apply { timeout = value }
var ContainerFilesystemExecGetOutputRequest.Builder.timeout: Float
  get() = timeout
  set(value) { timeout = value }

fun ContainerFilesystemExecRequest.Companion.newBuilder(): ContainerFilesystemExecRequest.Builder = ContainerFilesystemExecRequest.Builder()
fun ContainerFilesystemExecRequest.Companion.getDefaultInstance(): ContainerFilesystemExecRequest = ContainerFilesystemExecRequest()
fun ContainerFilesystemExecRequest.toByteArray(): ByteArray = encode()
val ContainerFilesystemExecRequest.fileOpenRequest: ContainerFileOpenRequest? get() = file_open_request
fun ContainerFilesystemExecRequest.hasFileOpenRequest(): Boolean = file_open_request != null
fun ContainerFilesystemExecRequest.Builder.setFileOpenRequest(value: ContainerFileOpenRequest?): ContainerFilesystemExecRequest.Builder = apply { file_open_request = value }
var ContainerFilesystemExecRequest.Builder.fileOpenRequest: ContainerFileOpenRequest?
  get() = file_open_request
  set(value) { file_open_request = value }
val ContainerFilesystemExecRequest.fileWriteRequest: ContainerFileWriteRequest? get() = file_write_request
fun ContainerFilesystemExecRequest.hasFileWriteRequest(): Boolean = file_write_request != null
fun ContainerFilesystemExecRequest.Builder.setFileWriteRequest(value: ContainerFileWriteRequest?): ContainerFilesystemExecRequest.Builder = apply { file_write_request = value }
var ContainerFilesystemExecRequest.Builder.fileWriteRequest: ContainerFileWriteRequest?
  get() = file_write_request
  set(value) { file_write_request = value }
val ContainerFilesystemExecRequest.fileReadRequest: ContainerFileReadRequest? get() = file_read_request
fun ContainerFilesystemExecRequest.hasFileReadRequest(): Boolean = file_read_request != null
fun ContainerFilesystemExecRequest.Builder.setFileReadRequest(value: ContainerFileReadRequest?): ContainerFilesystemExecRequest.Builder = apply { file_read_request = value }
var ContainerFilesystemExecRequest.Builder.fileReadRequest: ContainerFileReadRequest?
  get() = file_read_request
  set(value) { file_read_request = value }
val ContainerFilesystemExecRequest.fileFlushRequest: ContainerFileFlushRequest? get() = file_flush_request
fun ContainerFilesystemExecRequest.hasFileFlushRequest(): Boolean = file_flush_request != null
fun ContainerFilesystemExecRequest.Builder.setFileFlushRequest(value: ContainerFileFlushRequest?): ContainerFilesystemExecRequest.Builder = apply { file_flush_request = value }
var ContainerFilesystemExecRequest.Builder.fileFlushRequest: ContainerFileFlushRequest?
  get() = file_flush_request
  set(value) { file_flush_request = value }
val ContainerFilesystemExecRequest.fileReadLineRequest: ContainerFileReadLineRequest? get() = file_read_line_request
fun ContainerFilesystemExecRequest.hasFileReadLineRequest(): Boolean = file_read_line_request != null
fun ContainerFilesystemExecRequest.Builder.setFileReadLineRequest(value: ContainerFileReadLineRequest?): ContainerFilesystemExecRequest.Builder = apply { file_read_line_request = value }
var ContainerFilesystemExecRequest.Builder.fileReadLineRequest: ContainerFileReadLineRequest?
  get() = file_read_line_request
  set(value) { file_read_line_request = value }
val ContainerFilesystemExecRequest.fileSeekRequest: ContainerFileSeekRequest? get() = file_seek_request
fun ContainerFilesystemExecRequest.hasFileSeekRequest(): Boolean = file_seek_request != null
fun ContainerFilesystemExecRequest.Builder.setFileSeekRequest(value: ContainerFileSeekRequest?): ContainerFilesystemExecRequest.Builder = apply { file_seek_request = value }
var ContainerFilesystemExecRequest.Builder.fileSeekRequest: ContainerFileSeekRequest?
  get() = file_seek_request
  set(value) { file_seek_request = value }
val ContainerFilesystemExecRequest.fileDeleteBytesRequest: ContainerFileDeleteBytesRequest? get() = file_delete_bytes_request
fun ContainerFilesystemExecRequest.hasFileDeleteBytesRequest(): Boolean = file_delete_bytes_request != null
fun ContainerFilesystemExecRequest.Builder.setFileDeleteBytesRequest(value: ContainerFileDeleteBytesRequest?): ContainerFilesystemExecRequest.Builder = apply { file_delete_bytes_request = value }
var ContainerFilesystemExecRequest.Builder.fileDeleteBytesRequest: ContainerFileDeleteBytesRequest?
  get() = file_delete_bytes_request
  set(value) { file_delete_bytes_request = value }
val ContainerFilesystemExecRequest.fileWriteReplaceBytesRequest: ContainerFileWriteReplaceBytesRequest? get() = file_write_replace_bytes_request
fun ContainerFilesystemExecRequest.hasFileWriteReplaceBytesRequest(): Boolean = file_write_replace_bytes_request != null
fun ContainerFilesystemExecRequest.Builder.setFileWriteReplaceBytesRequest(value: ContainerFileWriteReplaceBytesRequest?): ContainerFilesystemExecRequest.Builder = apply { file_write_replace_bytes_request = value }
var ContainerFilesystemExecRequest.Builder.fileWriteReplaceBytesRequest: ContainerFileWriteReplaceBytesRequest?
  get() = file_write_replace_bytes_request
  set(value) { file_write_replace_bytes_request = value }
val ContainerFilesystemExecRequest.fileCloseRequest: ContainerFileCloseRequest? get() = file_close_request
fun ContainerFilesystemExecRequest.hasFileCloseRequest(): Boolean = file_close_request != null
fun ContainerFilesystemExecRequest.Builder.setFileCloseRequest(value: ContainerFileCloseRequest?): ContainerFilesystemExecRequest.Builder = apply { file_close_request = value }
var ContainerFilesystemExecRequest.Builder.fileCloseRequest: ContainerFileCloseRequest?
  get() = file_close_request
  set(value) { file_close_request = value }
val ContainerFilesystemExecRequest.fileLsRequest: ContainerFileLsRequest? get() = file_ls_request
fun ContainerFilesystemExecRequest.hasFileLsRequest(): Boolean = file_ls_request != null
fun ContainerFilesystemExecRequest.Builder.setFileLsRequest(value: ContainerFileLsRequest?): ContainerFilesystemExecRequest.Builder = apply { file_ls_request = value }
var ContainerFilesystemExecRequest.Builder.fileLsRequest: ContainerFileLsRequest?
  get() = file_ls_request
  set(value) { file_ls_request = value }
val ContainerFilesystemExecRequest.fileMkdirRequest: ContainerFileMkdirRequest? get() = file_mkdir_request
fun ContainerFilesystemExecRequest.hasFileMkdirRequest(): Boolean = file_mkdir_request != null
fun ContainerFilesystemExecRequest.Builder.setFileMkdirRequest(value: ContainerFileMkdirRequest?): ContainerFilesystemExecRequest.Builder = apply { file_mkdir_request = value }
var ContainerFilesystemExecRequest.Builder.fileMkdirRequest: ContainerFileMkdirRequest?
  get() = file_mkdir_request
  set(value) { file_mkdir_request = value }
val ContainerFilesystemExecRequest.fileRmRequest: ContainerFileRmRequest? get() = file_rm_request
fun ContainerFilesystemExecRequest.hasFileRmRequest(): Boolean = file_rm_request != null
fun ContainerFilesystemExecRequest.Builder.setFileRmRequest(value: ContainerFileRmRequest?): ContainerFilesystemExecRequest.Builder = apply { file_rm_request = value }
var ContainerFilesystemExecRequest.Builder.fileRmRequest: ContainerFileRmRequest?
  get() = file_rm_request
  set(value) { file_rm_request = value }
val ContainerFilesystemExecRequest.fileWatchRequest: ContainerFileWatchRequest? get() = file_watch_request
fun ContainerFilesystemExecRequest.hasFileWatchRequest(): Boolean = file_watch_request != null
fun ContainerFilesystemExecRequest.Builder.setFileWatchRequest(value: ContainerFileWatchRequest?): ContainerFilesystemExecRequest.Builder = apply { file_watch_request = value }
var ContainerFilesystemExecRequest.Builder.fileWatchRequest: ContainerFileWatchRequest?
  get() = file_watch_request
  set(value) { file_watch_request = value }
val ContainerFilesystemExecRequest.taskId: String get() = task_id
fun ContainerFilesystemExecRequest.Builder.setTaskId(value: String): ContainerFilesystemExecRequest.Builder = apply { task_id = value }
var ContainerFilesystemExecRequest.Builder.taskId: String
  get() = task_id
  set(value) { task_id = value }

fun ContainerFilesystemExecResponse.Companion.newBuilder(): ContainerFilesystemExecResponse.Builder = ContainerFilesystemExecResponse.Builder()
fun ContainerFilesystemExecResponse.Companion.getDefaultInstance(): ContainerFilesystemExecResponse = ContainerFilesystemExecResponse()
fun ContainerFilesystemExecResponse.toByteArray(): ByteArray = encode()
val ContainerFilesystemExecResponse.execId: String get() = exec_id
fun ContainerFilesystemExecResponse.Builder.setExecId(value: String): ContainerFilesystemExecResponse.Builder = apply { exec_id = value }
var ContainerFilesystemExecResponse.Builder.execId: String
  get() = exec_id
  set(value) { exec_id = value }
val ContainerFilesystemExecResponse.fileDescriptor: String? get() = file_descriptor
fun ContainerFilesystemExecResponse.hasFileDescriptor(): Boolean = file_descriptor != null
fun ContainerFilesystemExecResponse.Builder.setFileDescriptor(value: String?): ContainerFilesystemExecResponse.Builder = apply { file_descriptor = value }
var ContainerFilesystemExecResponse.Builder.fileDescriptor: String?
  get() = file_descriptor
  set(value) { file_descriptor = value }

fun CreationInfo.Companion.newBuilder(): CreationInfo.Builder = CreationInfo.Builder()
fun CreationInfo.Companion.getDefaultInstance(): CreationInfo = CreationInfo()
fun CreationInfo.toByteArray(): ByteArray = encode()
val CreationInfo.createdAt: Double get() = created_at
fun CreationInfo.Builder.setCreatedAt(value: Double): CreationInfo.Builder = apply { created_at = value }
var CreationInfo.Builder.createdAt: Double
  get() = created_at
  set(value) { created_at = value }
val CreationInfo.createdBy: String get() = created_by
fun CreationInfo.Builder.setCreatedBy(value: String): CreationInfo.Builder = apply { created_by = value }
var CreationInfo.Builder.createdBy: String
  get() = created_by
  set(value) { created_by = value }

fun CustomDomainConfig.Companion.newBuilder(): CustomDomainConfig.Builder = CustomDomainConfig.Builder()
fun CustomDomainConfig.Companion.getDefaultInstance(): CustomDomainConfig = CustomDomainConfig()
fun CustomDomainConfig.toByteArray(): ByteArray = encode()
fun CustomDomainConfig.Builder.setName(value: String): CustomDomainConfig.Builder = apply { name = value }
var CustomDomainConfig.Builder.name: String
  get() = name
  set(value) { name = value }

fun CustomDomainInfo.Companion.newBuilder(): CustomDomainInfo.Builder = CustomDomainInfo.Builder()
fun CustomDomainInfo.Companion.getDefaultInstance(): CustomDomainInfo = CustomDomainInfo()
fun CustomDomainInfo.toByteArray(): ByteArray = encode()
fun CustomDomainInfo.Builder.setUrl(value: String): CustomDomainInfo.Builder = apply { url = value }
var CustomDomainInfo.Builder.url: String
  get() = url
  set(value) { url = value }

fun FilesystemRuntimeOutputBatch.Companion.newBuilder(): FilesystemRuntimeOutputBatch.Builder = FilesystemRuntimeOutputBatch.Builder()
fun FilesystemRuntimeOutputBatch.Companion.getDefaultInstance(): FilesystemRuntimeOutputBatch = FilesystemRuntimeOutputBatch()
fun FilesystemRuntimeOutputBatch.toByteArray(): ByteArray = encode()
fun FilesystemRuntimeOutputBatch.hasError(): Boolean = error != null
fun FilesystemRuntimeOutputBatch.Builder.setError(value: SystemErrorMessage?): FilesystemRuntimeOutputBatch.Builder = apply { error = value }
var FilesystemRuntimeOutputBatch.Builder.error: SystemErrorMessage?
  get() = error
  set(value) { error = value }
val FilesystemRuntimeOutputBatch.batchIndex: Long get() = batch_index
fun FilesystemRuntimeOutputBatch.Builder.setBatchIndex(value: Long): FilesystemRuntimeOutputBatch.Builder = apply { batch_index = value }
var FilesystemRuntimeOutputBatch.Builder.batchIndex: Long
  get() = batch_index
  set(value) { batch_index = value }
fun FilesystemRuntimeOutputBatch.Builder.setEof(value: Boolean): FilesystemRuntimeOutputBatch.Builder = apply { eof = value }
var FilesystemRuntimeOutputBatch.Builder.eof: Boolean
  get() = eof
  set(value) { eof = value }
val FilesystemRuntimeOutputBatch.outputList: List<ByteString> get() = output
val FilesystemRuntimeOutputBatch.outputCount: Int get() = output.size
fun FilesystemRuntimeOutputBatch.Builder.addOutput(value: ByteString): FilesystemRuntimeOutputBatch.Builder = apply { output = output + value }
fun FilesystemRuntimeOutputBatch.Builder.addAllOutput(items: Iterable<ByteString>): FilesystemRuntimeOutputBatch.Builder = apply { this.output = this.output + items }
fun FilesystemRuntimeOutputBatch.Builder.clearOutput(): FilesystemRuntimeOutputBatch.Builder = apply { output = emptyList() }

fun Function.Companion.newBuilder(): Function.Builder = Function.Builder()
fun Function.Companion.getDefaultInstance(): Function = Function()
fun Function.toByteArray(): ByteArray = encode()
val Function.moduleName: String get() = module_name
fun Function.Builder.setModuleName(value: String): Function.Builder = apply { module_name = value }
var Function.Builder.moduleName: String
  get() = module_name
  set(value) { module_name = value }
val Function.functionName: String get() = function_name
fun Function.Builder.setFunctionName(value: String): Function.Builder = apply { function_name = value }
var Function.Builder.functionName: String
  get() = function_name
  set(value) { function_name = value }
val Function.imageId: String get() = image_id
fun Function.Builder.setImageId(value: String): Function.Builder = apply { image_id = value }
var Function.Builder.imageId: String
  get() = image_id
  set(value) { image_id = value }
val Function.functionSerialized: ByteString get() = function_serialized
fun Function.Builder.setFunctionSerialized(value: ByteString): Function.Builder = apply { function_serialized = value }
var Function.Builder.functionSerialized: ByteString
  get() = function_serialized
  set(value) { function_serialized = value }
fun Function.Builder.setFunctionSerialized(value: ProtoByteString): Function.Builder = apply { function_serialized = value.toByteArray().toByteString() }
val Function.definitionType: Function.DefinitionType get() = definition_type
fun Function.Builder.setDefinitionType(value: Function.DefinitionType): Function.Builder = apply { definition_type = value }
var Function.Builder.definitionType: Function.DefinitionType
  get() = definition_type
  set(value) { definition_type = value }
val Function.functionType: Function.FunctionType get() = function_type
fun Function.Builder.setFunctionType(value: Function.FunctionType): Function.Builder = apply { function_type = value }
var Function.Builder.functionType: Function.FunctionType
  get() = function_type
  set(value) { function_type = value }
fun Function.hasResources(): Boolean = resources != null
fun Function.Builder.setResources(value: Resources?): Function.Builder = apply { resources = value }
var Function.Builder.resources: Resources?
  get() = resources
  set(value) { resources = value }
val Function.rateLimit: RateLimit? get() = rate_limit
fun Function.hasRateLimit(): Boolean = rate_limit != null
fun Function.Builder.setRateLimit(value: RateLimit?): Function.Builder = apply { rate_limit = value }
var Function.Builder.rateLimit: RateLimit?
  get() = rate_limit
  set(value) { rate_limit = value }
val Function.webhookConfig: WebhookConfig? get() = webhook_config
fun Function.hasWebhookConfig(): Boolean = webhook_config != null
fun Function.Builder.setWebhookConfig(value: WebhookConfig?): Function.Builder = apply { webhook_config = value }
var Function.Builder.webhookConfig: WebhookConfig?
  get() = webhook_config
  set(value) { webhook_config = value }
val Function.proxyId: String? get() = proxy_id
fun Function.hasProxyId(): Boolean = proxy_id != null
fun Function.Builder.setProxyId(value: String?): Function.Builder = apply { proxy_id = value }
var Function.Builder.proxyId: String?
  get() = proxy_id
  set(value) { proxy_id = value }
val Function.retryPolicy: FunctionRetryPolicy? get() = retry_policy
fun Function.hasRetryPolicy(): Boolean = retry_policy != null
fun Function.Builder.setRetryPolicy(value: FunctionRetryPolicy?): Function.Builder = apply { retry_policy = value }
var Function.Builder.retryPolicy: FunctionRetryPolicy?
  get() = retry_policy
  set(value) { retry_policy = value }
val Function.concurrencyLimit: Int get() = concurrency_limit
fun Function.Builder.setConcurrencyLimit(value: Int): Function.Builder = apply { concurrency_limit = value }
var Function.Builder.concurrencyLimit: Int
  get() = concurrency_limit
  set(value) { concurrency_limit = value }
val Function.timeoutSecs: Int get() = timeout_secs
fun Function.Builder.setTimeoutSecs(value: Int): Function.Builder = apply { timeout_secs = value }
var Function.Builder.timeoutSecs: Int
  get() = timeout_secs
  set(value) { timeout_secs = value }
val Function.ptyInfo: PTYInfo? get() = pty_info
fun Function.hasPtyInfo(): Boolean = pty_info != null
fun Function.Builder.setPtyInfo(value: PTYInfo?): Function.Builder = apply { pty_info = value }
var Function.Builder.ptyInfo: PTYInfo?
  get() = pty_info
  set(value) { pty_info = value }
val Function.classSerialized: ByteString get() = class_serialized
fun Function.Builder.setClassSerialized(value: ByteString): Function.Builder = apply { class_serialized = value }
var Function.Builder.classSerialized: ByteString
  get() = class_serialized
  set(value) { class_serialized = value }
fun Function.Builder.setClassSerialized(value: ProtoByteString): Function.Builder = apply { class_serialized = value.toByteArray().toByteString() }
val Function.taskIdleTimeoutSecs: Int get() = task_idle_timeout_secs
fun Function.Builder.setTaskIdleTimeoutSecs(value: Int): Function.Builder = apply { task_idle_timeout_secs = value }
var Function.Builder.taskIdleTimeoutSecs: Int
  get() = task_idle_timeout_secs
  set(value) { task_idle_timeout_secs = value }
val Function.cloudProvider: CloudProvider? get() = cloud_provider
fun Function.hasCloudProvider(): Boolean = cloud_provider != null
fun Function.Builder.setCloudProvider(value: CloudProvider?): Function.Builder = apply { cloud_provider = value }
var Function.Builder.cloudProvider: CloudProvider?
  get() = cloud_provider
  set(value) { cloud_provider = value }
val Function.warmPoolSize: Int get() = warm_pool_size
fun Function.Builder.setWarmPoolSize(value: Int): Function.Builder = apply { warm_pool_size = value }
var Function.Builder.warmPoolSize: Int
  get() = warm_pool_size
  set(value) { warm_pool_size = value }
val Function.webUrl: String get() = web_url
fun Function.Builder.setWebUrl(value: String): Function.Builder = apply { web_url = value }
var Function.Builder.webUrl: String
  get() = web_url
  set(value) { web_url = value }
val Function.webUrlInfo: WebUrlInfo? get() = web_url_info
fun Function.hasWebUrlInfo(): Boolean = web_url_info != null
fun Function.Builder.setWebUrlInfo(value: WebUrlInfo?): Function.Builder = apply { web_url_info = value }
var Function.Builder.webUrlInfo: WebUrlInfo?
  get() = web_url_info
  set(value) { web_url_info = value }
fun Function.Builder.setRuntime(value: String): Function.Builder = apply { runtime = value }
var Function.Builder.runtime: String
  get() = runtime
  set(value) { runtime = value }
val Function.appName: String get() = app_name
fun Function.Builder.setAppName(value: String): Function.Builder = apply { app_name = value }
var Function.Builder.appName: String
  get() = app_name
  set(value) { app_name = value }
val Function.maxConcurrentInputs: Int get() = max_concurrent_inputs
fun Function.Builder.setMaxConcurrentInputs(value: Int): Function.Builder = apply { max_concurrent_inputs = value }
var Function.Builder.maxConcurrentInputs: Int
  get() = max_concurrent_inputs
  set(value) { max_concurrent_inputs = value }
val Function.workerId: String get() = worker_id
fun Function.Builder.setWorkerId(value: String): Function.Builder = apply { worker_id = value }
var Function.Builder.workerId: String
  get() = worker_id
  set(value) { worker_id = value }
val Function.runtimeDebug: Boolean get() = runtime_debug
fun Function.Builder.setRuntimeDebug(value: Boolean): Function.Builder = apply { runtime_debug = value }
var Function.Builder.runtimeDebug: Boolean
  get() = runtime_debug
  set(value) { runtime_debug = value }
val Function.isBuilderFunction: Boolean get() = is_builder_function
fun Function.Builder.setIsBuilderFunction(value: Boolean): Function.Builder = apply { is_builder_function = value }
var Function.Builder.isBuilderFunction: Boolean
  get() = is_builder_function
  set(value) { is_builder_function = value }
val Function.isAutoSnapshot: Boolean get() = is_auto_snapshot
fun Function.Builder.setIsAutoSnapshot(value: Boolean): Function.Builder = apply { is_auto_snapshot = value }
var Function.Builder.isAutoSnapshot: Boolean
  get() = is_auto_snapshot
  set(value) { is_auto_snapshot = value }
val Function.isMethod: Boolean get() = is_method
fun Function.Builder.setIsMethod(value: Boolean): Function.Builder = apply { is_method = value }
var Function.Builder.isMethod: Boolean
  get() = is_method
  set(value) { is_method = value }
val Function.isCheckpointingFunction: Boolean get() = is_checkpointing_function
fun Function.Builder.setIsCheckpointingFunction(value: Boolean): Function.Builder = apply { is_checkpointing_function = value }
var Function.Builder.isCheckpointingFunction: Boolean
  get() = is_checkpointing_function
  set(value) { is_checkpointing_function = value }
val Function.checkpointingEnabled: Boolean get() = checkpointing_enabled
fun Function.Builder.setCheckpointingEnabled(value: Boolean): Function.Builder = apply { checkpointing_enabled = value }
var Function.Builder.checkpointingEnabled: Boolean
  get() = checkpointing_enabled
  set(value) { checkpointing_enabled = value }
fun Function.hasCheckpoint(): Boolean = checkpoint != null
fun Function.Builder.setCheckpoint(value: CheckpointInfo?): Function.Builder = apply { checkpoint = value }
var Function.Builder.checkpoint: CheckpointInfo?
  get() = checkpoint
  set(value) { checkpoint = value }
val Function.blockNetwork: Boolean get() = block_network
fun Function.Builder.setBlockNetwork(value: Boolean): Function.Builder = apply { block_network = value }
var Function.Builder.blockNetwork: Boolean
  get() = block_network
  set(value) { block_network = value }
val Function.maxInputs: Int get() = max_inputs
fun Function.Builder.setMaxInputs(value: Int): Function.Builder = apply { max_inputs = value }
var Function.Builder.maxInputs: Int
  get() = max_inputs
  set(value) { max_inputs = value }
val Function.schedulerPlacement: SchedulerPlacement? get() = scheduler_placement
fun Function.hasSchedulerPlacement(): Boolean = scheduler_placement != null
fun Function.Builder.setSchedulerPlacement(value: SchedulerPlacement?): Function.Builder = apply { scheduler_placement = value }
var Function.Builder.schedulerPlacement: SchedulerPlacement?
  get() = scheduler_placement
  set(value) { scheduler_placement = value }
val Function.isClass: Boolean get() = is_class
fun Function.Builder.setIsClass(value: Boolean): Function.Builder = apply { is_class = value }
var Function.Builder.isClass: Boolean
  get() = is_class
  set(value) { is_class = value }
val Function.useFunctionId: String get() = use_function_id
fun Function.Builder.setUseFunctionId(value: String): Function.Builder = apply { use_function_id = value }
var Function.Builder.useFunctionId: String
  get() = use_function_id
  set(value) { use_function_id = value }
val Function.useMethodName: String get() = use_method_name
fun Function.Builder.setUseMethodName(value: String): Function.Builder = apply { use_method_name = value }
var Function.Builder.useMethodName: String
  get() = use_method_name
  set(value) { use_method_name = value }
val Function.classParameterInfo: ClassParameterInfo? get() = class_parameter_info
fun Function.hasClassParameterInfo(): Boolean = class_parameter_info != null
fun Function.Builder.setClassParameterInfo(value: ClassParameterInfo?): Function.Builder = apply { class_parameter_info = value }
var Function.Builder.classParameterInfo: ClassParameterInfo?
  get() = class_parameter_info
  set(value) { class_parameter_info = value }
val Function.batchMaxSize: Int get() = batch_max_size
fun Function.Builder.setBatchMaxSize(value: Int): Function.Builder = apply { batch_max_size = value }
var Function.Builder.batchMaxSize: Int
  get() = batch_max_size
  set(value) { batch_max_size = value }
val Function.batchLingerMs: Long get() = batch_linger_ms
fun Function.Builder.setBatchLingerMs(value: Long): Function.Builder = apply { batch_linger_ms = value }
var Function.Builder.batchLingerMs: Long
  get() = batch_linger_ms
  set(value) { batch_linger_ms = value }
val Function.i6pnEnabled: Boolean get() = i6pn_enabled
fun Function.Builder.setI6pnEnabled(value: Boolean): Function.Builder = apply { i6pn_enabled = value }
var Function.Builder.i6pnEnabled: Boolean
  get() = i6pn_enabled
  set(value) { i6pn_enabled = value }
val Function.ExperimentalConcurrentCancellations: Boolean get() = _experimental_concurrent_cancellations
fun Function.Builder.setExperimentalConcurrentCancellations(value: Boolean): Function.Builder = apply { _experimental_concurrent_cancellations = value }
var Function.Builder.ExperimentalConcurrentCancellations: Boolean
  get() = _experimental_concurrent_cancellations
  set(value) { _experimental_concurrent_cancellations = value }
val Function.targetConcurrentInputs: Int get() = target_concurrent_inputs
fun Function.Builder.setTargetConcurrentInputs(value: Int): Function.Builder = apply { target_concurrent_inputs = value }
var Function.Builder.targetConcurrentInputs: Int
  get() = target_concurrent_inputs
  set(value) { target_concurrent_inputs = value }
val Function.ExperimentalTaskTemplatesEnabled: Boolean get() = _experimental_task_templates_enabled
fun Function.Builder.setExperimentalTaskTemplatesEnabled(value: Boolean): Function.Builder = apply { _experimental_task_templates_enabled = value }
var Function.Builder.ExperimentalTaskTemplatesEnabled: Boolean
  get() = _experimental_task_templates_enabled
  set(value) { _experimental_task_templates_enabled = value }
val Function.ExperimentalGroupSize: Int get() = _experimental_group_size
fun Function.Builder.setExperimentalGroupSize(value: Int): Function.Builder = apply { _experimental_group_size = value }
var Function.Builder.ExperimentalGroupSize: Int
  get() = _experimental_group_size
  set(value) { _experimental_group_size = value }
fun Function.Builder.setUntrusted(value: Boolean): Function.Builder = apply { untrusted = value }
var Function.Builder.untrusted: Boolean
  get() = untrusted
  set(value) { untrusted = value }
val Function.ExperimentalBufferContainers: Int get() = _experimental_buffer_containers
fun Function.Builder.setExperimentalBufferContainers(value: Int): Function.Builder = apply { _experimental_buffer_containers = value }
var Function.Builder.ExperimentalBufferContainers: Int
  get() = _experimental_buffer_containers
  set(value) { _experimental_buffer_containers = value }
val Function.ExperimentalProxyIp: String? get() = _experimental_proxy_ip
fun Function.hasExperimentalProxyIp(): Boolean = _experimental_proxy_ip != null
fun Function.Builder.setExperimentalProxyIp(value: String?): Function.Builder = apply { _experimental_proxy_ip = value }
var Function.Builder.ExperimentalProxyIp: String?
  get() = _experimental_proxy_ip
  set(value) { _experimental_proxy_ip = value }
val Function.runtimePerfRecord: Boolean get() = runtime_perf_record
fun Function.Builder.setRuntimePerfRecord(value: Boolean): Function.Builder = apply { runtime_perf_record = value }
var Function.Builder.runtimePerfRecord: Boolean
  get() = runtime_perf_record
  set(value) { runtime_perf_record = value }
fun Function.hasSchedule(): Boolean = schedule != null
fun Function.Builder.setSchedule(value: Schedule?): Function.Builder = apply { schedule = value }
var Function.Builder.schedule: Schedule?
  get() = schedule
  set(value) { schedule = value }
val Function.snapshotDebug: Boolean get() = snapshot_debug
fun Function.Builder.setSnapshotDebug(value: Boolean): Function.Builder = apply { snapshot_debug = value }
var Function.Builder.snapshotDebug: Boolean
  get() = snapshot_debug
  set(value) { snapshot_debug = value }
val Function.methodDefinitionsSet: Boolean get() = method_definitions_set
fun Function.Builder.setMethodDefinitionsSet(value: Boolean): Function.Builder = apply { method_definitions_set = value }
var Function.Builder.methodDefinitionsSet: Boolean
  get() = method_definitions_set
  set(value) { method_definitions_set = value }
val Function.ExperimentalCustomScaling: Boolean get() = _experimental_custom_scaling
fun Function.Builder.setExperimentalCustomScaling(value: Boolean): Function.Builder = apply { _experimental_custom_scaling = value }
var Function.Builder.ExperimentalCustomScaling: Boolean
  get() = _experimental_custom_scaling
  set(value) { _experimental_custom_scaling = value }
val Function.cloudProviderStr: String get() = cloud_provider_str
fun Function.Builder.setCloudProviderStr(value: String): Function.Builder = apply { cloud_provider_str = value }
var Function.Builder.cloudProviderStr: String
  get() = cloud_provider_str
  set(value) { cloud_provider_str = value }
val Function.ExperimentalEnableGpuSnapshot: Boolean get() = _experimental_enable_gpu_snapshot
fun Function.Builder.setExperimentalEnableGpuSnapshot(value: Boolean): Function.Builder = apply { _experimental_enable_gpu_snapshot = value }
var Function.Builder.ExperimentalEnableGpuSnapshot: Boolean
  get() = _experimental_enable_gpu_snapshot
  set(value) { _experimental_enable_gpu_snapshot = value }
val Function.autoscalerSettings: AutoscalerSettings? get() = autoscaler_settings
fun Function.hasAutoscalerSettings(): Boolean = autoscaler_settings != null
fun Function.Builder.setAutoscalerSettings(value: AutoscalerSettings?): Function.Builder = apply { autoscaler_settings = value }
var Function.Builder.autoscalerSettings: AutoscalerSettings?
  get() = autoscaler_settings
  set(value) { autoscaler_settings = value }
val Function.functionSchema: FunctionSchema? get() = function_schema
fun Function.hasFunctionSchema(): Boolean = function_schema != null
fun Function.Builder.setFunctionSchema(value: FunctionSchema?): Function.Builder = apply { function_schema = value }
var Function.Builder.functionSchema: FunctionSchema?
  get() = function_schema
  set(value) { function_schema = value }
val Function.mountClientDependencies: Boolean get() = mount_client_dependencies
fun Function.Builder.setMountClientDependencies(value: Boolean): Function.Builder = apply { mount_client_dependencies = value }
var Function.Builder.mountClientDependencies: Boolean
  get() = mount_client_dependencies
  set(value) { mount_client_dependencies = value }
val Function.flashServiceLabel: String get() = flash_service_label
fun Function.Builder.setFlashServiceLabel(value: String): Function.Builder = apply { flash_service_label = value }
var Function.Builder.flashServiceLabel: String
  get() = flash_service_label
  set(value) { flash_service_label = value }
val Function.enableGpuSnapshot: Boolean get() = enable_gpu_snapshot
fun Function.Builder.setEnableGpuSnapshot(value: Boolean): Function.Builder = apply { enable_gpu_snapshot = value }
var Function.Builder.enableGpuSnapshot: Boolean
  get() = enable_gpu_snapshot
  set(value) { enable_gpu_snapshot = value }
val Function.startupTimeoutSecs: Int get() = startup_timeout_secs
fun Function.Builder.setStartupTimeoutSecs(value: Int): Function.Builder = apply { startup_timeout_secs = value }
var Function.Builder.startupTimeoutSecs: Int
  get() = startup_timeout_secs
  set(value) { startup_timeout_secs = value }
val Function.httpConfig: HTTPConfig? get() = http_config
fun Function.hasHttpConfig(): Boolean = http_config != null
fun Function.Builder.setHttpConfig(value: HTTPConfig?): Function.Builder = apply { http_config = value }
var Function.Builder.httpConfig: HTTPConfig?
  get() = http_config
  set(value) { http_config = value }
val Function.implementationName: String get() = implementation_name
fun Function.Builder.setImplementationName(value: String): Function.Builder = apply { implementation_name = value }
var Function.Builder.implementationName: String
  get() = implementation_name
  set(value) { implementation_name = value }
val Function.singleUseContainers: Boolean get() = single_use_containers
fun Function.Builder.setSingleUseContainers(value: Boolean): Function.Builder = apply { single_use_containers = value }
var Function.Builder.singleUseContainers: Boolean
  get() = single_use_containers
  set(value) { single_use_containers = value }
val Function.isServer: Boolean get() = is_server
fun Function.Builder.setIsServer(value: Boolean): Function.Builder = apply { is_server = value }
var Function.Builder.isServer: Boolean
  get() = is_server
  set(value) { is_server = value }
val Function.mountIds: List<String> get() = mount_ids
val Function.mountIdsList: List<String> get() = mount_ids
val Function.mountIdsCount: Int get() = mount_ids.size
fun Function.Builder.addMountIds(value: String): Function.Builder = apply { mount_ids = mount_ids + value }
fun Function.Builder.addAllMountIds(items: Iterable<String>): Function.Builder = apply { this.mount_ids = this.mount_ids + items }
fun Function.Builder.clearMountIds(): Function.Builder = apply { mount_ids = emptyList() }
val Function.secretIds: List<String> get() = secret_ids
val Function.secretIdsList: List<String> get() = secret_ids
val Function.secretIdsCount: Int get() = secret_ids.size
fun Function.Builder.addSecretIds(value: String): Function.Builder = apply { secret_ids = secret_ids + value }
fun Function.Builder.addAllSecretIds(items: Iterable<String>): Function.Builder = apply { this.secret_ids = this.secret_ids + items }
fun Function.Builder.clearSecretIds(): Function.Builder = apply { secret_ids = emptyList() }
val Function.sharedVolumeMounts: List<SharedVolumeMount> get() = shared_volume_mounts
val Function.sharedVolumeMountsList: List<SharedVolumeMount> get() = shared_volume_mounts
val Function.sharedVolumeMountsCount: Int get() = shared_volume_mounts.size
fun Function.Builder.addSharedVolumeMounts(value: SharedVolumeMount): Function.Builder = apply { shared_volume_mounts = shared_volume_mounts + value }
fun Function.Builder.addAllSharedVolumeMounts(items: Iterable<SharedVolumeMount>): Function.Builder = apply { this.shared_volume_mounts = this.shared_volume_mounts + items }
fun Function.Builder.clearSharedVolumeMounts(): Function.Builder = apply { shared_volume_mounts = emptyList() }
val Function.volumeMounts: List<VolumeMount> get() = volume_mounts
val Function.volumeMountsList: List<VolumeMount> get() = volume_mounts
val Function.volumeMountsCount: Int get() = volume_mounts.size
fun Function.Builder.addVolumeMounts(value: VolumeMount): Function.Builder = apply { volume_mounts = volume_mounts + value }
fun Function.Builder.addAllVolumeMounts(items: Iterable<VolumeMount>): Function.Builder = apply { this.volume_mounts = this.volume_mounts + items }
fun Function.Builder.clearVolumeMounts(): Function.Builder = apply { volume_mounts = emptyList() }
val Function.customDomainInfo: List<CustomDomainInfo> get() = custom_domain_info
val Function.customDomainInfoList: List<CustomDomainInfo> get() = custom_domain_info
val Function.customDomainInfoCount: Int get() = custom_domain_info.size
fun Function.Builder.addCustomDomainInfo(value: CustomDomainInfo): Function.Builder = apply { custom_domain_info = custom_domain_info + value }
fun Function.Builder.addAllCustomDomainInfo(items: Iterable<CustomDomainInfo>): Function.Builder = apply { this.custom_domain_info = this.custom_domain_info + items }
fun Function.Builder.clearCustomDomainInfo(): Function.Builder = apply { custom_domain_info = emptyList() }
val Function.objectDependencies: List<ObjectDependency> get() = object_dependencies
val Function.objectDependenciesList: List<ObjectDependency> get() = object_dependencies
val Function.objectDependenciesCount: Int get() = object_dependencies.size
fun Function.Builder.addObjectDependencies(value: ObjectDependency): Function.Builder = apply { object_dependencies = object_dependencies + value }
fun Function.Builder.addAllObjectDependencies(items: Iterable<ObjectDependency>): Function.Builder = apply { this.object_dependencies = this.object_dependencies + items }
fun Function.Builder.clearObjectDependencies(): Function.Builder = apply { object_dependencies = emptyList() }
val Function.s3Mounts: List<S3Mount> get() = s3_mounts
val Function.s3MountsList: List<S3Mount> get() = s3_mounts
val Function.s3MountsCount: Int get() = s3_mounts.size
fun Function.Builder.addS3Mounts(value: S3Mount): Function.Builder = apply { s3_mounts = s3_mounts + value }
fun Function.Builder.addAllS3Mounts(items: Iterable<S3Mount>): Function.Builder = apply { this.s3_mounts = this.s3_mounts + items }
fun Function.Builder.clearS3Mounts(): Function.Builder = apply { s3_mounts = emptyList() }
val Function.cloudBucketMounts: List<CloudBucketMount> get() = cloud_bucket_mounts
val Function.cloudBucketMountsList: List<CloudBucketMount> get() = cloud_bucket_mounts
val Function.cloudBucketMountsCount: Int get() = cloud_bucket_mounts.size
fun Function.Builder.addCloudBucketMounts(value: CloudBucketMount): Function.Builder = apply { cloud_bucket_mounts = cloud_bucket_mounts + value }
fun Function.Builder.addAllCloudBucketMounts(items: Iterable<CloudBucketMount>): Function.Builder = apply { this.cloud_bucket_mounts = this.cloud_bucket_mounts + items }
fun Function.Builder.clearCloudBucketMounts(): Function.Builder = apply { cloud_bucket_mounts = emptyList() }
val Function.ExperimentalTaskTemplates: List<TaskTemplate> get() = _experimental_task_templates
val Function.ExperimentalTaskTemplatesList: List<TaskTemplate> get() = _experimental_task_templates
val Function.ExperimentalTaskTemplatesCount: Int get() = _experimental_task_templates.size
fun Function.Builder.addExperimentalTaskTemplates(value: TaskTemplate): Function.Builder = apply { _experimental_task_templates = _experimental_task_templates + value }
fun Function.Builder.addAllExperimentalTaskTemplates(items: Iterable<TaskTemplate>): Function.Builder = apply { this._experimental_task_templates = this._experimental_task_templates + items }
fun Function.Builder.clearExperimentalTaskTemplates(): Function.Builder = apply { _experimental_task_templates = emptyList() }
val Function.methodDefinitions: Map<String, MethodDefinition> get() = method_definitions
val Function.methodDefinitionsMap: Map<String, MethodDefinition> get() = method_definitions
fun Function.Builder.putMethodDefinitions(key: String, value: MethodDefinition): Function.Builder = apply { method_definitions = method_definitions + mapOf(key to value) }
fun Function.Builder.putAllMethodDefinitions(entries: Map<String, MethodDefinition>): Function.Builder = apply { method_definitions = method_definitions + entries }
fun Function.Builder.clearMethodDefinitions(): Function.Builder = apply { method_definitions = emptyMap() }
val Function.experimentalOptions: Map<String, String> get() = experimental_options
val Function.experimentalOptionsMap: Map<String, String> get() = experimental_options
fun Function.Builder.putExperimentalOptions(key: String, value: String): Function.Builder = apply { experimental_options = experimental_options + mapOf(key to value) }
fun Function.Builder.putAllExperimentalOptions(entries: Map<String, String>): Function.Builder = apply { experimental_options = experimental_options + entries }
fun Function.Builder.clearExperimentalOptions(): Function.Builder = apply { experimental_options = emptyMap() }
val Function.flashServiceUrls: List<String> get() = flash_service_urls
val Function.flashServiceUrlsList: List<String> get() = flash_service_urls
val Function.flashServiceUrlsCount: Int get() = flash_service_urls.size
fun Function.Builder.addFlashServiceUrls(value: String): Function.Builder = apply { flash_service_urls = flash_service_urls + value }
fun Function.Builder.addAllFlashServiceUrls(items: Iterable<String>): Function.Builder = apply { this.flash_service_urls = this.flash_service_urls + items }
fun Function.Builder.clearFlashServiceUrls(): Function.Builder = apply { flash_service_urls = emptyList() }
val Function.supportedInputFormats: List<DataFormat> get() = supported_input_formats
val Function.supportedInputFormatsList: List<DataFormat> get() = supported_input_formats
val Function.supportedInputFormatsCount: Int get() = supported_input_formats.size
fun Function.Builder.addSupportedInputFormats(value: DataFormat): Function.Builder = apply { supported_input_formats = supported_input_formats + value }
fun Function.Builder.addAllSupportedInputFormats(items: Iterable<DataFormat>): Function.Builder = apply { this.supported_input_formats = this.supported_input_formats + items }
fun Function.Builder.clearSupportedInputFormats(): Function.Builder = apply { supported_input_formats = emptyList() }
val Function.supportedOutputFormats: List<DataFormat> get() = supported_output_formats
val Function.supportedOutputFormatsList: List<DataFormat> get() = supported_output_formats
val Function.supportedOutputFormatsCount: Int get() = supported_output_formats.size
fun Function.Builder.addSupportedOutputFormats(value: DataFormat): Function.Builder = apply { supported_output_formats = supported_output_formats + value }
fun Function.Builder.addAllSupportedOutputFormats(items: Iterable<DataFormat>): Function.Builder = apply { this.supported_output_formats = this.supported_output_formats + items }
fun Function.Builder.clearSupportedOutputFormats(): Function.Builder = apply { supported_output_formats = emptyList() }

fun FunctionBindParamsRequest.Companion.newBuilder(): FunctionBindParamsRequest.Builder = FunctionBindParamsRequest.Builder()
fun FunctionBindParamsRequest.Companion.getDefaultInstance(): FunctionBindParamsRequest = FunctionBindParamsRequest()
fun FunctionBindParamsRequest.toByteArray(): ByteArray = encode()
val FunctionBindParamsRequest.functionId: String get() = function_id
fun FunctionBindParamsRequest.Builder.setFunctionId(value: String): FunctionBindParamsRequest.Builder = apply { function_id = value }
var FunctionBindParamsRequest.Builder.functionId: String
  get() = function_id
  set(value) { function_id = value }
val FunctionBindParamsRequest.serializedParams: ByteString get() = serialized_params
fun FunctionBindParamsRequest.Builder.setSerializedParams(value: ByteString): FunctionBindParamsRequest.Builder = apply { serialized_params = value }
var FunctionBindParamsRequest.Builder.serializedParams: ByteString
  get() = serialized_params
  set(value) { serialized_params = value }
fun FunctionBindParamsRequest.Builder.setSerializedParams(value: ProtoByteString): FunctionBindParamsRequest.Builder = apply { serialized_params = value.toByteArray().toByteString() }
val FunctionBindParamsRequest.functionOptions: FunctionOptions? get() = function_options
fun FunctionBindParamsRequest.hasFunctionOptions(): Boolean = function_options != null
fun FunctionBindParamsRequest.Builder.setFunctionOptions(value: FunctionOptions?): FunctionBindParamsRequest.Builder = apply { function_options = value }
var FunctionBindParamsRequest.Builder.functionOptions: FunctionOptions?
  get() = function_options
  set(value) { function_options = value }
val FunctionBindParamsRequest.environmentName: String get() = environment_name
fun FunctionBindParamsRequest.Builder.setEnvironmentName(value: String): FunctionBindParamsRequest.Builder = apply { environment_name = value }
var FunctionBindParamsRequest.Builder.environmentName: String
  get() = environment_name
  set(value) { environment_name = value }
val FunctionBindParamsRequest.authSecret: String get() = auth_secret
fun FunctionBindParamsRequest.Builder.setAuthSecret(value: String): FunctionBindParamsRequest.Builder = apply { auth_secret = value }
var FunctionBindParamsRequest.Builder.authSecret: String
  get() = auth_secret
  set(value) { auth_secret = value }

fun FunctionBindParamsResponse.Companion.newBuilder(): FunctionBindParamsResponse.Builder = FunctionBindParamsResponse.Builder()
fun FunctionBindParamsResponse.Companion.getDefaultInstance(): FunctionBindParamsResponse = FunctionBindParamsResponse()
fun FunctionBindParamsResponse.toByteArray(): ByteArray = encode()
val FunctionBindParamsResponse.boundFunctionId: String get() = bound_function_id
fun FunctionBindParamsResponse.Builder.setBoundFunctionId(value: String): FunctionBindParamsResponse.Builder = apply { bound_function_id = value }
var FunctionBindParamsResponse.Builder.boundFunctionId: String
  get() = bound_function_id
  set(value) { bound_function_id = value }
val FunctionBindParamsResponse.handleMetadata: FunctionHandleMetadata? get() = handle_metadata
fun FunctionBindParamsResponse.hasHandleMetadata(): Boolean = handle_metadata != null
fun FunctionBindParamsResponse.Builder.setHandleMetadata(value: FunctionHandleMetadata?): FunctionBindParamsResponse.Builder = apply { handle_metadata = value }
var FunctionBindParamsResponse.Builder.handleMetadata: FunctionHandleMetadata?
  get() = handle_metadata
  set(value) { handle_metadata = value }

fun FunctionCallCancelRequest.Companion.newBuilder(): FunctionCallCancelRequest.Builder = FunctionCallCancelRequest.Builder()
fun FunctionCallCancelRequest.Companion.getDefaultInstance(): FunctionCallCancelRequest = FunctionCallCancelRequest()
fun FunctionCallCancelRequest.toByteArray(): ByteArray = encode()
val FunctionCallCancelRequest.functionCallId: String get() = function_call_id
fun FunctionCallCancelRequest.Builder.setFunctionCallId(value: String): FunctionCallCancelRequest.Builder = apply { function_call_id = value }
var FunctionCallCancelRequest.Builder.functionCallId: String
  get() = function_call_id
  set(value) { function_call_id = value }
val FunctionCallCancelRequest.terminateContainers: Boolean get() = terminate_containers
fun FunctionCallCancelRequest.Builder.setTerminateContainers(value: Boolean): FunctionCallCancelRequest.Builder = apply { terminate_containers = value }
var FunctionCallCancelRequest.Builder.terminateContainers: Boolean
  get() = terminate_containers
  set(value) { terminate_containers = value }
val FunctionCallCancelRequest.functionId: String? get() = function_id
fun FunctionCallCancelRequest.hasFunctionId(): Boolean = function_id != null
fun FunctionCallCancelRequest.Builder.setFunctionId(value: String?): FunctionCallCancelRequest.Builder = apply { function_id = value }
var FunctionCallCancelRequest.Builder.functionId: String?
  get() = function_id
  set(value) { function_id = value }

fun FunctionGetCurrentStatsRequest.Companion.newBuilder(): FunctionGetCurrentStatsRequest.Builder = FunctionGetCurrentStatsRequest.Builder()
fun FunctionGetCurrentStatsRequest.Companion.getDefaultInstance(): FunctionGetCurrentStatsRequest = FunctionGetCurrentStatsRequest()
fun FunctionGetCurrentStatsRequest.toByteArray(): ByteArray = encode()
val FunctionGetCurrentStatsRequest.functionId: String get() = function_id
fun FunctionGetCurrentStatsRequest.Builder.setFunctionId(value: String): FunctionGetCurrentStatsRequest.Builder = apply { function_id = value }
var FunctionGetCurrentStatsRequest.Builder.functionId: String
  get() = function_id
  set(value) { function_id = value }

fun FunctionGetOutputsItem.Companion.newBuilder(): FunctionGetOutputsItem.Builder = FunctionGetOutputsItem.Builder()
fun FunctionGetOutputsItem.Companion.getDefaultInstance(): FunctionGetOutputsItem = FunctionGetOutputsItem()
fun FunctionGetOutputsItem.toByteArray(): ByteArray = encode()
fun FunctionGetOutputsItem.hasResult(): Boolean = result != null
fun FunctionGetOutputsItem.Builder.setResult(value: GenericResult?): FunctionGetOutputsItem.Builder = apply { result = value }
var FunctionGetOutputsItem.Builder.result: GenericResult?
  get() = result
  set(value) { result = value }
fun FunctionGetOutputsItem.Builder.setIdx(value: Int): FunctionGetOutputsItem.Builder = apply { idx = value }
var FunctionGetOutputsItem.Builder.idx: Int
  get() = idx
  set(value) { idx = value }
val FunctionGetOutputsItem.inputId: String get() = input_id
fun FunctionGetOutputsItem.Builder.setInputId(value: String): FunctionGetOutputsItem.Builder = apply { input_id = value }
var FunctionGetOutputsItem.Builder.inputId: String
  get() = input_id
  set(value) { input_id = value }
val FunctionGetOutputsItem.dataFormat: DataFormat get() = data_format
fun FunctionGetOutputsItem.Builder.setDataFormat(value: DataFormat): FunctionGetOutputsItem.Builder = apply { data_format = value }
var FunctionGetOutputsItem.Builder.dataFormat: DataFormat
  get() = data_format
  set(value) { data_format = value }
val FunctionGetOutputsItem.taskId: String get() = task_id
fun FunctionGetOutputsItem.Builder.setTaskId(value: String): FunctionGetOutputsItem.Builder = apply { task_id = value }
var FunctionGetOutputsItem.Builder.taskId: String
  get() = task_id
  set(value) { task_id = value }
val FunctionGetOutputsItem.inputStartedAt: Double get() = input_started_at
fun FunctionGetOutputsItem.Builder.setInputStartedAt(value: Double): FunctionGetOutputsItem.Builder = apply { input_started_at = value }
var FunctionGetOutputsItem.Builder.inputStartedAt: Double
  get() = input_started_at
  set(value) { input_started_at = value }
val FunctionGetOutputsItem.outputCreatedAt: Double get() = output_created_at
fun FunctionGetOutputsItem.Builder.setOutputCreatedAt(value: Double): FunctionGetOutputsItem.Builder = apply { output_created_at = value }
var FunctionGetOutputsItem.Builder.outputCreatedAt: Double
  get() = output_created_at
  set(value) { output_created_at = value }
val FunctionGetOutputsItem.retryCount: Int get() = retry_count
fun FunctionGetOutputsItem.Builder.setRetryCount(value: Int): FunctionGetOutputsItem.Builder = apply { retry_count = value }
var FunctionGetOutputsItem.Builder.retryCount: Int
  get() = retry_count
  set(value) { retry_count = value }
val FunctionGetOutputsItem.fcTraceTag: String get() = fc_trace_tag
fun FunctionGetOutputsItem.Builder.setFcTraceTag(value: String): FunctionGetOutputsItem.Builder = apply { fc_trace_tag = value }
var FunctionGetOutputsItem.Builder.fcTraceTag: String
  get() = fc_trace_tag
  set(value) { fc_trace_tag = value }

fun FunctionGetOutputsRequest.Companion.newBuilder(): FunctionGetOutputsRequest.Builder = FunctionGetOutputsRequest.Builder()
fun FunctionGetOutputsRequest.Companion.getDefaultInstance(): FunctionGetOutputsRequest = FunctionGetOutputsRequest()
fun FunctionGetOutputsRequest.toByteArray(): ByteArray = encode()
val FunctionGetOutputsRequest.functionCallId: String get() = function_call_id
fun FunctionGetOutputsRequest.Builder.setFunctionCallId(value: String): FunctionGetOutputsRequest.Builder = apply { function_call_id = value }
var FunctionGetOutputsRequest.Builder.functionCallId: String
  get() = function_call_id
  set(value) { function_call_id = value }
val FunctionGetOutputsRequest.maxValues: Int get() = max_values
fun FunctionGetOutputsRequest.Builder.setMaxValues(value: Int): FunctionGetOutputsRequest.Builder = apply { max_values = value }
var FunctionGetOutputsRequest.Builder.maxValues: Int
  get() = max_values
  set(value) { max_values = value }
fun FunctionGetOutputsRequest.Builder.setTimeout(value: Float): FunctionGetOutputsRequest.Builder = apply { timeout = value }
var FunctionGetOutputsRequest.Builder.timeout: Float
  get() = timeout
  set(value) { timeout = value }
val FunctionGetOutputsRequest.lastEntryId: String get() = last_entry_id
fun FunctionGetOutputsRequest.Builder.setLastEntryId(value: String): FunctionGetOutputsRequest.Builder = apply { last_entry_id = value }
var FunctionGetOutputsRequest.Builder.lastEntryId: String
  get() = last_entry_id
  set(value) { last_entry_id = value }
val FunctionGetOutputsRequest.clearOnSuccess: Boolean get() = clear_on_success
fun FunctionGetOutputsRequest.Builder.setClearOnSuccess(value: Boolean): FunctionGetOutputsRequest.Builder = apply { clear_on_success = value }
var FunctionGetOutputsRequest.Builder.clearOnSuccess: Boolean
  get() = clear_on_success
  set(value) { clear_on_success = value }
val FunctionGetOutputsRequest.requestedAt: Double get() = requested_at
fun FunctionGetOutputsRequest.Builder.setRequestedAt(value: Double): FunctionGetOutputsRequest.Builder = apply { requested_at = value }
var FunctionGetOutputsRequest.Builder.requestedAt: Double
  get() = requested_at
  set(value) { requested_at = value }
val FunctionGetOutputsRequest.startIdx: Int? get() = start_idx
fun FunctionGetOutputsRequest.hasStartIdx(): Boolean = start_idx != null
fun FunctionGetOutputsRequest.Builder.setStartIdx(value: Int?): FunctionGetOutputsRequest.Builder = apply { start_idx = value }
var FunctionGetOutputsRequest.Builder.startIdx: Int?
  get() = start_idx
  set(value) { start_idx = value }
val FunctionGetOutputsRequest.endIdx: Int? get() = end_idx
fun FunctionGetOutputsRequest.hasEndIdx(): Boolean = end_idx != null
fun FunctionGetOutputsRequest.Builder.setEndIdx(value: Int?): FunctionGetOutputsRequest.Builder = apply { end_idx = value }
var FunctionGetOutputsRequest.Builder.endIdx: Int?
  get() = end_idx
  set(value) { end_idx = value }
val FunctionGetOutputsRequest.inputJwts: List<String> get() = input_jwts
val FunctionGetOutputsRequest.inputJwtsList: List<String> get() = input_jwts
val FunctionGetOutputsRequest.inputJwtsCount: Int get() = input_jwts.size
fun FunctionGetOutputsRequest.Builder.addInputJwts(value: String): FunctionGetOutputsRequest.Builder = apply { input_jwts = input_jwts + value }
fun FunctionGetOutputsRequest.Builder.addAllInputJwts(items: Iterable<String>): FunctionGetOutputsRequest.Builder = apply { this.input_jwts = this.input_jwts + items }
fun FunctionGetOutputsRequest.Builder.clearInputJwts(): FunctionGetOutputsRequest.Builder = apply { input_jwts = emptyList() }

fun FunctionGetOutputsResponse.Companion.newBuilder(): FunctionGetOutputsResponse.Builder = FunctionGetOutputsResponse.Builder()
fun FunctionGetOutputsResponse.Companion.getDefaultInstance(): FunctionGetOutputsResponse = FunctionGetOutputsResponse()
fun FunctionGetOutputsResponse.toByteArray(): ByteArray = encode()
val FunctionGetOutputsResponse.lastEntryId: String get() = last_entry_id
fun FunctionGetOutputsResponse.Builder.setLastEntryId(value: String): FunctionGetOutputsResponse.Builder = apply { last_entry_id = value }
var FunctionGetOutputsResponse.Builder.lastEntryId: String
  get() = last_entry_id
  set(value) { last_entry_id = value }
val FunctionGetOutputsResponse.numUnfinishedInputs: Int get() = num_unfinished_inputs
fun FunctionGetOutputsResponse.Builder.setNumUnfinishedInputs(value: Int): FunctionGetOutputsResponse.Builder = apply { num_unfinished_inputs = value }
var FunctionGetOutputsResponse.Builder.numUnfinishedInputs: Int
  get() = num_unfinished_inputs
  set(value) { num_unfinished_inputs = value }
val FunctionGetOutputsResponse.idxsList: List<Int> get() = idxs
val FunctionGetOutputsResponse.idxsCount: Int get() = idxs.size
fun FunctionGetOutputsResponse.Builder.addIdxs(value: Int): FunctionGetOutputsResponse.Builder = apply { idxs = idxs + value }
fun FunctionGetOutputsResponse.Builder.addAllIdxs(items: Iterable<Int>): FunctionGetOutputsResponse.Builder = apply { this.idxs = this.idxs + items }
fun FunctionGetOutputsResponse.Builder.clearIdxs(): FunctionGetOutputsResponse.Builder = apply { idxs = emptyList() }
val FunctionGetOutputsResponse.outputsList: List<FunctionGetOutputsItem> get() = outputs
val FunctionGetOutputsResponse.outputsCount: Int get() = outputs.size
fun FunctionGetOutputsResponse.Builder.addOutputs(value: FunctionGetOutputsItem): FunctionGetOutputsResponse.Builder = apply { outputs = outputs + value }
fun FunctionGetOutputsResponse.Builder.addAllOutputs(items: Iterable<FunctionGetOutputsItem>): FunctionGetOutputsResponse.Builder = apply { this.outputs = this.outputs + items }
fun FunctionGetOutputsResponse.Builder.clearOutputs(): FunctionGetOutputsResponse.Builder = apply { outputs = emptyList() }

fun FunctionGetRequest.Companion.newBuilder(): FunctionGetRequest.Builder = FunctionGetRequest.Builder()
fun FunctionGetRequest.Companion.getDefaultInstance(): FunctionGetRequest = FunctionGetRequest()
fun FunctionGetRequest.toByteArray(): ByteArray = encode()
val FunctionGetRequest.appName: String get() = app_name
fun FunctionGetRequest.Builder.setAppName(value: String): FunctionGetRequest.Builder = apply { app_name = value }
var FunctionGetRequest.Builder.appName: String
  get() = app_name
  set(value) { app_name = value }
val FunctionGetRequest.objectTag: String get() = object_tag
fun FunctionGetRequest.Builder.setObjectTag(value: String): FunctionGetRequest.Builder = apply { object_tag = value }
var FunctionGetRequest.Builder.objectTag: String
  get() = object_tag
  set(value) { object_tag = value }
val FunctionGetRequest.environmentName: String get() = environment_name
fun FunctionGetRequest.Builder.setEnvironmentName(value: String): FunctionGetRequest.Builder = apply { environment_name = value }
var FunctionGetRequest.Builder.environmentName: String
  get() = environment_name
  set(value) { environment_name = value }

fun FunctionGetResponse.Companion.newBuilder(): FunctionGetResponse.Builder = FunctionGetResponse.Builder()
fun FunctionGetResponse.Companion.getDefaultInstance(): FunctionGetResponse = FunctionGetResponse()
fun FunctionGetResponse.toByteArray(): ByteArray = encode()
val FunctionGetResponse.functionId: String get() = function_id
fun FunctionGetResponse.Builder.setFunctionId(value: String): FunctionGetResponse.Builder = apply { function_id = value }
var FunctionGetResponse.Builder.functionId: String
  get() = function_id
  set(value) { function_id = value }
val FunctionGetResponse.handleMetadata: FunctionHandleMetadata? get() = handle_metadata
fun FunctionGetResponse.hasHandleMetadata(): Boolean = handle_metadata != null
fun FunctionGetResponse.Builder.setHandleMetadata(value: FunctionHandleMetadata?): FunctionGetResponse.Builder = apply { handle_metadata = value }
var FunctionGetResponse.Builder.handleMetadata: FunctionHandleMetadata?
  get() = handle_metadata
  set(value) { handle_metadata = value }
val FunctionGetResponse.serverWarnings: List<Warning> get() = server_warnings
val FunctionGetResponse.serverWarningsList: List<Warning> get() = server_warnings
val FunctionGetResponse.serverWarningsCount: Int get() = server_warnings.size
fun FunctionGetResponse.Builder.addServerWarnings(value: Warning): FunctionGetResponse.Builder = apply { server_warnings = server_warnings + value }
fun FunctionGetResponse.Builder.addAllServerWarnings(items: Iterable<Warning>): FunctionGetResponse.Builder = apply { this.server_warnings = this.server_warnings + items }
fun FunctionGetResponse.Builder.clearServerWarnings(): FunctionGetResponse.Builder = apply { server_warnings = emptyList() }

fun FunctionHandleMetadata.Companion.newBuilder(): FunctionHandleMetadata.Builder = FunctionHandleMetadata.Builder()
fun FunctionHandleMetadata.Companion.getDefaultInstance(): FunctionHandleMetadata = FunctionHandleMetadata()
fun FunctionHandleMetadata.toByteArray(): ByteArray = encode()
val FunctionHandleMetadata.functionName: String get() = function_name
fun FunctionHandleMetadata.Builder.setFunctionName(value: String): FunctionHandleMetadata.Builder = apply { function_name = value }
var FunctionHandleMetadata.Builder.functionName: String
  get() = function_name
  set(value) { function_name = value }
val FunctionHandleMetadata.functionType: Function.FunctionType get() = function_type
fun FunctionHandleMetadata.Builder.setFunctionType(value: Function.FunctionType): FunctionHandleMetadata.Builder = apply { function_type = value }
var FunctionHandleMetadata.Builder.functionType: Function.FunctionType
  get() = function_type
  set(value) { function_type = value }
val FunctionHandleMetadata.webUrl: String get() = web_url
fun FunctionHandleMetadata.Builder.setWebUrl(value: String): FunctionHandleMetadata.Builder = apply { web_url = value }
var FunctionHandleMetadata.Builder.webUrl: String
  get() = web_url
  set(value) { web_url = value }
val FunctionHandleMetadata.isMethod: Boolean get() = is_method
fun FunctionHandleMetadata.Builder.setIsMethod(value: Boolean): FunctionHandleMetadata.Builder = apply { is_method = value }
var FunctionHandleMetadata.Builder.isMethod: Boolean
  get() = is_method
  set(value) { is_method = value }
val FunctionHandleMetadata.useFunctionId: String get() = use_function_id
fun FunctionHandleMetadata.Builder.setUseFunctionId(value: String): FunctionHandleMetadata.Builder = apply { use_function_id = value }
var FunctionHandleMetadata.Builder.useFunctionId: String
  get() = use_function_id
  set(value) { use_function_id = value }
val FunctionHandleMetadata.useMethodName: String get() = use_method_name
fun FunctionHandleMetadata.Builder.setUseMethodName(value: String): FunctionHandleMetadata.Builder = apply { use_method_name = value }
var FunctionHandleMetadata.Builder.useMethodName: String
  get() = use_method_name
  set(value) { use_method_name = value }
val FunctionHandleMetadata.definitionId: String get() = definition_id
fun FunctionHandleMetadata.Builder.setDefinitionId(value: String): FunctionHandleMetadata.Builder = apply { definition_id = value }
var FunctionHandleMetadata.Builder.definitionId: String
  get() = definition_id
  set(value) { definition_id = value }
val FunctionHandleMetadata.classParameterInfo: ClassParameterInfo? get() = class_parameter_info
fun FunctionHandleMetadata.hasClassParameterInfo(): Boolean = class_parameter_info != null
fun FunctionHandleMetadata.Builder.setClassParameterInfo(value: ClassParameterInfo?): FunctionHandleMetadata.Builder = apply { class_parameter_info = value }
var FunctionHandleMetadata.Builder.classParameterInfo: ClassParameterInfo?
  get() = class_parameter_info
  set(value) { class_parameter_info = value }
val FunctionHandleMetadata.functionSchema: FunctionSchema? get() = function_schema
fun FunctionHandleMetadata.hasFunctionSchema(): Boolean = function_schema != null
fun FunctionHandleMetadata.Builder.setFunctionSchema(value: FunctionSchema?): FunctionHandleMetadata.Builder = apply { function_schema = value }
var FunctionHandleMetadata.Builder.functionSchema: FunctionSchema?
  get() = function_schema
  set(value) { function_schema = value }
val FunctionHandleMetadata.inputPlaneUrl: String? get() = input_plane_url
fun FunctionHandleMetadata.hasInputPlaneUrl(): Boolean = input_plane_url != null
fun FunctionHandleMetadata.Builder.setInputPlaneUrl(value: String?): FunctionHandleMetadata.Builder = apply { input_plane_url = value }
var FunctionHandleMetadata.Builder.inputPlaneUrl: String?
  get() = input_plane_url
  set(value) { input_plane_url = value }
val FunctionHandleMetadata.inputPlaneRegion: String? get() = input_plane_region
fun FunctionHandleMetadata.hasInputPlaneRegion(): Boolean = input_plane_region != null
fun FunctionHandleMetadata.Builder.setInputPlaneRegion(value: String?): FunctionHandleMetadata.Builder = apply { input_plane_region = value }
var FunctionHandleMetadata.Builder.inputPlaneRegion: String?
  get() = input_plane_region
  set(value) { input_plane_region = value }
val FunctionHandleMetadata.maxObjectSizeBytes: Long? get() = max_object_size_bytes
fun FunctionHandleMetadata.hasMaxObjectSizeBytes(): Boolean = max_object_size_bytes != null
fun FunctionHandleMetadata.Builder.setMaxObjectSizeBytes(value: Long?): FunctionHandleMetadata.Builder = apply { max_object_size_bytes = value }
var FunctionHandleMetadata.Builder.maxObjectSizeBytes: Long?
  get() = max_object_size_bytes
  set(value) { max_object_size_bytes = value }
val FunctionHandleMetadata.methodHandleMetadata: Map<String, FunctionHandleMetadata> get() = method_handle_metadata
val FunctionHandleMetadata.methodHandleMetadataMap: Map<String, FunctionHandleMetadata> get() = method_handle_metadata
fun FunctionHandleMetadata.Builder.putMethodHandleMetadata(key: String, value: FunctionHandleMetadata): FunctionHandleMetadata.Builder = apply { method_handle_metadata = method_handle_metadata + mapOf(key to value) }
fun FunctionHandleMetadata.Builder.putAllMethodHandleMetadata(entries: Map<String, FunctionHandleMetadata>): FunctionHandleMetadata.Builder = apply { method_handle_metadata = method_handle_metadata + entries }
fun FunctionHandleMetadata.Builder.clearMethodHandleMetadata(): FunctionHandleMetadata.Builder = apply { method_handle_metadata = emptyMap() }
val FunctionHandleMetadata.ExperimentalFlashUrls: List<String> get() = _experimental_flash_urls
val FunctionHandleMetadata.ExperimentalFlashUrlsList: List<String> get() = _experimental_flash_urls
val FunctionHandleMetadata.ExperimentalFlashUrlsCount: Int get() = _experimental_flash_urls.size
fun FunctionHandleMetadata.Builder.addExperimentalFlashUrls(value: String): FunctionHandleMetadata.Builder = apply { _experimental_flash_urls = _experimental_flash_urls + value }
fun FunctionHandleMetadata.Builder.addAllExperimentalFlashUrls(items: Iterable<String>): FunctionHandleMetadata.Builder = apply { this._experimental_flash_urls = this._experimental_flash_urls + items }
fun FunctionHandleMetadata.Builder.clearExperimentalFlashUrls(): FunctionHandleMetadata.Builder = apply { _experimental_flash_urls = emptyList() }
val FunctionHandleMetadata.supportedInputFormats: List<DataFormat> get() = supported_input_formats
val FunctionHandleMetadata.supportedInputFormatsList: List<DataFormat> get() = supported_input_formats
val FunctionHandleMetadata.supportedInputFormatsCount: Int get() = supported_input_formats.size
fun FunctionHandleMetadata.Builder.addSupportedInputFormats(value: DataFormat): FunctionHandleMetadata.Builder = apply { supported_input_formats = supported_input_formats + value }
fun FunctionHandleMetadata.Builder.addAllSupportedInputFormats(items: Iterable<DataFormat>): FunctionHandleMetadata.Builder = apply { this.supported_input_formats = this.supported_input_formats + items }
fun FunctionHandleMetadata.Builder.clearSupportedInputFormats(): FunctionHandleMetadata.Builder = apply { supported_input_formats = emptyList() }
val FunctionHandleMetadata.supportedOutputFormats: List<DataFormat> get() = supported_output_formats
val FunctionHandleMetadata.supportedOutputFormatsList: List<DataFormat> get() = supported_output_formats
val FunctionHandleMetadata.supportedOutputFormatsCount: Int get() = supported_output_formats.size
fun FunctionHandleMetadata.Builder.addSupportedOutputFormats(value: DataFormat): FunctionHandleMetadata.Builder = apply { supported_output_formats = supported_output_formats + value }
fun FunctionHandleMetadata.Builder.addAllSupportedOutputFormats(items: Iterable<DataFormat>): FunctionHandleMetadata.Builder = apply { this.supported_output_formats = this.supported_output_formats + items }
fun FunctionHandleMetadata.Builder.clearSupportedOutputFormats(): FunctionHandleMetadata.Builder = apply { supported_output_formats = emptyList() }

fun FunctionInput.Companion.newBuilder(): FunctionInput.Builder = FunctionInput.Builder()
fun FunctionInput.Companion.getDefaultInstance(): FunctionInput = FunctionInput()
fun FunctionInput.toByteArray(): ByteArray = encode()
fun FunctionInput.hasArgs(): Boolean = args != null
fun FunctionInput.Builder.setArgs(value: ByteString?): FunctionInput.Builder = apply { args = value }
var FunctionInput.Builder.args: ByteString?
  get() = args
  set(value) { args = value }
fun FunctionInput.Builder.setArgs(value: ProtoByteString?): FunctionInput.Builder = apply { args = value?.toByteArray()?.toByteString() }
val FunctionInput.argsBlobId: String? get() = args_blob_id
fun FunctionInput.hasArgsBlobId(): Boolean = args_blob_id != null
fun FunctionInput.Builder.setArgsBlobId(value: String?): FunctionInput.Builder = apply { args_blob_id = value }
var FunctionInput.Builder.argsBlobId: String?
  get() = args_blob_id
  set(value) { args_blob_id = value }
val FunctionInput.finalInput: Boolean get() = final_input
fun FunctionInput.Builder.setFinalInput(value: Boolean): FunctionInput.Builder = apply { final_input = value }
var FunctionInput.Builder.finalInput: Boolean
  get() = final_input
  set(value) { final_input = value }
val FunctionInput.dataFormat: DataFormat get() = data_format
fun FunctionInput.Builder.setDataFormat(value: DataFormat): FunctionInput.Builder = apply { data_format = value }
var FunctionInput.Builder.dataFormat: DataFormat
  get() = data_format
  set(value) { data_format = value }
val FunctionInput.methodName: String? get() = method_name
fun FunctionInput.hasMethodName(): Boolean = method_name != null
fun FunctionInput.Builder.setMethodName(value: String?): FunctionInput.Builder = apply { method_name = value }
var FunctionInput.Builder.methodName: String?
  get() = method_name
  set(value) { method_name = value }

fun FunctionMapRequest.Companion.newBuilder(): FunctionMapRequest.Builder = FunctionMapRequest.Builder()
fun FunctionMapRequest.Companion.getDefaultInstance(): FunctionMapRequest = FunctionMapRequest()
fun FunctionMapRequest.toByteArray(): ByteArray = encode()
val FunctionMapRequest.functionId: String get() = function_id
fun FunctionMapRequest.Builder.setFunctionId(value: String): FunctionMapRequest.Builder = apply { function_id = value }
var FunctionMapRequest.Builder.functionId: String
  get() = function_id
  set(value) { function_id = value }
val FunctionMapRequest.parentInputId: String get() = parent_input_id
fun FunctionMapRequest.Builder.setParentInputId(value: String): FunctionMapRequest.Builder = apply { parent_input_id = value }
var FunctionMapRequest.Builder.parentInputId: String
  get() = parent_input_id
  set(value) { parent_input_id = value }
val FunctionMapRequest.returnExceptions: Boolean get() = return_exceptions
fun FunctionMapRequest.Builder.setReturnExceptions(value: Boolean): FunctionMapRequest.Builder = apply { return_exceptions = value }
var FunctionMapRequest.Builder.returnExceptions: Boolean
  get() = return_exceptions
  set(value) { return_exceptions = value }
val FunctionMapRequest.functionCallType: FunctionCallType get() = function_call_type
fun FunctionMapRequest.Builder.setFunctionCallType(value: FunctionCallType): FunctionMapRequest.Builder = apply { function_call_type = value }
var FunctionMapRequest.Builder.functionCallType: FunctionCallType
  get() = function_call_type
  set(value) { function_call_type = value }
val FunctionMapRequest.functionCallInvocationType: FunctionCallInvocationType get() = function_call_invocation_type
fun FunctionMapRequest.Builder.setFunctionCallInvocationType(value: FunctionCallInvocationType): FunctionMapRequest.Builder = apply { function_call_invocation_type = value }
var FunctionMapRequest.Builder.functionCallInvocationType: FunctionCallInvocationType
  get() = function_call_invocation_type
  set(value) { function_call_invocation_type = value }
val FunctionMapRequest.fromSpawnMap: Boolean get() = from_spawn_map
fun FunctionMapRequest.Builder.setFromSpawnMap(value: Boolean): FunctionMapRequest.Builder = apply { from_spawn_map = value }
var FunctionMapRequest.Builder.fromSpawnMap: Boolean
  get() = from_spawn_map
  set(value) { from_spawn_map = value }
val FunctionMapRequest.pipelinedInputs: List<FunctionPutInputsItem> get() = pipelined_inputs
val FunctionMapRequest.pipelinedInputsList: List<FunctionPutInputsItem> get() = pipelined_inputs
val FunctionMapRequest.pipelinedInputsCount: Int get() = pipelined_inputs.size
fun FunctionMapRequest.Builder.addPipelinedInputs(value: FunctionPutInputsItem): FunctionMapRequest.Builder = apply { pipelined_inputs = pipelined_inputs + value }
fun FunctionMapRequest.Builder.addAllPipelinedInputs(items: Iterable<FunctionPutInputsItem>): FunctionMapRequest.Builder = apply { this.pipelined_inputs = this.pipelined_inputs + items }
fun FunctionMapRequest.Builder.clearPipelinedInputs(): FunctionMapRequest.Builder = apply { pipelined_inputs = emptyList() }

fun FunctionMapResponse.Companion.newBuilder(): FunctionMapResponse.Builder = FunctionMapResponse.Builder()
fun FunctionMapResponse.Companion.getDefaultInstance(): FunctionMapResponse = FunctionMapResponse()
fun FunctionMapResponse.toByteArray(): ByteArray = encode()
val FunctionMapResponse.functionCallId: String get() = function_call_id
fun FunctionMapResponse.Builder.setFunctionCallId(value: String): FunctionMapResponse.Builder = apply { function_call_id = value }
var FunctionMapResponse.Builder.functionCallId: String
  get() = function_call_id
  set(value) { function_call_id = value }
val FunctionMapResponse.retryPolicy: FunctionRetryPolicy? get() = retry_policy
fun FunctionMapResponse.hasRetryPolicy(): Boolean = retry_policy != null
fun FunctionMapResponse.Builder.setRetryPolicy(value: FunctionRetryPolicy?): FunctionMapResponse.Builder = apply { retry_policy = value }
var FunctionMapResponse.Builder.retryPolicy: FunctionRetryPolicy?
  get() = retry_policy
  set(value) { retry_policy = value }
val FunctionMapResponse.functionCallJwt: String get() = function_call_jwt
fun FunctionMapResponse.Builder.setFunctionCallJwt(value: String): FunctionMapResponse.Builder = apply { function_call_jwt = value }
var FunctionMapResponse.Builder.functionCallJwt: String
  get() = function_call_jwt
  set(value) { function_call_jwt = value }
val FunctionMapResponse.syncClientRetriesEnabled: Boolean get() = sync_client_retries_enabled
fun FunctionMapResponse.Builder.setSyncClientRetriesEnabled(value: Boolean): FunctionMapResponse.Builder = apply { sync_client_retries_enabled = value }
var FunctionMapResponse.Builder.syncClientRetriesEnabled: Boolean
  get() = sync_client_retries_enabled
  set(value) { sync_client_retries_enabled = value }
val FunctionMapResponse.maxInputsOutstanding: Int get() = max_inputs_outstanding
fun FunctionMapResponse.Builder.setMaxInputsOutstanding(value: Int): FunctionMapResponse.Builder = apply { max_inputs_outstanding = value }
var FunctionMapResponse.Builder.maxInputsOutstanding: Int
  get() = max_inputs_outstanding
  set(value) { max_inputs_outstanding = value }
val FunctionMapResponse.pipelinedInputs: List<FunctionPutInputsResponseItem> get() = pipelined_inputs
val FunctionMapResponse.pipelinedInputsList: List<FunctionPutInputsResponseItem> get() = pipelined_inputs
val FunctionMapResponse.pipelinedInputsCount: Int get() = pipelined_inputs.size
fun FunctionMapResponse.Builder.addPipelinedInputs(value: FunctionPutInputsResponseItem): FunctionMapResponse.Builder = apply { pipelined_inputs = pipelined_inputs + value }
fun FunctionMapResponse.Builder.addAllPipelinedInputs(items: Iterable<FunctionPutInputsResponseItem>): FunctionMapResponse.Builder = apply { this.pipelined_inputs = this.pipelined_inputs + items }
fun FunctionMapResponse.Builder.clearPipelinedInputs(): FunctionMapResponse.Builder = apply { pipelined_inputs = emptyList() }

fun FunctionOptions.Companion.newBuilder(): FunctionOptions.Builder = FunctionOptions.Builder()
fun FunctionOptions.Companion.getDefaultInstance(): FunctionOptions = FunctionOptions()
fun FunctionOptions.toByteArray(): ByteArray = encode()
fun FunctionOptions.hasResources(): Boolean = resources != null
fun FunctionOptions.Builder.setResources(value: Resources?): FunctionOptions.Builder = apply { resources = value }
var FunctionOptions.Builder.resources: Resources?
  get() = resources
  set(value) { resources = value }
val FunctionOptions.retryPolicy: FunctionRetryPolicy? get() = retry_policy
fun FunctionOptions.hasRetryPolicy(): Boolean = retry_policy != null
fun FunctionOptions.Builder.setRetryPolicy(value: FunctionRetryPolicy?): FunctionOptions.Builder = apply { retry_policy = value }
var FunctionOptions.Builder.retryPolicy: FunctionRetryPolicy?
  get() = retry_policy
  set(value) { retry_policy = value }
val FunctionOptions.concurrencyLimit: Int? get() = concurrency_limit
fun FunctionOptions.hasConcurrencyLimit(): Boolean = concurrency_limit != null
fun FunctionOptions.Builder.setConcurrencyLimit(value: Int?): FunctionOptions.Builder = apply { concurrency_limit = value }
var FunctionOptions.Builder.concurrencyLimit: Int?
  get() = concurrency_limit
  set(value) { concurrency_limit = value }
val FunctionOptions.timeoutSecs: Int? get() = timeout_secs
fun FunctionOptions.hasTimeoutSecs(): Boolean = timeout_secs != null
fun FunctionOptions.Builder.setTimeoutSecs(value: Int?): FunctionOptions.Builder = apply { timeout_secs = value }
var FunctionOptions.Builder.timeoutSecs: Int?
  get() = timeout_secs
  set(value) { timeout_secs = value }
val FunctionOptions.taskIdleTimeoutSecs: Int? get() = task_idle_timeout_secs
fun FunctionOptions.hasTaskIdleTimeoutSecs(): Boolean = task_idle_timeout_secs != null
fun FunctionOptions.Builder.setTaskIdleTimeoutSecs(value: Int?): FunctionOptions.Builder = apply { task_idle_timeout_secs = value }
var FunctionOptions.Builder.taskIdleTimeoutSecs: Int?
  get() = task_idle_timeout_secs
  set(value) { task_idle_timeout_secs = value }
val FunctionOptions.warmPoolSize: Int? get() = warm_pool_size
fun FunctionOptions.hasWarmPoolSize(): Boolean = warm_pool_size != null
fun FunctionOptions.Builder.setWarmPoolSize(value: Int?): FunctionOptions.Builder = apply { warm_pool_size = value }
var FunctionOptions.Builder.warmPoolSize: Int?
  get() = warm_pool_size
  set(value) { warm_pool_size = value }
val FunctionOptions.targetConcurrentInputs: Int? get() = target_concurrent_inputs
fun FunctionOptions.hasTargetConcurrentInputs(): Boolean = target_concurrent_inputs != null
fun FunctionOptions.Builder.setTargetConcurrentInputs(value: Int?): FunctionOptions.Builder = apply { target_concurrent_inputs = value }
var FunctionOptions.Builder.targetConcurrentInputs: Int?
  get() = target_concurrent_inputs
  set(value) { target_concurrent_inputs = value }
val FunctionOptions.replaceVolumeMounts: Boolean get() = replace_volume_mounts
fun FunctionOptions.Builder.setReplaceVolumeMounts(value: Boolean): FunctionOptions.Builder = apply { replace_volume_mounts = value }
var FunctionOptions.Builder.replaceVolumeMounts: Boolean
  get() = replace_volume_mounts
  set(value) { replace_volume_mounts = value }
val FunctionOptions.replaceSecretIds: Boolean get() = replace_secret_ids
fun FunctionOptions.Builder.setReplaceSecretIds(value: Boolean): FunctionOptions.Builder = apply { replace_secret_ids = value }
var FunctionOptions.Builder.replaceSecretIds: Boolean
  get() = replace_secret_ids
  set(value) { replace_secret_ids = value }
val FunctionOptions.bufferContainers: Int? get() = buffer_containers
fun FunctionOptions.hasBufferContainers(): Boolean = buffer_containers != null
fun FunctionOptions.Builder.setBufferContainers(value: Int?): FunctionOptions.Builder = apply { buffer_containers = value }
var FunctionOptions.Builder.bufferContainers: Int?
  get() = buffer_containers
  set(value) { buffer_containers = value }
val FunctionOptions.maxConcurrentInputs: Int? get() = max_concurrent_inputs
fun FunctionOptions.hasMaxConcurrentInputs(): Boolean = max_concurrent_inputs != null
fun FunctionOptions.Builder.setMaxConcurrentInputs(value: Int?): FunctionOptions.Builder = apply { max_concurrent_inputs = value }
var FunctionOptions.Builder.maxConcurrentInputs: Int?
  get() = max_concurrent_inputs
  set(value) { max_concurrent_inputs = value }
val FunctionOptions.batchMaxSize: Int? get() = batch_max_size
fun FunctionOptions.hasBatchMaxSize(): Boolean = batch_max_size != null
fun FunctionOptions.Builder.setBatchMaxSize(value: Int?): FunctionOptions.Builder = apply { batch_max_size = value }
var FunctionOptions.Builder.batchMaxSize: Int?
  get() = batch_max_size
  set(value) { batch_max_size = value }
val FunctionOptions.batchLingerMs: Long? get() = batch_linger_ms
fun FunctionOptions.hasBatchLingerMs(): Boolean = batch_linger_ms != null
fun FunctionOptions.Builder.setBatchLingerMs(value: Long?): FunctionOptions.Builder = apply { batch_linger_ms = value }
var FunctionOptions.Builder.batchLingerMs: Long?
  get() = batch_linger_ms
  set(value) { batch_linger_ms = value }
val FunctionOptions.schedulerPlacement: SchedulerPlacement? get() = scheduler_placement
fun FunctionOptions.hasSchedulerPlacement(): Boolean = scheduler_placement != null
fun FunctionOptions.Builder.setSchedulerPlacement(value: SchedulerPlacement?): FunctionOptions.Builder = apply { scheduler_placement = value }
var FunctionOptions.Builder.schedulerPlacement: SchedulerPlacement?
  get() = scheduler_placement
  set(value) { scheduler_placement = value }
val FunctionOptions.cloudProviderStr: String? get() = cloud_provider_str
fun FunctionOptions.hasCloudProviderStr(): Boolean = cloud_provider_str != null
fun FunctionOptions.Builder.setCloudProviderStr(value: String?): FunctionOptions.Builder = apply { cloud_provider_str = value }
var FunctionOptions.Builder.cloudProviderStr: String?
  get() = cloud_provider_str
  set(value) { cloud_provider_str = value }
val FunctionOptions.replaceCloudBucketMounts: Boolean get() = replace_cloud_bucket_mounts
fun FunctionOptions.Builder.setReplaceCloudBucketMounts(value: Boolean): FunctionOptions.Builder = apply { replace_cloud_bucket_mounts = value }
var FunctionOptions.Builder.replaceCloudBucketMounts: Boolean
  get() = replace_cloud_bucket_mounts
  set(value) { replace_cloud_bucket_mounts = value }
val FunctionOptions.secretIds: List<String> get() = secret_ids
val FunctionOptions.secretIdsList: List<String> get() = secret_ids
val FunctionOptions.secretIdsCount: Int get() = secret_ids.size
fun FunctionOptions.Builder.addSecretIds(value: String): FunctionOptions.Builder = apply { secret_ids = secret_ids + value }
fun FunctionOptions.Builder.addAllSecretIds(items: Iterable<String>): FunctionOptions.Builder = apply { this.secret_ids = this.secret_ids + items }
fun FunctionOptions.Builder.clearSecretIds(): FunctionOptions.Builder = apply { secret_ids = emptyList() }
val FunctionOptions.mountIds: List<String> get() = mount_ids
val FunctionOptions.mountIdsList: List<String> get() = mount_ids
val FunctionOptions.mountIdsCount: Int get() = mount_ids.size
fun FunctionOptions.Builder.addMountIds(value: String): FunctionOptions.Builder = apply { mount_ids = mount_ids + value }
fun FunctionOptions.Builder.addAllMountIds(items: Iterable<String>): FunctionOptions.Builder = apply { this.mount_ids = this.mount_ids + items }
fun FunctionOptions.Builder.clearMountIds(): FunctionOptions.Builder = apply { mount_ids = emptyList() }
val FunctionOptions.volumeMounts: List<VolumeMount> get() = volume_mounts
val FunctionOptions.volumeMountsList: List<VolumeMount> get() = volume_mounts
val FunctionOptions.volumeMountsCount: Int get() = volume_mounts.size
fun FunctionOptions.Builder.addVolumeMounts(value: VolumeMount): FunctionOptions.Builder = apply { volume_mounts = volume_mounts + value }
fun FunctionOptions.Builder.addAllVolumeMounts(items: Iterable<VolumeMount>): FunctionOptions.Builder = apply { this.volume_mounts = this.volume_mounts + items }
fun FunctionOptions.Builder.clearVolumeMounts(): FunctionOptions.Builder = apply { volume_mounts = emptyList() }
val FunctionOptions.cloudBucketMounts: List<CloudBucketMount> get() = cloud_bucket_mounts
val FunctionOptions.cloudBucketMountsList: List<CloudBucketMount> get() = cloud_bucket_mounts
val FunctionOptions.cloudBucketMountsCount: Int get() = cloud_bucket_mounts.size
fun FunctionOptions.Builder.addCloudBucketMounts(value: CloudBucketMount): FunctionOptions.Builder = apply { cloud_bucket_mounts = cloud_bucket_mounts + value }
fun FunctionOptions.Builder.addAllCloudBucketMounts(items: Iterable<CloudBucketMount>): FunctionOptions.Builder = apply { this.cloud_bucket_mounts = this.cloud_bucket_mounts + items }
fun FunctionOptions.Builder.clearCloudBucketMounts(): FunctionOptions.Builder = apply { cloud_bucket_mounts = emptyList() }

fun FunctionPutInputsItem.Companion.newBuilder(): FunctionPutInputsItem.Builder = FunctionPutInputsItem.Builder()
fun FunctionPutInputsItem.Companion.getDefaultInstance(): FunctionPutInputsItem = FunctionPutInputsItem()
fun FunctionPutInputsItem.toByteArray(): ByteArray = encode()
fun FunctionPutInputsItem.Builder.setIdx(value: Int): FunctionPutInputsItem.Builder = apply { idx = value }
var FunctionPutInputsItem.Builder.idx: Int
  get() = idx
  set(value) { idx = value }
fun FunctionPutInputsItem.hasInput(): Boolean = input != null
fun FunctionPutInputsItem.Builder.setInput(value: FunctionInput?): FunctionPutInputsItem.Builder = apply { input = value }
var FunctionPutInputsItem.Builder.input: FunctionInput?
  get() = input
  set(value) { input = value }
val FunctionPutInputsItem.r2Failed: Boolean get() = r2_failed
fun FunctionPutInputsItem.Builder.setR2Failed(value: Boolean): FunctionPutInputsItem.Builder = apply { r2_failed = value }
var FunctionPutInputsItem.Builder.r2Failed: Boolean
  get() = r2_failed
  set(value) { r2_failed = value }
val FunctionPutInputsItem.r2ThroughputBytesS: Long get() = r2_throughput_bytes_s
fun FunctionPutInputsItem.Builder.setR2ThroughputBytesS(value: Long): FunctionPutInputsItem.Builder = apply { r2_throughput_bytes_s = value }
var FunctionPutInputsItem.Builder.r2ThroughputBytesS: Long
  get() = r2_throughput_bytes_s
  set(value) { r2_throughput_bytes_s = value }

fun FunctionPutInputsResponseItem.Companion.newBuilder(): FunctionPutInputsResponseItem.Builder = FunctionPutInputsResponseItem.Builder()
fun FunctionPutInputsResponseItem.Companion.getDefaultInstance(): FunctionPutInputsResponseItem = FunctionPutInputsResponseItem()
fun FunctionPutInputsResponseItem.toByteArray(): ByteArray = encode()
fun FunctionPutInputsResponseItem.Builder.setIdx(value: Int): FunctionPutInputsResponseItem.Builder = apply { idx = value }
var FunctionPutInputsResponseItem.Builder.idx: Int
  get() = idx
  set(value) { idx = value }
val FunctionPutInputsResponseItem.inputId: String get() = input_id
fun FunctionPutInputsResponseItem.Builder.setInputId(value: String): FunctionPutInputsResponseItem.Builder = apply { input_id = value }
var FunctionPutInputsResponseItem.Builder.inputId: String
  get() = input_id
  set(value) { input_id = value }
val FunctionPutInputsResponseItem.inputJwt: String get() = input_jwt
fun FunctionPutInputsResponseItem.Builder.setInputJwt(value: String): FunctionPutInputsResponseItem.Builder = apply { input_jwt = value }
var FunctionPutInputsResponseItem.Builder.inputJwt: String
  get() = input_jwt
  set(value) { input_jwt = value }

fun FunctionRetryInputsItem.Companion.newBuilder(): FunctionRetryInputsItem.Builder = FunctionRetryInputsItem.Builder()
fun FunctionRetryInputsItem.Companion.getDefaultInstance(): FunctionRetryInputsItem = FunctionRetryInputsItem()
fun FunctionRetryInputsItem.toByteArray(): ByteArray = encode()
val FunctionRetryInputsItem.inputJwt: String get() = input_jwt
fun FunctionRetryInputsItem.Builder.setInputJwt(value: String): FunctionRetryInputsItem.Builder = apply { input_jwt = value }
var FunctionRetryInputsItem.Builder.inputJwt: String
  get() = input_jwt
  set(value) { input_jwt = value }
fun FunctionRetryInputsItem.hasInput(): Boolean = input != null
fun FunctionRetryInputsItem.Builder.setInput(value: FunctionInput?): FunctionRetryInputsItem.Builder = apply { input = value }
var FunctionRetryInputsItem.Builder.input: FunctionInput?
  get() = input
  set(value) { input = value }
val FunctionRetryInputsItem.retryCount: Int get() = retry_count
fun FunctionRetryInputsItem.Builder.setRetryCount(value: Int): FunctionRetryInputsItem.Builder = apply { retry_count = value }
var FunctionRetryInputsItem.Builder.retryCount: Int
  get() = retry_count
  set(value) { retry_count = value }

fun FunctionRetryInputsRequest.Companion.newBuilder(): FunctionRetryInputsRequest.Builder = FunctionRetryInputsRequest.Builder()
fun FunctionRetryInputsRequest.Companion.getDefaultInstance(): FunctionRetryInputsRequest = FunctionRetryInputsRequest()
fun FunctionRetryInputsRequest.toByteArray(): ByteArray = encode()
val FunctionRetryInputsRequest.functionCallJwt: String get() = function_call_jwt
fun FunctionRetryInputsRequest.Builder.setFunctionCallJwt(value: String): FunctionRetryInputsRequest.Builder = apply { function_call_jwt = value }
var FunctionRetryInputsRequest.Builder.functionCallJwt: String
  get() = function_call_jwt
  set(value) { function_call_jwt = value }
val FunctionRetryInputsRequest.inputsList: List<FunctionRetryInputsItem> get() = inputs
val FunctionRetryInputsRequest.inputsCount: Int get() = inputs.size
fun FunctionRetryInputsRequest.Builder.addInputs(value: FunctionRetryInputsItem): FunctionRetryInputsRequest.Builder = apply { inputs = inputs + value }
fun FunctionRetryInputsRequest.Builder.addAllInputs(items: Iterable<FunctionRetryInputsItem>): FunctionRetryInputsRequest.Builder = apply { this.inputs = this.inputs + items }
fun FunctionRetryInputsRequest.Builder.clearInputs(): FunctionRetryInputsRequest.Builder = apply { inputs = emptyList() }

fun FunctionRetryInputsResponse.Companion.newBuilder(): FunctionRetryInputsResponse.Builder = FunctionRetryInputsResponse.Builder()
fun FunctionRetryInputsResponse.Companion.getDefaultInstance(): FunctionRetryInputsResponse = FunctionRetryInputsResponse()
fun FunctionRetryInputsResponse.toByteArray(): ByteArray = encode()
val FunctionRetryInputsResponse.inputJwts: List<String> get() = input_jwts
val FunctionRetryInputsResponse.inputJwtsList: List<String> get() = input_jwts
val FunctionRetryInputsResponse.inputJwtsCount: Int get() = input_jwts.size
fun FunctionRetryInputsResponse.Builder.addInputJwts(value: String): FunctionRetryInputsResponse.Builder = apply { input_jwts = input_jwts + value }
fun FunctionRetryInputsResponse.Builder.addAllInputJwts(items: Iterable<String>): FunctionRetryInputsResponse.Builder = apply { this.input_jwts = this.input_jwts + items }
fun FunctionRetryInputsResponse.Builder.clearInputJwts(): FunctionRetryInputsResponse.Builder = apply { input_jwts = emptyList() }

fun FunctionRetryPolicy.Companion.newBuilder(): FunctionRetryPolicy.Builder = FunctionRetryPolicy.Builder()
fun FunctionRetryPolicy.Companion.getDefaultInstance(): FunctionRetryPolicy = FunctionRetryPolicy()
fun FunctionRetryPolicy.toByteArray(): ByteArray = encode()
val FunctionRetryPolicy.backoffCoefficient: Float get() = backoff_coefficient
fun FunctionRetryPolicy.Builder.setBackoffCoefficient(value: Float): FunctionRetryPolicy.Builder = apply { backoff_coefficient = value }
var FunctionRetryPolicy.Builder.backoffCoefficient: Float
  get() = backoff_coefficient
  set(value) { backoff_coefficient = value }
val FunctionRetryPolicy.initialDelayMs: Int get() = initial_delay_ms
fun FunctionRetryPolicy.Builder.setInitialDelayMs(value: Int): FunctionRetryPolicy.Builder = apply { initial_delay_ms = value }
var FunctionRetryPolicy.Builder.initialDelayMs: Int
  get() = initial_delay_ms
  set(value) { initial_delay_ms = value }
val FunctionRetryPolicy.maxDelayMs: Int get() = max_delay_ms
fun FunctionRetryPolicy.Builder.setMaxDelayMs(value: Int): FunctionRetryPolicy.Builder = apply { max_delay_ms = value }
var FunctionRetryPolicy.Builder.maxDelayMs: Int
  get() = max_delay_ms
  set(value) { max_delay_ms = value }
fun FunctionRetryPolicy.Builder.setRetries(value: Int): FunctionRetryPolicy.Builder = apply { retries = value }
var FunctionRetryPolicy.Builder.retries: Int
  get() = retries
  set(value) { retries = value }

fun FunctionSchema.Companion.newBuilder(): FunctionSchema.Builder = FunctionSchema.Builder()
fun FunctionSchema.Companion.getDefaultInstance(): FunctionSchema = FunctionSchema()
fun FunctionSchema.toByteArray(): ByteArray = encode()
val FunctionSchema.schemaType: FunctionSchema.FunctionSchemaType get() = schema_type
fun FunctionSchema.Builder.setSchemaType(value: FunctionSchema.FunctionSchemaType): FunctionSchema.Builder = apply { schema_type = value }
var FunctionSchema.Builder.schemaType: FunctionSchema.FunctionSchemaType
  get() = schema_type
  set(value) { schema_type = value }
val FunctionSchema.returnType: GenericPayloadType? get() = return_type
fun FunctionSchema.hasReturnType(): Boolean = return_type != null
fun FunctionSchema.Builder.setReturnType(value: GenericPayloadType?): FunctionSchema.Builder = apply { return_type = value }
var FunctionSchema.Builder.returnType: GenericPayloadType?
  get() = return_type
  set(value) { return_type = value }
val FunctionSchema.argumentsList: List<ClassParameterSpec> get() = arguments
val FunctionSchema.argumentsCount: Int get() = arguments.size
fun FunctionSchema.Builder.addArguments(value: ClassParameterSpec): FunctionSchema.Builder = apply { arguments = arguments + value }
fun FunctionSchema.Builder.addAllArguments(items: Iterable<ClassParameterSpec>): FunctionSchema.Builder = apply { this.arguments = this.arguments + items }
fun FunctionSchema.Builder.clearArguments(): FunctionSchema.Builder = apply { arguments = emptyList() }

fun FunctionStats.Companion.newBuilder(): FunctionStats.Builder = FunctionStats.Builder()
fun FunctionStats.Companion.getDefaultInstance(): FunctionStats = FunctionStats()
fun FunctionStats.toByteArray(): ByteArray = encode()
fun FunctionStats.Builder.setBacklog(value: Int): FunctionStats.Builder = apply { backlog = value }
var FunctionStats.Builder.backlog: Int
  get() = backlog
  set(value) { backlog = value }
val FunctionStats.numTotalTasks: Int get() = num_total_tasks
fun FunctionStats.Builder.setNumTotalTasks(value: Int): FunctionStats.Builder = apply { num_total_tasks = value }
var FunctionStats.Builder.numTotalTasks: Int
  get() = num_total_tasks
  set(value) { num_total_tasks = value }

fun FunctionUpdateSchedulingParamsRequest.Companion.newBuilder(): FunctionUpdateSchedulingParamsRequest.Builder = FunctionUpdateSchedulingParamsRequest.Builder()
fun FunctionUpdateSchedulingParamsRequest.Companion.getDefaultInstance(): FunctionUpdateSchedulingParamsRequest = FunctionUpdateSchedulingParamsRequest()
fun FunctionUpdateSchedulingParamsRequest.toByteArray(): ByteArray = encode()
val FunctionUpdateSchedulingParamsRequest.functionId: String get() = function_id
fun FunctionUpdateSchedulingParamsRequest.Builder.setFunctionId(value: String): FunctionUpdateSchedulingParamsRequest.Builder = apply { function_id = value }
var FunctionUpdateSchedulingParamsRequest.Builder.functionId: String
  get() = function_id
  set(value) { function_id = value }
val FunctionUpdateSchedulingParamsRequest.warmPoolSizeOverride: Int get() = warm_pool_size_override
fun FunctionUpdateSchedulingParamsRequest.Builder.setWarmPoolSizeOverride(value: Int): FunctionUpdateSchedulingParamsRequest.Builder = apply { warm_pool_size_override = value }
var FunctionUpdateSchedulingParamsRequest.Builder.warmPoolSizeOverride: Int
  get() = warm_pool_size_override
  set(value) { warm_pool_size_override = value }
fun FunctionUpdateSchedulingParamsRequest.hasSettings(): Boolean = settings != null
fun FunctionUpdateSchedulingParamsRequest.Builder.setSettings(value: AutoscalerSettings?): FunctionUpdateSchedulingParamsRequest.Builder = apply { settings = value }
var FunctionUpdateSchedulingParamsRequest.Builder.settings: AutoscalerSettings?
  get() = settings
  set(value) { settings = value }

fun GPUConfig.Companion.newBuilder(): GPUConfig.Builder = GPUConfig.Builder()
fun GPUConfig.Companion.getDefaultInstance(): GPUConfig = GPUConfig()
fun GPUConfig.toByteArray(): ByteArray = encode()
fun GPUConfig.Builder.setType(value: GPUType): GPUConfig.Builder = apply { type = value }
var GPUConfig.Builder.type: GPUType
  get() = type
  set(value) { type = value }
fun GPUConfig.Builder.setCount(value: Int): GPUConfig.Builder = apply { count = value }
var GPUConfig.Builder.count: Int
  get() = count
  set(value) { count = value }
val GPUConfig.gpuType: String get() = gpu_type
fun GPUConfig.Builder.setGpuType(value: String): GPUConfig.Builder = apply { gpu_type = value }
var GPUConfig.Builder.gpuType: String
  get() = gpu_type
  set(value) { gpu_type = value }

fun GenericPayloadType.Companion.newBuilder(): GenericPayloadType.Builder = GenericPayloadType.Builder()
fun GenericPayloadType.Companion.getDefaultInstance(): GenericPayloadType = GenericPayloadType()
fun GenericPayloadType.toByteArray(): ByteArray = encode()
val GenericPayloadType.baseType: ParameterType get() = base_type
fun GenericPayloadType.Builder.setBaseType(value: ParameterType): GenericPayloadType.Builder = apply { base_type = value }
var GenericPayloadType.Builder.baseType: ParameterType
  get() = base_type
  set(value) { base_type = value }
val GenericPayloadType.subTypes: List<GenericPayloadType> get() = sub_types
val GenericPayloadType.subTypesList: List<GenericPayloadType> get() = sub_types
val GenericPayloadType.subTypesCount: Int get() = sub_types.size
fun GenericPayloadType.Builder.addSubTypes(value: GenericPayloadType): GenericPayloadType.Builder = apply { sub_types = sub_types + value }
fun GenericPayloadType.Builder.addAllSubTypes(items: Iterable<GenericPayloadType>): GenericPayloadType.Builder = apply { this.sub_types = this.sub_types + items }
fun GenericPayloadType.Builder.clearSubTypes(): GenericPayloadType.Builder = apply { sub_types = emptyList() }

fun GenericResult.Companion.newBuilder(): GenericResult.Builder = GenericResult.Builder()
fun GenericResult.Companion.getDefaultInstance(): GenericResult = GenericResult()
fun GenericResult.toByteArray(): ByteArray = encode()
fun GenericResult.Builder.setStatus(value: GenericResult.GenericStatus): GenericResult.Builder = apply { status = value }
var GenericResult.Builder.status: GenericResult.GenericStatus
  get() = status
  set(value) { status = value }
fun GenericResult.Builder.setException(value: String): GenericResult.Builder = apply { exception = value }
var GenericResult.Builder.exception: String
  get() = exception
  set(value) { exception = value }
fun GenericResult.Builder.setExitcode(value: Int): GenericResult.Builder = apply { exitcode = value }
var GenericResult.Builder.exitcode: Int
  get() = exitcode
  set(value) { exitcode = value }
fun GenericResult.Builder.setTraceback(value: String): GenericResult.Builder = apply { traceback = value }
var GenericResult.Builder.traceback: String
  get() = traceback
  set(value) { traceback = value }
val GenericResult.serializedTb: ByteString get() = serialized_tb
fun GenericResult.Builder.setSerializedTb(value: ByteString): GenericResult.Builder = apply { serialized_tb = value }
var GenericResult.Builder.serializedTb: ByteString
  get() = serialized_tb
  set(value) { serialized_tb = value }
fun GenericResult.Builder.setSerializedTb(value: ProtoByteString): GenericResult.Builder = apply { serialized_tb = value.toByteArray().toByteString() }
val GenericResult.tbLineCache: ByteString get() = tb_line_cache
fun GenericResult.Builder.setTbLineCache(value: ByteString): GenericResult.Builder = apply { tb_line_cache = value }
var GenericResult.Builder.tbLineCache: ByteString
  get() = tb_line_cache
  set(value) { tb_line_cache = value }
fun GenericResult.Builder.setTbLineCache(value: ProtoByteString): GenericResult.Builder = apply { tb_line_cache = value.toByteArray().toByteString() }
val GenericResult.data: ByteString? get() = data_
fun GenericResult.hasData(): Boolean = data_ != null
fun GenericResult.Builder.setData(value: ByteString?): GenericResult.Builder = apply { data_ = value }
var GenericResult.Builder.data: ByteString?
  get() = data_
  set(value) { data_ = value }
fun GenericResult.Builder.setData(value: ProtoByteString?): GenericResult.Builder = apply { data_ = value?.toByteArray()?.toByteString() }
val GenericResult.dataBlobId: String? get() = data_blob_id
fun GenericResult.hasDataBlobId(): Boolean = data_blob_id != null
fun GenericResult.Builder.setDataBlobId(value: String?): GenericResult.Builder = apply { data_blob_id = value }
var GenericResult.Builder.dataBlobId: String?
  get() = data_blob_id
  set(value) { data_blob_id = value }
val GenericResult.propagationReason: String get() = propagation_reason
fun GenericResult.Builder.setPropagationReason(value: String): GenericResult.Builder = apply { propagation_reason = value }
var GenericResult.Builder.propagationReason: String
  get() = propagation_reason
  set(value) { propagation_reason = value }

fun HTTPConfig.Companion.newBuilder(): HTTPConfig.Builder = HTTPConfig.Builder()
fun HTTPConfig.Companion.getDefaultInstance(): HTTPConfig = HTTPConfig()
fun HTTPConfig.toByteArray(): ByteArray = encode()
fun HTTPConfig.Builder.setPort(value: Int): HTTPConfig.Builder = apply { port = value }
var HTTPConfig.Builder.port: Int
  get() = port
  set(value) { port = value }
val HTTPConfig.startupTimeout: Int get() = startup_timeout
fun HTTPConfig.Builder.setStartupTimeout(value: Int): HTTPConfig.Builder = apply { startup_timeout = value }
var HTTPConfig.Builder.startupTimeout: Int
  get() = startup_timeout
  set(value) { startup_timeout = value }
val HTTPConfig.exitGracePeriod: Int get() = exit_grace_period
fun HTTPConfig.Builder.setExitGracePeriod(value: Int): HTTPConfig.Builder = apply { exit_grace_period = value }
var HTTPConfig.Builder.exitGracePeriod: Int
  get() = exit_grace_period
  set(value) { exit_grace_period = value }
val HTTPConfig.h2Enabled: Boolean get() = h2_enabled
fun HTTPConfig.Builder.setH2Enabled(value: Boolean): HTTPConfig.Builder = apply { h2_enabled = value }
var HTTPConfig.Builder.h2Enabled: Boolean
  get() = h2_enabled
  set(value) { h2_enabled = value }
val HTTPConfig.targetConcurrency: Int get() = target_concurrency
fun HTTPConfig.Builder.setTargetConcurrency(value: Int): HTTPConfig.Builder = apply { target_concurrency = value }
var HTTPConfig.Builder.targetConcurrency: Int
  get() = target_concurrency
  set(value) { target_concurrency = value }
val HTTPConfig.proxyRegions: List<String> get() = proxy_regions
val HTTPConfig.proxyRegionsList: List<String> get() = proxy_regions
val HTTPConfig.proxyRegionsCount: Int get() = proxy_regions.size
fun HTTPConfig.Builder.addProxyRegions(value: String): HTTPConfig.Builder = apply { proxy_regions = proxy_regions + value }
fun HTTPConfig.Builder.addAllProxyRegions(items: Iterable<String>): HTTPConfig.Builder = apply { this.proxy_regions = this.proxy_regions + items }
fun HTTPConfig.Builder.clearProxyRegions(): HTTPConfig.Builder = apply { proxy_regions = emptyList() }

fun Image.Companion.newBuilder(): Image.Builder = Image.Builder()
fun Image.Companion.getDefaultInstance(): Image = Image()
fun Image.toByteArray(): ByteArray = encode()
fun Image.Builder.setVersion(value: String): Image.Builder = apply { version = value }
var Image.Builder.version: String
  get() = version
  set(value) { version = value }
val Image.contextMountId: String get() = context_mount_id
fun Image.Builder.setContextMountId(value: String): Image.Builder = apply { context_mount_id = value }
var Image.Builder.contextMountId: String
  get() = context_mount_id
  set(value) { context_mount_id = value }
val Image.gpuConfig: GPUConfig? get() = gpu_config
fun Image.hasGpuConfig(): Boolean = gpu_config != null
fun Image.Builder.setGpuConfig(value: GPUConfig?): Image.Builder = apply { gpu_config = value }
var Image.Builder.gpuConfig: GPUConfig?
  get() = gpu_config
  set(value) { gpu_config = value }
val Image.imageRegistryConfig: ImageRegistryConfig? get() = image_registry_config
fun Image.hasImageRegistryConfig(): Boolean = image_registry_config != null
fun Image.Builder.setImageRegistryConfig(value: ImageRegistryConfig?): Image.Builder = apply { image_registry_config = value }
var Image.Builder.imageRegistryConfig: ImageRegistryConfig?
  get() = image_registry_config
  set(value) { image_registry_config = value }
val Image.buildFunctionDef: String get() = build_function_def
fun Image.Builder.setBuildFunctionDef(value: String): Image.Builder = apply { build_function_def = value }
var Image.Builder.buildFunctionDef: String
  get() = build_function_def
  set(value) { build_function_def = value }
val Image.buildFunctionGlobals: ByteString get() = build_function_globals
fun Image.Builder.setBuildFunctionGlobals(value: ByteString): Image.Builder = apply { build_function_globals = value }
var Image.Builder.buildFunctionGlobals: ByteString
  get() = build_function_globals
  set(value) { build_function_globals = value }
fun Image.Builder.setBuildFunctionGlobals(value: ProtoByteString): Image.Builder = apply { build_function_globals = value.toByteArray().toByteString() }
fun Image.Builder.setRuntime(value: String): Image.Builder = apply { runtime = value }
var Image.Builder.runtime: String
  get() = runtime
  set(value) { runtime = value }
val Image.runtimeDebug: Boolean get() = runtime_debug
fun Image.Builder.setRuntimeDebug(value: Boolean): Image.Builder = apply { runtime_debug = value }
var Image.Builder.runtimeDebug: Boolean
  get() = runtime_debug
  set(value) { runtime_debug = value }
val Image.buildFunction: BuildFunction? get() = build_function
fun Image.hasBuildFunction(): Boolean = build_function != null
fun Image.Builder.setBuildFunction(value: BuildFunction?): Image.Builder = apply { build_function = value }
var Image.Builder.buildFunction: BuildFunction?
  get() = build_function
  set(value) { build_function = value }
val Image.baseImages: List<BaseImage> get() = base_images
val Image.baseImagesList: List<BaseImage> get() = base_images
val Image.baseImagesCount: Int get() = base_images.size
fun Image.Builder.addBaseImages(value: BaseImage): Image.Builder = apply { base_images = base_images + value }
fun Image.Builder.addAllBaseImages(items: Iterable<BaseImage>): Image.Builder = apply { this.base_images = this.base_images + items }
fun Image.Builder.clearBaseImages(): Image.Builder = apply { base_images = emptyList() }
val Image.dockerfileCommands: List<String> get() = dockerfile_commands
val Image.dockerfileCommandsList: List<String> get() = dockerfile_commands
val Image.dockerfileCommandsCount: Int get() = dockerfile_commands.size
fun Image.Builder.addDockerfileCommands(value: String): Image.Builder = apply { dockerfile_commands = dockerfile_commands + value }
fun Image.Builder.addAllDockerfileCommands(items: Iterable<String>): Image.Builder = apply { this.dockerfile_commands = this.dockerfile_commands + items }
fun Image.Builder.clearDockerfileCommands(): Image.Builder = apply { dockerfile_commands = emptyList() }
val Image.contextFiles: List<ImageContextFile> get() = context_files
val Image.contextFilesList: List<ImageContextFile> get() = context_files
val Image.contextFilesCount: Int get() = context_files.size
fun Image.Builder.addContextFiles(value: ImageContextFile): Image.Builder = apply { context_files = context_files + value }
fun Image.Builder.addAllContextFiles(items: Iterable<ImageContextFile>): Image.Builder = apply { this.context_files = this.context_files + items }
fun Image.Builder.clearContextFiles(): Image.Builder = apply { context_files = emptyList() }
val Image.secretIds: List<String> get() = secret_ids
val Image.secretIdsList: List<String> get() = secret_ids
val Image.secretIdsCount: Int get() = secret_ids.size
fun Image.Builder.addSecretIds(value: String): Image.Builder = apply { secret_ids = secret_ids + value }
fun Image.Builder.addAllSecretIds(items: Iterable<String>): Image.Builder = apply { this.secret_ids = this.secret_ids + items }
fun Image.Builder.clearSecretIds(): Image.Builder = apply { secret_ids = emptyList() }
val Image.buildArgs: Map<String, String> get() = build_args
val Image.buildArgsMap: Map<String, String> get() = build_args
fun Image.Builder.putBuildArgs(key: String, value: String): Image.Builder = apply { build_args = build_args + mapOf(key to value) }
fun Image.Builder.putAllBuildArgs(entries: Map<String, String>): Image.Builder = apply { build_args = build_args + entries }
fun Image.Builder.clearBuildArgs(): Image.Builder = apply { build_args = emptyMap() }
val Image.volumeMounts: List<VolumeMount> get() = volume_mounts
val Image.volumeMountsList: List<VolumeMount> get() = volume_mounts
val Image.volumeMountsCount: Int get() = volume_mounts.size
fun Image.Builder.addVolumeMounts(value: VolumeMount): Image.Builder = apply { volume_mounts = volume_mounts + value }
fun Image.Builder.addAllVolumeMounts(items: Iterable<VolumeMount>): Image.Builder = apply { this.volume_mounts = this.volume_mounts + items }
fun Image.Builder.clearVolumeMounts(): Image.Builder = apply { volume_mounts = emptyList() }

fun ImageContextFile.Companion.newBuilder(): ImageContextFile.Builder = ImageContextFile.Builder()
fun ImageContextFile.Companion.getDefaultInstance(): ImageContextFile = ImageContextFile()
fun ImageContextFile.toByteArray(): ByteArray = encode()
fun ImageContextFile.Builder.setFilename(value: String): ImageContextFile.Builder = apply { filename = value }
var ImageContextFile.Builder.filename: String
  get() = filename
  set(value) { filename = value }
val ImageContextFile.data: ByteString get() = data_
fun ImageContextFile.Builder.setData(value: ByteString): ImageContextFile.Builder = apply { data_ = value }
var ImageContextFile.Builder.data: ByteString
  get() = data_
  set(value) { data_ = value }
fun ImageContextFile.Builder.setData(value: ProtoByteString): ImageContextFile.Builder = apply { data_ = value.toByteArray().toByteString() }

fun ImageDeleteRequest.Companion.newBuilder(): ImageDeleteRequest.Builder = ImageDeleteRequest.Builder()
fun ImageDeleteRequest.Companion.getDefaultInstance(): ImageDeleteRequest = ImageDeleteRequest()
fun ImageDeleteRequest.toByteArray(): ByteArray = encode()
val ImageDeleteRequest.imageId: String get() = image_id
fun ImageDeleteRequest.Builder.setImageId(value: String): ImageDeleteRequest.Builder = apply { image_id = value }
var ImageDeleteRequest.Builder.imageId: String
  get() = image_id
  set(value) { image_id = value }

fun ImageFromIdRequest.Companion.newBuilder(): ImageFromIdRequest.Builder = ImageFromIdRequest.Builder()
fun ImageFromIdRequest.Companion.getDefaultInstance(): ImageFromIdRequest = ImageFromIdRequest()
fun ImageFromIdRequest.toByteArray(): ByteArray = encode()
val ImageFromIdRequest.imageId: String get() = image_id
fun ImageFromIdRequest.Builder.setImageId(value: String): ImageFromIdRequest.Builder = apply { image_id = value }
var ImageFromIdRequest.Builder.imageId: String
  get() = image_id
  set(value) { image_id = value }

fun ImageFromIdResponse.Companion.newBuilder(): ImageFromIdResponse.Builder = ImageFromIdResponse.Builder()
fun ImageFromIdResponse.Companion.getDefaultInstance(): ImageFromIdResponse = ImageFromIdResponse()
fun ImageFromIdResponse.toByteArray(): ByteArray = encode()
val ImageFromIdResponse.imageId: String get() = image_id
fun ImageFromIdResponse.Builder.setImageId(value: String): ImageFromIdResponse.Builder = apply { image_id = value }
var ImageFromIdResponse.Builder.imageId: String
  get() = image_id
  set(value) { image_id = value }
fun ImageFromIdResponse.hasMetadata(): Boolean = metadata != null
fun ImageFromIdResponse.Builder.setMetadata(value: ImageMetadata?): ImageFromIdResponse.Builder = apply { metadata = value }
var ImageFromIdResponse.Builder.metadata: ImageMetadata?
  get() = metadata
  set(value) { metadata = value }

fun ImageGetOrCreateRequest.Companion.newBuilder(): ImageGetOrCreateRequest.Builder = ImageGetOrCreateRequest.Builder()
fun ImageGetOrCreateRequest.Companion.getDefaultInstance(): ImageGetOrCreateRequest = ImageGetOrCreateRequest()
fun ImageGetOrCreateRequest.toByteArray(): ByteArray = encode()
fun ImageGetOrCreateRequest.hasImage(): Boolean = image != null
fun ImageGetOrCreateRequest.Builder.setImage(value: Image?): ImageGetOrCreateRequest.Builder = apply { image = value }
var ImageGetOrCreateRequest.Builder.image: Image?
  get() = image
  set(value) { image = value }
val ImageGetOrCreateRequest.appId: String get() = app_id
fun ImageGetOrCreateRequest.Builder.setAppId(value: String): ImageGetOrCreateRequest.Builder = apply { app_id = value }
var ImageGetOrCreateRequest.Builder.appId: String
  get() = app_id
  set(value) { app_id = value }
val ImageGetOrCreateRequest.existingImageId: String get() = existing_image_id
fun ImageGetOrCreateRequest.Builder.setExistingImageId(value: String): ImageGetOrCreateRequest.Builder = apply { existing_image_id = value }
var ImageGetOrCreateRequest.Builder.existingImageId: String
  get() = existing_image_id
  set(value) { existing_image_id = value }
val ImageGetOrCreateRequest.buildFunctionId: String get() = build_function_id
fun ImageGetOrCreateRequest.Builder.setBuildFunctionId(value: String): ImageGetOrCreateRequest.Builder = apply { build_function_id = value }
var ImageGetOrCreateRequest.Builder.buildFunctionId: String
  get() = build_function_id
  set(value) { build_function_id = value }
val ImageGetOrCreateRequest.forceBuild: Boolean get() = force_build
fun ImageGetOrCreateRequest.Builder.setForceBuild(value: Boolean): ImageGetOrCreateRequest.Builder = apply { force_build = value }
var ImageGetOrCreateRequest.Builder.forceBuild: Boolean
  get() = force_build
  set(value) { force_build = value }
fun ImageGetOrCreateRequest.Builder.setNamespace(value: DeploymentNamespace): ImageGetOrCreateRequest.Builder = apply { namespace = value }
var ImageGetOrCreateRequest.Builder.namespace: DeploymentNamespace
  get() = namespace
  set(value) { namespace = value }
val ImageGetOrCreateRequest.builderVersion: String get() = builder_version
fun ImageGetOrCreateRequest.Builder.setBuilderVersion(value: String): ImageGetOrCreateRequest.Builder = apply { builder_version = value }
var ImageGetOrCreateRequest.Builder.builderVersion: String
  get() = builder_version
  set(value) { builder_version = value }
val ImageGetOrCreateRequest.allowGlobalDeployment: Boolean get() = allow_global_deployment
fun ImageGetOrCreateRequest.Builder.setAllowGlobalDeployment(value: Boolean): ImageGetOrCreateRequest.Builder = apply { allow_global_deployment = value }
var ImageGetOrCreateRequest.Builder.allowGlobalDeployment: Boolean
  get() = allow_global_deployment
  set(value) { allow_global_deployment = value }
val ImageGetOrCreateRequest.ignoreCache: Boolean get() = ignore_cache
fun ImageGetOrCreateRequest.Builder.setIgnoreCache(value: Boolean): ImageGetOrCreateRequest.Builder = apply { ignore_cache = value }
var ImageGetOrCreateRequest.Builder.ignoreCache: Boolean
  get() = ignore_cache
  set(value) { ignore_cache = value }

fun ImageGetOrCreateResponse.Companion.newBuilder(): ImageGetOrCreateResponse.Builder = ImageGetOrCreateResponse.Builder()
fun ImageGetOrCreateResponse.Companion.getDefaultInstance(): ImageGetOrCreateResponse = ImageGetOrCreateResponse()
fun ImageGetOrCreateResponse.toByteArray(): ByteArray = encode()
val ImageGetOrCreateResponse.imageId: String get() = image_id
fun ImageGetOrCreateResponse.Builder.setImageId(value: String): ImageGetOrCreateResponse.Builder = apply { image_id = value }
var ImageGetOrCreateResponse.Builder.imageId: String
  get() = image_id
  set(value) { image_id = value }
fun ImageGetOrCreateResponse.hasResult(): Boolean = result != null
fun ImageGetOrCreateResponse.Builder.setResult(value: GenericResult?): ImageGetOrCreateResponse.Builder = apply { result = value }
var ImageGetOrCreateResponse.Builder.result: GenericResult?
  get() = result
  set(value) { result = value }
fun ImageGetOrCreateResponse.hasMetadata(): Boolean = metadata != null
fun ImageGetOrCreateResponse.Builder.setMetadata(value: ImageMetadata?): ImageGetOrCreateResponse.Builder = apply { metadata = value }
var ImageGetOrCreateResponse.Builder.metadata: ImageMetadata?
  get() = metadata
  set(value) { metadata = value }

fun ImageJoinStreamingRequest.Companion.newBuilder(): ImageJoinStreamingRequest.Builder = ImageJoinStreamingRequest.Builder()
fun ImageJoinStreamingRequest.Companion.getDefaultInstance(): ImageJoinStreamingRequest = ImageJoinStreamingRequest()
fun ImageJoinStreamingRequest.toByteArray(): ByteArray = encode()
val ImageJoinStreamingRequest.imageId: String get() = image_id
fun ImageJoinStreamingRequest.Builder.setImageId(value: String): ImageJoinStreamingRequest.Builder = apply { image_id = value }
var ImageJoinStreamingRequest.Builder.imageId: String
  get() = image_id
  set(value) { image_id = value }
fun ImageJoinStreamingRequest.Builder.setTimeout(value: Float): ImageJoinStreamingRequest.Builder = apply { timeout = value }
var ImageJoinStreamingRequest.Builder.timeout: Float
  get() = timeout
  set(value) { timeout = value }
val ImageJoinStreamingRequest.lastEntryId: String get() = last_entry_id
fun ImageJoinStreamingRequest.Builder.setLastEntryId(value: String): ImageJoinStreamingRequest.Builder = apply { last_entry_id = value }
var ImageJoinStreamingRequest.Builder.lastEntryId: String
  get() = last_entry_id
  set(value) { last_entry_id = value }
val ImageJoinStreamingRequest.includeLogsForFinished: Boolean get() = include_logs_for_finished
fun ImageJoinStreamingRequest.Builder.setIncludeLogsForFinished(value: Boolean): ImageJoinStreamingRequest.Builder = apply { include_logs_for_finished = value }
var ImageJoinStreamingRequest.Builder.includeLogsForFinished: Boolean
  get() = include_logs_for_finished
  set(value) { include_logs_for_finished = value }

fun ImageJoinStreamingResponse.Companion.newBuilder(): ImageJoinStreamingResponse.Builder = ImageJoinStreamingResponse.Builder()
fun ImageJoinStreamingResponse.Companion.getDefaultInstance(): ImageJoinStreamingResponse = ImageJoinStreamingResponse()
fun ImageJoinStreamingResponse.toByteArray(): ByteArray = encode()
fun ImageJoinStreamingResponse.hasResult(): Boolean = result != null
fun ImageJoinStreamingResponse.Builder.setResult(value: GenericResult?): ImageJoinStreamingResponse.Builder = apply { result = value }
var ImageJoinStreamingResponse.Builder.result: GenericResult?
  get() = result
  set(value) { result = value }
val ImageJoinStreamingResponse.entryId: String get() = entry_id
fun ImageJoinStreamingResponse.Builder.setEntryId(value: String): ImageJoinStreamingResponse.Builder = apply { entry_id = value }
var ImageJoinStreamingResponse.Builder.entryId: String
  get() = entry_id
  set(value) { entry_id = value }
fun ImageJoinStreamingResponse.Builder.setEof(value: Boolean): ImageJoinStreamingResponse.Builder = apply { eof = value }
var ImageJoinStreamingResponse.Builder.eof: Boolean
  get() = eof
  set(value) { eof = value }
fun ImageJoinStreamingResponse.hasMetadata(): Boolean = metadata != null
fun ImageJoinStreamingResponse.Builder.setMetadata(value: ImageMetadata?): ImageJoinStreamingResponse.Builder = apply { metadata = value }
var ImageJoinStreamingResponse.Builder.metadata: ImageMetadata?
  get() = metadata
  set(value) { metadata = value }
val ImageJoinStreamingResponse.taskLogs: List<TaskLogs> get() = task_logs
val ImageJoinStreamingResponse.taskLogsList: List<TaskLogs> get() = task_logs
val ImageJoinStreamingResponse.taskLogsCount: Int get() = task_logs.size
fun ImageJoinStreamingResponse.Builder.addTaskLogs(value: TaskLogs): ImageJoinStreamingResponse.Builder = apply { task_logs = task_logs + value }
fun ImageJoinStreamingResponse.Builder.addAllTaskLogs(items: Iterable<TaskLogs>): ImageJoinStreamingResponse.Builder = apply { this.task_logs = this.task_logs + items }
fun ImageJoinStreamingResponse.Builder.clearTaskLogs(): ImageJoinStreamingResponse.Builder = apply { task_logs = emptyList() }

fun ImageMetadata.Companion.newBuilder(): ImageMetadata.Builder = ImageMetadata.Builder()
fun ImageMetadata.Companion.getDefaultInstance(): ImageMetadata = ImageMetadata()
fun ImageMetadata.toByteArray(): ByteArray = encode()
val ImageMetadata.pythonVersionInfo: String? get() = python_version_info
fun ImageMetadata.hasPythonVersionInfo(): Boolean = python_version_info != null
fun ImageMetadata.Builder.setPythonVersionInfo(value: String?): ImageMetadata.Builder = apply { python_version_info = value }
var ImageMetadata.Builder.pythonVersionInfo: String?
  get() = python_version_info
  set(value) { python_version_info = value }
fun ImageMetadata.hasWorkdir(): Boolean = workdir != null
fun ImageMetadata.Builder.setWorkdir(value: String?): ImageMetadata.Builder = apply { workdir = value }
var ImageMetadata.Builder.workdir: String?
  get() = workdir
  set(value) { workdir = value }
val ImageMetadata.libcVersionInfo: String? get() = libc_version_info
fun ImageMetadata.hasLibcVersionInfo(): Boolean = libc_version_info != null
fun ImageMetadata.Builder.setLibcVersionInfo(value: String?): ImageMetadata.Builder = apply { libc_version_info = value }
var ImageMetadata.Builder.libcVersionInfo: String?
  get() = libc_version_info
  set(value) { libc_version_info = value }
val ImageMetadata.imageBuilderVersion: String? get() = image_builder_version
fun ImageMetadata.hasImageBuilderVersion(): Boolean = image_builder_version != null
fun ImageMetadata.Builder.setImageBuilderVersion(value: String?): ImageMetadata.Builder = apply { image_builder_version = value }
var ImageMetadata.Builder.imageBuilderVersion: String?
  get() = image_builder_version
  set(value) { image_builder_version = value }
val ImageMetadata.pythonPackages: Map<String, String> get() = python_packages
val ImageMetadata.pythonPackagesMap: Map<String, String> get() = python_packages
fun ImageMetadata.Builder.putPythonPackages(key: String, value: String): ImageMetadata.Builder = apply { python_packages = python_packages + mapOf(key to value) }
fun ImageMetadata.Builder.putAllPythonPackages(entries: Map<String, String>): ImageMetadata.Builder = apply { python_packages = python_packages + entries }
fun ImageMetadata.Builder.clearPythonPackages(): ImageMetadata.Builder = apply { python_packages = emptyMap() }

fun ImageRegistryConfig.Companion.newBuilder(): ImageRegistryConfig.Builder = ImageRegistryConfig.Builder()
fun ImageRegistryConfig.Companion.getDefaultInstance(): ImageRegistryConfig = ImageRegistryConfig()
fun ImageRegistryConfig.toByteArray(): ByteArray = encode()
val ImageRegistryConfig.registryAuthType: RegistryAuthType get() = registry_auth_type
fun ImageRegistryConfig.Builder.setRegistryAuthType(value: RegistryAuthType): ImageRegistryConfig.Builder = apply { registry_auth_type = value }
var ImageRegistryConfig.Builder.registryAuthType: RegistryAuthType
  get() = registry_auth_type
  set(value) { registry_auth_type = value }
val ImageRegistryConfig.secretId: String get() = secret_id
fun ImageRegistryConfig.Builder.setSecretId(value: String): ImageRegistryConfig.Builder = apply { secret_id = value }
var ImageRegistryConfig.Builder.secretId: String
  get() = secret_id
  set(value) { secret_id = value }

fun MethodDefinition.Companion.newBuilder(): MethodDefinition.Builder = MethodDefinition.Builder()
fun MethodDefinition.Companion.getDefaultInstance(): MethodDefinition = MethodDefinition()
fun MethodDefinition.toByteArray(): ByteArray = encode()
val MethodDefinition.functionName: String get() = function_name
fun MethodDefinition.Builder.setFunctionName(value: String): MethodDefinition.Builder = apply { function_name = value }
var MethodDefinition.Builder.functionName: String
  get() = function_name
  set(value) { function_name = value }
val MethodDefinition.functionType: Function.FunctionType get() = function_type
fun MethodDefinition.Builder.setFunctionType(value: Function.FunctionType): MethodDefinition.Builder = apply { function_type = value }
var MethodDefinition.Builder.functionType: Function.FunctionType
  get() = function_type
  set(value) { function_type = value }
val MethodDefinition.webhookConfig: WebhookConfig? get() = webhook_config
fun MethodDefinition.hasWebhookConfig(): Boolean = webhook_config != null
fun MethodDefinition.Builder.setWebhookConfig(value: WebhookConfig?): MethodDefinition.Builder = apply { webhook_config = value }
var MethodDefinition.Builder.webhookConfig: WebhookConfig?
  get() = webhook_config
  set(value) { webhook_config = value }
val MethodDefinition.webUrl: String get() = web_url
fun MethodDefinition.Builder.setWebUrl(value: String): MethodDefinition.Builder = apply { web_url = value }
var MethodDefinition.Builder.webUrl: String
  get() = web_url
  set(value) { web_url = value }
val MethodDefinition.webUrlInfo: WebUrlInfo? get() = web_url_info
fun MethodDefinition.hasWebUrlInfo(): Boolean = web_url_info != null
fun MethodDefinition.Builder.setWebUrlInfo(value: WebUrlInfo?): MethodDefinition.Builder = apply { web_url_info = value }
var MethodDefinition.Builder.webUrlInfo: WebUrlInfo?
  get() = web_url_info
  set(value) { web_url_info = value }
val MethodDefinition.functionSchema: FunctionSchema? get() = function_schema
fun MethodDefinition.hasFunctionSchema(): Boolean = function_schema != null
fun MethodDefinition.Builder.setFunctionSchema(value: FunctionSchema?): MethodDefinition.Builder = apply { function_schema = value }
var MethodDefinition.Builder.functionSchema: FunctionSchema?
  get() = function_schema
  set(value) { function_schema = value }
val MethodDefinition.customDomainInfo: List<CustomDomainInfo> get() = custom_domain_info
val MethodDefinition.customDomainInfoList: List<CustomDomainInfo> get() = custom_domain_info
val MethodDefinition.customDomainInfoCount: Int get() = custom_domain_info.size
fun MethodDefinition.Builder.addCustomDomainInfo(value: CustomDomainInfo): MethodDefinition.Builder = apply { custom_domain_info = custom_domain_info + value }
fun MethodDefinition.Builder.addAllCustomDomainInfo(items: Iterable<CustomDomainInfo>): MethodDefinition.Builder = apply { this.custom_domain_info = this.custom_domain_info + items }
fun MethodDefinition.Builder.clearCustomDomainInfo(): MethodDefinition.Builder = apply { custom_domain_info = emptyList() }
val MethodDefinition.supportedInputFormats: List<DataFormat> get() = supported_input_formats
val MethodDefinition.supportedInputFormatsList: List<DataFormat> get() = supported_input_formats
val MethodDefinition.supportedInputFormatsCount: Int get() = supported_input_formats.size
fun MethodDefinition.Builder.addSupportedInputFormats(value: DataFormat): MethodDefinition.Builder = apply { supported_input_formats = supported_input_formats + value }
fun MethodDefinition.Builder.addAllSupportedInputFormats(items: Iterable<DataFormat>): MethodDefinition.Builder = apply { this.supported_input_formats = this.supported_input_formats + items }
fun MethodDefinition.Builder.clearSupportedInputFormats(): MethodDefinition.Builder = apply { supported_input_formats = emptyList() }
val MethodDefinition.supportedOutputFormats: List<DataFormat> get() = supported_output_formats
val MethodDefinition.supportedOutputFormatsList: List<DataFormat> get() = supported_output_formats
val MethodDefinition.supportedOutputFormatsCount: Int get() = supported_output_formats.size
fun MethodDefinition.Builder.addSupportedOutputFormats(value: DataFormat): MethodDefinition.Builder = apply { supported_output_formats = supported_output_formats + value }
fun MethodDefinition.Builder.addAllSupportedOutputFormats(items: Iterable<DataFormat>): MethodDefinition.Builder = apply { this.supported_output_formats = this.supported_output_formats + items }
fun MethodDefinition.Builder.clearSupportedOutputFormats(): MethodDefinition.Builder = apply { supported_output_formats = emptyList() }

fun MultiPartUpload.Companion.newBuilder(): MultiPartUpload.Builder = MultiPartUpload.Builder()
fun MultiPartUpload.Companion.getDefaultInstance(): MultiPartUpload = MultiPartUpload()
fun MultiPartUpload.toByteArray(): ByteArray = encode()
val MultiPartUpload.partLength: Long get() = part_length
fun MultiPartUpload.Builder.setPartLength(value: Long): MultiPartUpload.Builder = apply { part_length = value }
var MultiPartUpload.Builder.partLength: Long
  get() = part_length
  set(value) { part_length = value }
val MultiPartUpload.completionUrl: String get() = completion_url
fun MultiPartUpload.Builder.setCompletionUrl(value: String): MultiPartUpload.Builder = apply { completion_url = value }
var MultiPartUpload.Builder.completionUrl: String
  get() = completion_url
  set(value) { completion_url = value }
val MultiPartUpload.uploadUrls: List<String> get() = upload_urls
val MultiPartUpload.uploadUrlsList: List<String> get() = upload_urls
val MultiPartUpload.uploadUrlsCount: Int get() = upload_urls.size
fun MultiPartUpload.Builder.addUploadUrls(value: String): MultiPartUpload.Builder = apply { upload_urls = upload_urls + value }
fun MultiPartUpload.Builder.addAllUploadUrls(items: Iterable<String>): MultiPartUpload.Builder = apply { this.upload_urls = this.upload_urls + items }
fun MultiPartUpload.Builder.clearUploadUrls(): MultiPartUpload.Builder = apply { upload_urls = emptyList() }

fun MultiPartUploadList.Companion.newBuilder(): MultiPartUploadList.Builder = MultiPartUploadList.Builder()
fun MultiPartUploadList.Companion.getDefaultInstance(): MultiPartUploadList = MultiPartUploadList()
fun MultiPartUploadList.toByteArray(): ByteArray = encode()
val MultiPartUploadList.itemsList: List<MultiPartUpload> get() = items
val MultiPartUploadList.itemsCount: Int get() = items.size
fun MultiPartUploadList.Builder.addItems(value: MultiPartUpload): MultiPartUploadList.Builder = apply { items = items + value }
fun MultiPartUploadList.Builder.addAllItems(items: Iterable<MultiPartUpload>): MultiPartUploadList.Builder = apply { this.items = this.items + items }
fun MultiPartUploadList.Builder.clearItems(): MultiPartUploadList.Builder = apply { items = emptyList() }

fun NetworkAccess.Companion.newBuilder(): NetworkAccess.Builder = NetworkAccess.Builder()
fun NetworkAccess.Companion.getDefaultInstance(): NetworkAccess = NetworkAccess()
fun NetworkAccess.toByteArray(): ByteArray = encode()
val NetworkAccess.networkAccessType: NetworkAccess.NetworkAccessType get() = network_access_type
fun NetworkAccess.Builder.setNetworkAccessType(value: NetworkAccess.NetworkAccessType): NetworkAccess.Builder = apply { network_access_type = value }
var NetworkAccess.Builder.networkAccessType: NetworkAccess.NetworkAccessType
  get() = network_access_type
  set(value) { network_access_type = value }
val NetworkAccess.allowedCidrs: List<String> get() = allowed_cidrs
val NetworkAccess.allowedCidrsList: List<String> get() = allowed_cidrs
val NetworkAccess.allowedCidrsCount: Int get() = allowed_cidrs.size
fun NetworkAccess.Builder.addAllowedCidrs(value: String): NetworkAccess.Builder = apply { allowed_cidrs = allowed_cidrs + value }
fun NetworkAccess.Builder.addAllAllowedCidrs(items: Iterable<String>): NetworkAccess.Builder = apply { this.allowed_cidrs = this.allowed_cidrs + items }
fun NetworkAccess.Builder.clearAllowedCidrs(): NetworkAccess.Builder = apply { allowed_cidrs = emptyList() }

fun ObjectDependency.Companion.newBuilder(): ObjectDependency.Builder = ObjectDependency.Builder()
fun ObjectDependency.Companion.getDefaultInstance(): ObjectDependency = ObjectDependency()
fun ObjectDependency.toByteArray(): ByteArray = encode()
val ObjectDependency.objectId: String get() = object_id
fun ObjectDependency.Builder.setObjectId(value: String): ObjectDependency.Builder = apply { object_id = value }
var ObjectDependency.Builder.objectId: String
  get() = object_id
  set(value) { object_id = value }

fun PTYInfo.Companion.newBuilder(): PTYInfo.Builder = PTYInfo.Builder()
fun PTYInfo.Companion.getDefaultInstance(): PTYInfo = PTYInfo()
fun PTYInfo.toByteArray(): ByteArray = encode()
fun PTYInfo.Builder.setEnabled(value: Boolean): PTYInfo.Builder = apply { enabled = value }
var PTYInfo.Builder.enabled: Boolean
  get() = enabled
  set(value) { enabled = value }
val PTYInfo.winszRows: Int get() = winsz_rows
fun PTYInfo.Builder.setWinszRows(value: Int): PTYInfo.Builder = apply { winsz_rows = value }
var PTYInfo.Builder.winszRows: Int
  get() = winsz_rows
  set(value) { winsz_rows = value }
val PTYInfo.winszCols: Int get() = winsz_cols
fun PTYInfo.Builder.setWinszCols(value: Int): PTYInfo.Builder = apply { winsz_cols = value }
var PTYInfo.Builder.winszCols: Int
  get() = winsz_cols
  set(value) { winsz_cols = value }
val PTYInfo.envTerm: String get() = env_term
fun PTYInfo.Builder.setEnvTerm(value: String): PTYInfo.Builder = apply { env_term = value }
var PTYInfo.Builder.envTerm: String
  get() = env_term
  set(value) { env_term = value }
val PTYInfo.envColorterm: String get() = env_colorterm
fun PTYInfo.Builder.setEnvColorterm(value: String): PTYInfo.Builder = apply { env_colorterm = value }
var PTYInfo.Builder.envColorterm: String
  get() = env_colorterm
  set(value) { env_colorterm = value }
val PTYInfo.envTermProgram: String get() = env_term_program
fun PTYInfo.Builder.setEnvTermProgram(value: String): PTYInfo.Builder = apply { env_term_program = value }
var PTYInfo.Builder.envTermProgram: String
  get() = env_term_program
  set(value) { env_term_program = value }
val PTYInfo.ptyType: PTYInfo.PTYType get() = pty_type
fun PTYInfo.Builder.setPtyType(value: PTYInfo.PTYType): PTYInfo.Builder = apply { pty_type = value }
var PTYInfo.Builder.ptyType: PTYInfo.PTYType
  get() = pty_type
  set(value) { pty_type = value }
val PTYInfo.noTerminateOnIdleStdin: Boolean get() = no_terminate_on_idle_stdin
fun PTYInfo.Builder.setNoTerminateOnIdleStdin(value: Boolean): PTYInfo.Builder = apply { no_terminate_on_idle_stdin = value }
var PTYInfo.Builder.noTerminateOnIdleStdin: Boolean
  get() = no_terminate_on_idle_stdin
  set(value) { no_terminate_on_idle_stdin = value }

fun PortSpec.Companion.newBuilder(): PortSpec.Builder = PortSpec.Builder()
fun PortSpec.Companion.getDefaultInstance(): PortSpec = PortSpec()
fun PortSpec.toByteArray(): ByteArray = encode()
fun PortSpec.Builder.setPort(value: Int): PortSpec.Builder = apply { port = value }
var PortSpec.Builder.port: Int
  get() = port
  set(value) { port = value }
fun PortSpec.Builder.setUnencrypted(value: Boolean): PortSpec.Builder = apply { unencrypted = value }
var PortSpec.Builder.unencrypted: Boolean
  get() = unencrypted
  set(value) { unencrypted = value }
val PortSpec.tunnelType: TunnelType? get() = tunnel_type
fun PortSpec.hasTunnelType(): Boolean = tunnel_type != null
fun PortSpec.Builder.setTunnelType(value: TunnelType?): PortSpec.Builder = apply { tunnel_type = value }
var PortSpec.Builder.tunnelType: TunnelType?
  get() = tunnel_type
  set(value) { tunnel_type = value }

fun PortSpecs.Companion.newBuilder(): PortSpecs.Builder = PortSpecs.Builder()
fun PortSpecs.Companion.getDefaultInstance(): PortSpecs = PortSpecs()
fun PortSpecs.toByteArray(): ByteArray = encode()
val PortSpecs.portsList: List<PortSpec> get() = ports
val PortSpecs.portsCount: Int get() = ports.size
fun PortSpecs.Builder.addPorts(value: PortSpec): PortSpecs.Builder = apply { ports = ports + value }
fun PortSpecs.Builder.addAllPorts(items: Iterable<PortSpec>): PortSpecs.Builder = apply { this.ports = this.ports + items }
fun PortSpecs.Builder.clearPorts(): PortSpecs.Builder = apply { ports = emptyList() }

fun Proxy.Companion.newBuilder(): Proxy.Builder = Proxy.Builder()
fun Proxy.Companion.getDefaultInstance(): Proxy = Proxy()
fun Proxy.toByteArray(): ByteArray = encode()
fun Proxy.Builder.setName(value: String): Proxy.Builder = apply { name = value }
var Proxy.Builder.name: String
  get() = name
  set(value) { name = value }
val Proxy.createdAt: Double get() = created_at
fun Proxy.Builder.setCreatedAt(value: Double): Proxy.Builder = apply { created_at = value }
var Proxy.Builder.createdAt: Double
  get() = created_at
  set(value) { created_at = value }
val Proxy.environmentName: String get() = environment_name
fun Proxy.Builder.setEnvironmentName(value: String): Proxy.Builder = apply { environment_name = value }
var Proxy.Builder.environmentName: String
  get() = environment_name
  set(value) { environment_name = value }
val Proxy.proxyId: String get() = proxy_id
fun Proxy.Builder.setProxyId(value: String): Proxy.Builder = apply { proxy_id = value }
var Proxy.Builder.proxyId: String
  get() = proxy_id
  set(value) { proxy_id = value }
fun Proxy.Builder.setRegion(value: String): Proxy.Builder = apply { region = value }
var Proxy.Builder.region: String
  get() = region
  set(value) { region = value }
val Proxy.proxyIps: List<ProxyIp> get() = proxy_ips
val Proxy.proxyIpsList: List<ProxyIp> get() = proxy_ips
val Proxy.proxyIpsCount: Int get() = proxy_ips.size
fun Proxy.Builder.addProxyIps(value: ProxyIp): Proxy.Builder = apply { proxy_ips = proxy_ips + value }
fun Proxy.Builder.addAllProxyIps(items: Iterable<ProxyIp>): Proxy.Builder = apply { this.proxy_ips = this.proxy_ips + items }
fun Proxy.Builder.clearProxyIps(): Proxy.Builder = apply { proxy_ips = emptyList() }

fun ProxyGetRequest.Companion.newBuilder(): ProxyGetRequest.Builder = ProxyGetRequest.Builder()
fun ProxyGetRequest.Companion.getDefaultInstance(): ProxyGetRequest = ProxyGetRequest()
fun ProxyGetRequest.toByteArray(): ByteArray = encode()
fun ProxyGetRequest.Builder.setName(value: String): ProxyGetRequest.Builder = apply { name = value }
var ProxyGetRequest.Builder.name: String
  get() = name
  set(value) { name = value }
val ProxyGetRequest.environmentName: String get() = environment_name
fun ProxyGetRequest.Builder.setEnvironmentName(value: String): ProxyGetRequest.Builder = apply { environment_name = value }
var ProxyGetRequest.Builder.environmentName: String
  get() = environment_name
  set(value) { environment_name = value }

fun ProxyGetResponse.Companion.newBuilder(): ProxyGetResponse.Builder = ProxyGetResponse.Builder()
fun ProxyGetResponse.Companion.getDefaultInstance(): ProxyGetResponse = ProxyGetResponse()
fun ProxyGetResponse.toByteArray(): ByteArray = encode()
fun ProxyGetResponse.hasProxy(): Boolean = proxy != null
fun ProxyGetResponse.Builder.setProxy(value: Proxy?): ProxyGetResponse.Builder = apply { proxy = value }
var ProxyGetResponse.Builder.proxy: Proxy?
  get() = proxy
  set(value) { proxy = value }

fun ProxyIp.Companion.newBuilder(): ProxyIp.Builder = ProxyIp.Builder()
fun ProxyIp.Companion.getDefaultInstance(): ProxyIp = ProxyIp()
fun ProxyIp.toByteArray(): ByteArray = encode()
val ProxyIp.proxyIp: String get() = proxy_ip
fun ProxyIp.Builder.setProxyIp(value: String): ProxyIp.Builder = apply { proxy_ip = value }
var ProxyIp.Builder.proxyIp: String
  get() = proxy_ip
  set(value) { proxy_ip = value }
fun ProxyIp.Builder.setStatus(value: ProxyIpStatus): ProxyIp.Builder = apply { status = value }
var ProxyIp.Builder.status: ProxyIpStatus
  get() = status
  set(value) { status = value }
val ProxyIp.createdAt: Double get() = created_at
fun ProxyIp.Builder.setCreatedAt(value: Double): ProxyIp.Builder = apply { created_at = value }
var ProxyIp.Builder.createdAt: Double
  get() = created_at
  set(value) { created_at = value }
val ProxyIp.environmentName: String get() = environment_name
fun ProxyIp.Builder.setEnvironmentName(value: String): ProxyIp.Builder = apply { environment_name = value }
var ProxyIp.Builder.environmentName: String
  get() = environment_name
  set(value) { environment_name = value }

fun QueueClearRequest.Companion.newBuilder(): QueueClearRequest.Builder = QueueClearRequest.Builder()
fun QueueClearRequest.Companion.getDefaultInstance(): QueueClearRequest = QueueClearRequest()
fun QueueClearRequest.toByteArray(): ByteArray = encode()
val QueueClearRequest.queueId: String get() = queue_id
fun QueueClearRequest.Builder.setQueueId(value: String): QueueClearRequest.Builder = apply { queue_id = value }
var QueueClearRequest.Builder.queueId: String
  get() = queue_id
  set(value) { queue_id = value }
val QueueClearRequest.partitionKey: ByteString get() = partition_key
fun QueueClearRequest.Builder.setPartitionKey(value: ByteString): QueueClearRequest.Builder = apply { partition_key = value }
var QueueClearRequest.Builder.partitionKey: ByteString
  get() = partition_key
  set(value) { partition_key = value }
fun QueueClearRequest.Builder.setPartitionKey(value: ProtoByteString): QueueClearRequest.Builder = apply { partition_key = value.toByteArray().toByteString() }
val QueueClearRequest.allPartitions: Boolean get() = all_partitions
fun QueueClearRequest.Builder.setAllPartitions(value: Boolean): QueueClearRequest.Builder = apply { all_partitions = value }
var QueueClearRequest.Builder.allPartitions: Boolean
  get() = all_partitions
  set(value) { all_partitions = value }

fun QueueDeleteRequest.Companion.newBuilder(): QueueDeleteRequest.Builder = QueueDeleteRequest.Builder()
fun QueueDeleteRequest.Companion.getDefaultInstance(): QueueDeleteRequest = QueueDeleteRequest()
fun QueueDeleteRequest.toByteArray(): ByteArray = encode()
val QueueDeleteRequest.queueId: String get() = queue_id
fun QueueDeleteRequest.Builder.setQueueId(value: String): QueueDeleteRequest.Builder = apply { queue_id = value }
var QueueDeleteRequest.Builder.queueId: String
  get() = queue_id
  set(value) { queue_id = value }

fun QueueGetOrCreateRequest.Companion.newBuilder(): QueueGetOrCreateRequest.Builder = QueueGetOrCreateRequest.Builder()
fun QueueGetOrCreateRequest.Companion.getDefaultInstance(): QueueGetOrCreateRequest = QueueGetOrCreateRequest()
fun QueueGetOrCreateRequest.toByteArray(): ByteArray = encode()
val QueueGetOrCreateRequest.deploymentName: String get() = deployment_name
fun QueueGetOrCreateRequest.Builder.setDeploymentName(value: String): QueueGetOrCreateRequest.Builder = apply { deployment_name = value }
var QueueGetOrCreateRequest.Builder.deploymentName: String
  get() = deployment_name
  set(value) { deployment_name = value }
val QueueGetOrCreateRequest.environmentName: String get() = environment_name
fun QueueGetOrCreateRequest.Builder.setEnvironmentName(value: String): QueueGetOrCreateRequest.Builder = apply { environment_name = value }
var QueueGetOrCreateRequest.Builder.environmentName: String
  get() = environment_name
  set(value) { environment_name = value }
val QueueGetOrCreateRequest.objectCreationType: ObjectCreationType get() = object_creation_type
fun QueueGetOrCreateRequest.Builder.setObjectCreationType(value: ObjectCreationType): QueueGetOrCreateRequest.Builder = apply { object_creation_type = value }
var QueueGetOrCreateRequest.Builder.objectCreationType: ObjectCreationType
  get() = object_creation_type
  set(value) { object_creation_type = value }

fun QueueGetOrCreateResponse.Companion.newBuilder(): QueueGetOrCreateResponse.Builder = QueueGetOrCreateResponse.Builder()
fun QueueGetOrCreateResponse.Companion.getDefaultInstance(): QueueGetOrCreateResponse = QueueGetOrCreateResponse()
fun QueueGetOrCreateResponse.toByteArray(): ByteArray = encode()
val QueueGetOrCreateResponse.queueId: String get() = queue_id
fun QueueGetOrCreateResponse.Builder.setQueueId(value: String): QueueGetOrCreateResponse.Builder = apply { queue_id = value }
var QueueGetOrCreateResponse.Builder.queueId: String
  get() = queue_id
  set(value) { queue_id = value }
fun QueueGetOrCreateResponse.hasMetadata(): Boolean = metadata != null
fun QueueGetOrCreateResponse.Builder.setMetadata(value: QueueMetadata?): QueueGetOrCreateResponse.Builder = apply { metadata = value }
var QueueGetOrCreateResponse.Builder.metadata: QueueMetadata?
  get() = metadata
  set(value) { metadata = value }

fun QueueGetRequest.Companion.newBuilder(): QueueGetRequest.Builder = QueueGetRequest.Builder()
fun QueueGetRequest.Companion.getDefaultInstance(): QueueGetRequest = QueueGetRequest()
fun QueueGetRequest.toByteArray(): ByteArray = encode()
val QueueGetRequest.queueId: String get() = queue_id
fun QueueGetRequest.Builder.setQueueId(value: String): QueueGetRequest.Builder = apply { queue_id = value }
var QueueGetRequest.Builder.queueId: String
  get() = queue_id
  set(value) { queue_id = value }
fun QueueGetRequest.Builder.setTimeout(value: Float): QueueGetRequest.Builder = apply { timeout = value }
var QueueGetRequest.Builder.timeout: Float
  get() = timeout
  set(value) { timeout = value }
val QueueGetRequest.nValues: Int get() = n_values
fun QueueGetRequest.Builder.setNValues(value: Int): QueueGetRequest.Builder = apply { n_values = value }
var QueueGetRequest.Builder.nValues: Int
  get() = n_values
  set(value) { n_values = value }
val QueueGetRequest.partitionKey: ByteString get() = partition_key
fun QueueGetRequest.Builder.setPartitionKey(value: ByteString): QueueGetRequest.Builder = apply { partition_key = value }
var QueueGetRequest.Builder.partitionKey: ByteString
  get() = partition_key
  set(value) { partition_key = value }
fun QueueGetRequest.Builder.setPartitionKey(value: ProtoByteString): QueueGetRequest.Builder = apply { partition_key = value.toByteArray().toByteString() }

fun QueueGetResponse.Companion.newBuilder(): QueueGetResponse.Builder = QueueGetResponse.Builder()
fun QueueGetResponse.Companion.getDefaultInstance(): QueueGetResponse = QueueGetResponse()
fun QueueGetResponse.toByteArray(): ByteArray = encode()
val QueueGetResponse.valuesList: List<ByteString> get() = values
val QueueGetResponse.valuesCount: Int get() = values.size
fun QueueGetResponse.Builder.addValues(value: ByteString): QueueGetResponse.Builder = apply { values = values + value }
fun QueueGetResponse.Builder.addAllValues(items: Iterable<ByteString>): QueueGetResponse.Builder = apply { this.values = this.values + items }
fun QueueGetResponse.Builder.clearValues(): QueueGetResponse.Builder = apply { values = emptyList() }

fun QueueHeartbeatRequest.Companion.newBuilder(): QueueHeartbeatRequest.Builder = QueueHeartbeatRequest.Builder()
fun QueueHeartbeatRequest.Companion.getDefaultInstance(): QueueHeartbeatRequest = QueueHeartbeatRequest()
fun QueueHeartbeatRequest.toByteArray(): ByteArray = encode()
val QueueHeartbeatRequest.queueId: String get() = queue_id
fun QueueHeartbeatRequest.Builder.setQueueId(value: String): QueueHeartbeatRequest.Builder = apply { queue_id = value }
var QueueHeartbeatRequest.Builder.queueId: String
  get() = queue_id
  set(value) { queue_id = value }

fun QueueItem.Companion.newBuilder(): QueueItem.Builder = QueueItem.Builder()
fun QueueItem.Companion.getDefaultInstance(): QueueItem = QueueItem()
fun QueueItem.toByteArray(): ByteArray = encode()
val QueueItem.value: ByteString get() = value_
fun QueueItem.Builder.setValue(value: ByteString): QueueItem.Builder = apply { value_ = value }
var QueueItem.Builder.value: ByteString
  get() = value_
  set(value) { value_ = value }
fun QueueItem.Builder.setValue(value: ProtoByteString): QueueItem.Builder = apply { value_ = value.toByteArray().toByteString() }
val QueueItem.entryId: String get() = entry_id
fun QueueItem.Builder.setEntryId(value: String): QueueItem.Builder = apply { entry_id = value }
var QueueItem.Builder.entryId: String
  get() = entry_id
  set(value) { entry_id = value }

fun QueueLenRequest.Companion.newBuilder(): QueueLenRequest.Builder = QueueLenRequest.Builder()
fun QueueLenRequest.Companion.getDefaultInstance(): QueueLenRequest = QueueLenRequest()
fun QueueLenRequest.toByteArray(): ByteArray = encode()
val QueueLenRequest.queueId: String get() = queue_id
fun QueueLenRequest.Builder.setQueueId(value: String): QueueLenRequest.Builder = apply { queue_id = value }
var QueueLenRequest.Builder.queueId: String
  get() = queue_id
  set(value) { queue_id = value }
val QueueLenRequest.partitionKey: ByteString get() = partition_key
fun QueueLenRequest.Builder.setPartitionKey(value: ByteString): QueueLenRequest.Builder = apply { partition_key = value }
var QueueLenRequest.Builder.partitionKey: ByteString
  get() = partition_key
  set(value) { partition_key = value }
fun QueueLenRequest.Builder.setPartitionKey(value: ProtoByteString): QueueLenRequest.Builder = apply { partition_key = value.toByteArray().toByteString() }
fun QueueLenRequest.Builder.setTotal(value: Boolean): QueueLenRequest.Builder = apply { total = value }
var QueueLenRequest.Builder.total: Boolean
  get() = total
  set(value) { total = value }

fun QueueLenResponse.Companion.newBuilder(): QueueLenResponse.Builder = QueueLenResponse.Builder()
fun QueueLenResponse.Companion.getDefaultInstance(): QueueLenResponse = QueueLenResponse()
fun QueueLenResponse.toByteArray(): ByteArray = encode()
fun QueueLenResponse.Builder.setLen(value: Int): QueueLenResponse.Builder = apply { len = value }
var QueueLenResponse.Builder.len: Int
  get() = len
  set(value) { len = value }

fun QueueMetadata.Companion.newBuilder(): QueueMetadata.Builder = QueueMetadata.Builder()
fun QueueMetadata.Companion.getDefaultInstance(): QueueMetadata = QueueMetadata()
fun QueueMetadata.toByteArray(): ByteArray = encode()
fun QueueMetadata.Builder.setName(value: String): QueueMetadata.Builder = apply { name = value }
var QueueMetadata.Builder.name: String
  get() = name
  set(value) { name = value }
val QueueMetadata.creationInfo: CreationInfo? get() = creation_info
fun QueueMetadata.hasCreationInfo(): Boolean = creation_info != null
fun QueueMetadata.Builder.setCreationInfo(value: CreationInfo?): QueueMetadata.Builder = apply { creation_info = value }
var QueueMetadata.Builder.creationInfo: CreationInfo?
  get() = creation_info
  set(value) { creation_info = value }

fun QueueNextItemsRequest.Companion.newBuilder(): QueueNextItemsRequest.Builder = QueueNextItemsRequest.Builder()
fun QueueNextItemsRequest.Companion.getDefaultInstance(): QueueNextItemsRequest = QueueNextItemsRequest()
fun QueueNextItemsRequest.toByteArray(): ByteArray = encode()
val QueueNextItemsRequest.queueId: String get() = queue_id
fun QueueNextItemsRequest.Builder.setQueueId(value: String): QueueNextItemsRequest.Builder = apply { queue_id = value }
var QueueNextItemsRequest.Builder.queueId: String
  get() = queue_id
  set(value) { queue_id = value }
val QueueNextItemsRequest.partitionKey: ByteString get() = partition_key
fun QueueNextItemsRequest.Builder.setPartitionKey(value: ByteString): QueueNextItemsRequest.Builder = apply { partition_key = value }
var QueueNextItemsRequest.Builder.partitionKey: ByteString
  get() = partition_key
  set(value) { partition_key = value }
fun QueueNextItemsRequest.Builder.setPartitionKey(value: ProtoByteString): QueueNextItemsRequest.Builder = apply { partition_key = value.toByteArray().toByteString() }
val QueueNextItemsRequest.lastEntryId: String get() = last_entry_id
fun QueueNextItemsRequest.Builder.setLastEntryId(value: String): QueueNextItemsRequest.Builder = apply { last_entry_id = value }
var QueueNextItemsRequest.Builder.lastEntryId: String
  get() = last_entry_id
  set(value) { last_entry_id = value }
val QueueNextItemsRequest.itemPollTimeout: Float get() = item_poll_timeout
fun QueueNextItemsRequest.Builder.setItemPollTimeout(value: Float): QueueNextItemsRequest.Builder = apply { item_poll_timeout = value }
var QueueNextItemsRequest.Builder.itemPollTimeout: Float
  get() = item_poll_timeout
  set(value) { item_poll_timeout = value }

fun QueueNextItemsResponse.Companion.newBuilder(): QueueNextItemsResponse.Builder = QueueNextItemsResponse.Builder()
fun QueueNextItemsResponse.Companion.getDefaultInstance(): QueueNextItemsResponse = QueueNextItemsResponse()
fun QueueNextItemsResponse.toByteArray(): ByteArray = encode()
val QueueNextItemsResponse.itemsList: List<QueueItem> get() = items
val QueueNextItemsResponse.itemsCount: Int get() = items.size
fun QueueNextItemsResponse.Builder.addItems(value: QueueItem): QueueNextItemsResponse.Builder = apply { items = items + value }
fun QueueNextItemsResponse.Builder.addAllItems(items: Iterable<QueueItem>): QueueNextItemsResponse.Builder = apply { this.items = this.items + items }
fun QueueNextItemsResponse.Builder.clearItems(): QueueNextItemsResponse.Builder = apply { items = emptyList() }

fun QueuePutRequest.Companion.newBuilder(): QueuePutRequest.Builder = QueuePutRequest.Builder()
fun QueuePutRequest.Companion.getDefaultInstance(): QueuePutRequest = QueuePutRequest()
fun QueuePutRequest.toByteArray(): ByteArray = encode()
val QueuePutRequest.queueId: String get() = queue_id
fun QueuePutRequest.Builder.setQueueId(value: String): QueuePutRequest.Builder = apply { queue_id = value }
var QueuePutRequest.Builder.queueId: String
  get() = queue_id
  set(value) { queue_id = value }
val QueuePutRequest.partitionKey: ByteString get() = partition_key
fun QueuePutRequest.Builder.setPartitionKey(value: ByteString): QueuePutRequest.Builder = apply { partition_key = value }
var QueuePutRequest.Builder.partitionKey: ByteString
  get() = partition_key
  set(value) { partition_key = value }
fun QueuePutRequest.Builder.setPartitionKey(value: ProtoByteString): QueuePutRequest.Builder = apply { partition_key = value.toByteArray().toByteString() }
val QueuePutRequest.partitionTtlSeconds: Int get() = partition_ttl_seconds
fun QueuePutRequest.Builder.setPartitionTtlSeconds(value: Int): QueuePutRequest.Builder = apply { partition_ttl_seconds = value }
var QueuePutRequest.Builder.partitionTtlSeconds: Int
  get() = partition_ttl_seconds
  set(value) { partition_ttl_seconds = value }
val QueuePutRequest.valuesList: List<ByteString> get() = values
val QueuePutRequest.valuesCount: Int get() = values.size
fun QueuePutRequest.Builder.addValues(value: ByteString): QueuePutRequest.Builder = apply { values = values + value }
fun QueuePutRequest.Builder.addAllValues(items: Iterable<ByteString>): QueuePutRequest.Builder = apply { this.values = this.values + items }
fun QueuePutRequest.Builder.clearValues(): QueuePutRequest.Builder = apply { values = emptyList() }

fun RateLimit.Companion.newBuilder(): RateLimit.Builder = RateLimit.Builder()
fun RateLimit.Companion.getDefaultInstance(): RateLimit = RateLimit()
fun RateLimit.toByteArray(): ByteArray = encode()
fun RateLimit.Builder.setLimit(value: Int): RateLimit.Builder = apply { limit = value }
var RateLimit.Builder.limit: Int
  get() = limit
  set(value) { limit = value }
fun RateLimit.Builder.setInterval(value: RateLimitInterval): RateLimit.Builder = apply { interval = value }
var RateLimit.Builder.interval: RateLimitInterval
  get() = interval
  set(value) { interval = value }

fun ResourceInfo.Companion.newBuilder(): ResourceInfo.Builder = ResourceInfo.Builder()
fun ResourceInfo.Companion.getDefaultInstance(): ResourceInfo = ResourceInfo()
fun ResourceInfo.toByteArray(): ByteArray = encode()
val ResourceInfo.memoryMb: ResourceInfo.ResourceValue? get() = memory_mb
fun ResourceInfo.hasMemoryMb(): Boolean = memory_mb != null
fun ResourceInfo.Builder.setMemoryMb(value: ResourceInfo.ResourceValue?): ResourceInfo.Builder = apply { memory_mb = value }
var ResourceInfo.Builder.memoryMb: ResourceInfo.ResourceValue?
  get() = memory_mb
  set(value) { memory_mb = value }
val ResourceInfo.milliCpu: ResourceInfo.ResourceValue? get() = milli_cpu
fun ResourceInfo.hasMilliCpu(): Boolean = milli_cpu != null
fun ResourceInfo.Builder.setMilliCpu(value: ResourceInfo.ResourceValue?): ResourceInfo.Builder = apply { milli_cpu = value }
var ResourceInfo.Builder.milliCpu: ResourceInfo.ResourceValue?
  get() = milli_cpu
  set(value) { milli_cpu = value }
val ResourceInfo.gpuType: String get() = gpu_type
fun ResourceInfo.Builder.setGpuType(value: String): ResourceInfo.Builder = apply { gpu_type = value }
var ResourceInfo.Builder.gpuType: String
  get() = gpu_type
  set(value) { gpu_type = value }
val ResourceInfo.memoryMbMax: Int get() = memory_mb_max
fun ResourceInfo.Builder.setMemoryMbMax(value: Int): ResourceInfo.Builder = apply { memory_mb_max = value }
var ResourceInfo.Builder.memoryMbMax: Int
  get() = memory_mb_max
  set(value) { memory_mb_max = value }
val ResourceInfo.ephemeralDiskMb: Int get() = ephemeral_disk_mb
fun ResourceInfo.Builder.setEphemeralDiskMb(value: Int): ResourceInfo.Builder = apply { ephemeral_disk_mb = value }
var ResourceInfo.Builder.ephemeralDiskMb: Int
  get() = ephemeral_disk_mb
  set(value) { ephemeral_disk_mb = value }
val ResourceInfo.milliCpuMax: Int get() = milli_cpu_max
fun ResourceInfo.Builder.setMilliCpuMax(value: Int): ResourceInfo.Builder = apply { milli_cpu_max = value }
var ResourceInfo.Builder.milliCpuMax: Int
  get() = milli_cpu_max
  set(value) { milli_cpu_max = value }

fun Resources.Companion.newBuilder(): Resources.Builder = Resources.Builder()
fun Resources.Companion.getDefaultInstance(): Resources = Resources()
fun Resources.toByteArray(): ByteArray = encode()
val Resources.memoryMb: Int get() = memory_mb
fun Resources.Builder.setMemoryMb(value: Int): Resources.Builder = apply { memory_mb = value }
var Resources.Builder.memoryMb: Int
  get() = memory_mb
  set(value) { memory_mb = value }
val Resources.milliCpu: Int get() = milli_cpu
fun Resources.Builder.setMilliCpu(value: Int): Resources.Builder = apply { milli_cpu = value }
var Resources.Builder.milliCpu: Int
  get() = milli_cpu
  set(value) { milli_cpu = value }
val Resources.gpuConfig: GPUConfig? get() = gpu_config
fun Resources.hasGpuConfig(): Boolean = gpu_config != null
fun Resources.Builder.setGpuConfig(value: GPUConfig?): Resources.Builder = apply { gpu_config = value }
var Resources.Builder.gpuConfig: GPUConfig?
  get() = gpu_config
  set(value) { gpu_config = value }
val Resources.memoryMbMax: Int get() = memory_mb_max
fun Resources.Builder.setMemoryMbMax(value: Int): Resources.Builder = apply { memory_mb_max = value }
var Resources.Builder.memoryMbMax: Int
  get() = memory_mb_max
  set(value) { memory_mb_max = value }
val Resources.ephemeralDiskMb: Int get() = ephemeral_disk_mb
fun Resources.Builder.setEphemeralDiskMb(value: Int): Resources.Builder = apply { ephemeral_disk_mb = value }
var Resources.Builder.ephemeralDiskMb: Int
  get() = ephemeral_disk_mb
  set(value) { ephemeral_disk_mb = value }
val Resources.milliCpuMax: Int get() = milli_cpu_max
fun Resources.Builder.setMilliCpuMax(value: Int): Resources.Builder = apply { milli_cpu_max = value }
var Resources.Builder.milliCpuMax: Int
  get() = milli_cpu_max
  set(value) { milli_cpu_max = value }
fun Resources.Builder.setRdma(value: Boolean): Resources.Builder = apply { rdma = value }
var Resources.Builder.rdma: Boolean
  get() = rdma
  set(value) { rdma = value }

fun S3Mount.Companion.newBuilder(): S3Mount.Builder = S3Mount.Builder()
fun S3Mount.Companion.getDefaultInstance(): S3Mount = S3Mount()
fun S3Mount.toByteArray(): ByteArray = encode()
val S3Mount.bucketName: String get() = bucket_name
fun S3Mount.Builder.setBucketName(value: String): S3Mount.Builder = apply { bucket_name = value }
var S3Mount.Builder.bucketName: String
  get() = bucket_name
  set(value) { bucket_name = value }
val S3Mount.mountPath: String get() = mount_path
fun S3Mount.Builder.setMountPath(value: String): S3Mount.Builder = apply { mount_path = value }
var S3Mount.Builder.mountPath: String
  get() = mount_path
  set(value) { mount_path = value }
val S3Mount.credentialsSecretId: String get() = credentials_secret_id
fun S3Mount.Builder.setCredentialsSecretId(value: String): S3Mount.Builder = apply { credentials_secret_id = value }
var S3Mount.Builder.credentialsSecretId: String
  get() = credentials_secret_id
  set(value) { credentials_secret_id = value }
val S3Mount.readOnly: Boolean get() = read_only
fun S3Mount.Builder.setReadOnly(value: Boolean): S3Mount.Builder = apply { read_only = value }
var S3Mount.Builder.readOnly: Boolean
  get() = read_only
  set(value) { read_only = value }

fun Sandbox.Companion.newBuilder(): Sandbox.Builder = Sandbox.Builder()
fun Sandbox.Companion.getDefaultInstance(): Sandbox = Sandbox()
fun Sandbox.toByteArray(): ByteArray = encode()
val Sandbox.imageId: String get() = image_id
fun Sandbox.Builder.setImageId(value: String): Sandbox.Builder = apply { image_id = value }
var Sandbox.Builder.imageId: String
  get() = image_id
  set(value) { image_id = value }
fun Sandbox.hasResources(): Boolean = resources != null
fun Sandbox.Builder.setResources(value: Resources?): Sandbox.Builder = apply { resources = value }
var Sandbox.Builder.resources: Resources?
  get() = resources
  set(value) { resources = value }
val Sandbox.cloudProvider: CloudProvider get() = cloud_provider
fun Sandbox.Builder.setCloudProvider(value: CloudProvider): Sandbox.Builder = apply { cloud_provider = value }
var Sandbox.Builder.cloudProvider: CloudProvider
  get() = cloud_provider
  set(value) { cloud_provider = value }
val Sandbox.timeoutSecs: Int get() = timeout_secs
fun Sandbox.Builder.setTimeoutSecs(value: Int): Sandbox.Builder = apply { timeout_secs = value }
var Sandbox.Builder.timeoutSecs: Int
  get() = timeout_secs
  set(value) { timeout_secs = value }
fun Sandbox.hasWorkdir(): Boolean = workdir != null
fun Sandbox.Builder.setWorkdir(value: String?): Sandbox.Builder = apply { workdir = value }
var Sandbox.Builder.workdir: String?
  get() = workdir
  set(value) { workdir = value }
val Sandbox.runtimeDebug: Boolean get() = runtime_debug
fun Sandbox.Builder.setRuntimeDebug(value: Boolean): Sandbox.Builder = apply { runtime_debug = value }
var Sandbox.Builder.runtimeDebug: Boolean
  get() = runtime_debug
  set(value) { runtime_debug = value }
val Sandbox.blockNetwork: Boolean get() = block_network
fun Sandbox.Builder.setBlockNetwork(value: Boolean): Sandbox.Builder = apply { block_network = value }
var Sandbox.Builder.blockNetwork: Boolean
  get() = block_network
  set(value) { block_network = value }
val Sandbox.ptyInfo: PTYInfo? get() = pty_info
fun Sandbox.hasPtyInfo(): Boolean = pty_info != null
fun Sandbox.Builder.setPtyInfo(value: PTYInfo?): Sandbox.Builder = apply { pty_info = value }
var Sandbox.Builder.ptyInfo: PTYInfo?
  get() = pty_info
  set(value) { pty_info = value }
val Sandbox.schedulerPlacement: SchedulerPlacement? get() = scheduler_placement
fun Sandbox.hasSchedulerPlacement(): Boolean = scheduler_placement != null
fun Sandbox.Builder.setSchedulerPlacement(value: SchedulerPlacement?): Sandbox.Builder = apply { scheduler_placement = value }
var Sandbox.Builder.schedulerPlacement: SchedulerPlacement?
  get() = scheduler_placement
  set(value) { scheduler_placement = value }
val Sandbox.workerId: String get() = worker_id
fun Sandbox.Builder.setWorkerId(value: String): Sandbox.Builder = apply { worker_id = value }
var Sandbox.Builder.workerId: String
  get() = worker_id
  set(value) { worker_id = value }
val Sandbox.openPorts: PortSpecs? get() = open_ports
fun Sandbox.hasOpenPorts(): Boolean = open_ports != null
fun Sandbox.Builder.setOpenPorts(value: PortSpecs?): Sandbox.Builder = apply { open_ports = value }
var Sandbox.Builder.openPorts: PortSpecs?
  get() = open_ports
  set(value) { open_ports = value }
val Sandbox.i6pnEnabled: Boolean get() = i6pn_enabled
fun Sandbox.Builder.setI6pnEnabled(value: Boolean): Sandbox.Builder = apply { i6pn_enabled = value }
var Sandbox.Builder.i6pnEnabled: Boolean
  get() = i6pn_enabled
  set(value) { i6pn_enabled = value }
val Sandbox.networkAccess: NetworkAccess? get() = network_access
fun Sandbox.hasNetworkAccess(): Boolean = network_access != null
fun Sandbox.Builder.setNetworkAccess(value: NetworkAccess?): Sandbox.Builder = apply { network_access = value }
var Sandbox.Builder.networkAccess: NetworkAccess?
  get() = network_access
  set(value) { network_access = value }
val Sandbox.proxyId: String? get() = proxy_id
fun Sandbox.hasProxyId(): Boolean = proxy_id != null
fun Sandbox.Builder.setProxyId(value: String?): Sandbox.Builder = apply { proxy_id = value }
var Sandbox.Builder.proxyId: String?
  get() = proxy_id
  set(value) { proxy_id = value }
val Sandbox.enableSnapshot: Boolean get() = enable_snapshot
fun Sandbox.Builder.setEnableSnapshot(value: Boolean): Sandbox.Builder = apply { enable_snapshot = value }
var Sandbox.Builder.enableSnapshot: Boolean
  get() = enable_snapshot
  set(value) { enable_snapshot = value }
val Sandbox.snapshotVersion: Int? get() = snapshot_version
fun Sandbox.hasSnapshotVersion(): Boolean = snapshot_version != null
fun Sandbox.Builder.setSnapshotVersion(value: Int?): Sandbox.Builder = apply { snapshot_version = value }
var Sandbox.Builder.snapshotVersion: Int?
  get() = snapshot_version
  set(value) { snapshot_version = value }
val Sandbox.cloudProviderStr: String get() = cloud_provider_str
fun Sandbox.Builder.setCloudProviderStr(value: String): Sandbox.Builder = apply { cloud_provider_str = value }
var Sandbox.Builder.cloudProviderStr: String
  get() = cloud_provider_str
  set(value) { cloud_provider_str = value }
val Sandbox.runscRuntimeVersion: String? get() = runsc_runtime_version
fun Sandbox.hasRunscRuntimeVersion(): Boolean = runsc_runtime_version != null
fun Sandbox.Builder.setRunscRuntimeVersion(value: String?): Sandbox.Builder = apply { runsc_runtime_version = value }
var Sandbox.Builder.runscRuntimeVersion: String?
  get() = runsc_runtime_version
  set(value) { runsc_runtime_version = value }
fun Sandbox.hasRuntime(): Boolean = runtime != null
fun Sandbox.Builder.setRuntime(value: String?): Sandbox.Builder = apply { runtime = value }
var Sandbox.Builder.runtime: String?
  get() = runtime
  set(value) { runtime = value }
fun Sandbox.Builder.setVerbose(value: Boolean): Sandbox.Builder = apply { verbose = value }
var Sandbox.Builder.verbose: Boolean
  get() = verbose
  set(value) { verbose = value }
fun Sandbox.hasName(): Boolean = name != null
fun Sandbox.Builder.setName(value: String?): Sandbox.Builder = apply { name = value }
var Sandbox.Builder.name: String?
  get() = name
  set(value) { name = value }
val Sandbox.idleTimeoutSecs: Int? get() = idle_timeout_secs
fun Sandbox.hasIdleTimeoutSecs(): Boolean = idle_timeout_secs != null
fun Sandbox.Builder.setIdleTimeoutSecs(value: Int?): Sandbox.Builder = apply { idle_timeout_secs = value }
var Sandbox.Builder.idleTimeoutSecs: Int?
  get() = idle_timeout_secs
  set(value) { idle_timeout_secs = value }
val Sandbox.directSandboxCommandsEnabled: Boolean get() = direct_sandbox_commands_enabled
fun Sandbox.Builder.setDirectSandboxCommandsEnabled(value: Boolean): Sandbox.Builder = apply { direct_sandbox_commands_enabled = value }
var Sandbox.Builder.directSandboxCommandsEnabled: Boolean
  get() = direct_sandbox_commands_enabled
  set(value) { direct_sandbox_commands_enabled = value }
val Sandbox.RestoreInstanceType: String get() = _restore_instance_type
fun Sandbox.Builder.setRestoreInstanceType(value: String): Sandbox.Builder = apply { _restore_instance_type = value }
var Sandbox.Builder.RestoreInstanceType: String
  get() = _restore_instance_type
  set(value) { _restore_instance_type = value }
val Sandbox.customDomain: String get() = custom_domain
fun Sandbox.Builder.setCustomDomain(value: String): Sandbox.Builder = apply { custom_domain = value }
var Sandbox.Builder.customDomain: String
  get() = custom_domain
  set(value) { custom_domain = value }
val Sandbox.entrypointArgs: List<String> get() = entrypoint_args
val Sandbox.entrypointArgsList: List<String> get() = entrypoint_args
val Sandbox.entrypointArgsCount: Int get() = entrypoint_args.size
fun Sandbox.Builder.addEntrypointArgs(value: String): Sandbox.Builder = apply { entrypoint_args = entrypoint_args + value }
fun Sandbox.Builder.addAllEntrypointArgs(items: Iterable<String>): Sandbox.Builder = apply { this.entrypoint_args = this.entrypoint_args + items }
fun Sandbox.Builder.clearEntrypointArgs(): Sandbox.Builder = apply { entrypoint_args = emptyList() }
val Sandbox.mountIds: List<String> get() = mount_ids
val Sandbox.mountIdsList: List<String> get() = mount_ids
val Sandbox.mountIdsCount: Int get() = mount_ids.size
fun Sandbox.Builder.addMountIds(value: String): Sandbox.Builder = apply { mount_ids = mount_ids + value }
fun Sandbox.Builder.addAllMountIds(items: Iterable<String>): Sandbox.Builder = apply { this.mount_ids = this.mount_ids + items }
fun Sandbox.Builder.clearMountIds(): Sandbox.Builder = apply { mount_ids = emptyList() }
val Sandbox.secretIds: List<String> get() = secret_ids
val Sandbox.secretIdsList: List<String> get() = secret_ids
val Sandbox.secretIdsCount: Int get() = secret_ids.size
fun Sandbox.Builder.addSecretIds(value: String): Sandbox.Builder = apply { secret_ids = secret_ids + value }
fun Sandbox.Builder.addAllSecretIds(items: Iterable<String>): Sandbox.Builder = apply { this.secret_ids = this.secret_ids + items }
fun Sandbox.Builder.clearSecretIds(): Sandbox.Builder = apply { secret_ids = emptyList() }
val Sandbox.nfsMounts: List<SharedVolumeMount> get() = nfs_mounts
val Sandbox.nfsMountsList: List<SharedVolumeMount> get() = nfs_mounts
val Sandbox.nfsMountsCount: Int get() = nfs_mounts.size
fun Sandbox.Builder.addNfsMounts(value: SharedVolumeMount): Sandbox.Builder = apply { nfs_mounts = nfs_mounts + value }
fun Sandbox.Builder.addAllNfsMounts(items: Iterable<SharedVolumeMount>): Sandbox.Builder = apply { this.nfs_mounts = this.nfs_mounts + items }
fun Sandbox.Builder.clearNfsMounts(): Sandbox.Builder = apply { nfs_mounts = emptyList() }
val Sandbox.s3Mounts: List<S3Mount> get() = s3_mounts
val Sandbox.s3MountsList: List<S3Mount> get() = s3_mounts
val Sandbox.s3MountsCount: Int get() = s3_mounts.size
fun Sandbox.Builder.addS3Mounts(value: S3Mount): Sandbox.Builder = apply { s3_mounts = s3_mounts + value }
fun Sandbox.Builder.addAllS3Mounts(items: Iterable<S3Mount>): Sandbox.Builder = apply { this.s3_mounts = this.s3_mounts + items }
fun Sandbox.Builder.clearS3Mounts(): Sandbox.Builder = apply { s3_mounts = emptyList() }
val Sandbox.cloudBucketMounts: List<CloudBucketMount> get() = cloud_bucket_mounts
val Sandbox.cloudBucketMountsList: List<CloudBucketMount> get() = cloud_bucket_mounts
val Sandbox.cloudBucketMountsCount: Int get() = cloud_bucket_mounts.size
fun Sandbox.Builder.addCloudBucketMounts(value: CloudBucketMount): Sandbox.Builder = apply { cloud_bucket_mounts = cloud_bucket_mounts + value }
fun Sandbox.Builder.addAllCloudBucketMounts(items: Iterable<CloudBucketMount>): Sandbox.Builder = apply { this.cloud_bucket_mounts = this.cloud_bucket_mounts + items }
fun Sandbox.Builder.clearCloudBucketMounts(): Sandbox.Builder = apply { cloud_bucket_mounts = emptyList() }
val Sandbox.volumeMounts: List<VolumeMount> get() = volume_mounts
val Sandbox.volumeMountsList: List<VolumeMount> get() = volume_mounts
val Sandbox.volumeMountsCount: Int get() = volume_mounts.size
fun Sandbox.Builder.addVolumeMounts(value: VolumeMount): Sandbox.Builder = apply { volume_mounts = volume_mounts + value }
fun Sandbox.Builder.addAllVolumeMounts(items: Iterable<VolumeMount>): Sandbox.Builder = apply { this.volume_mounts = this.volume_mounts + items }
fun Sandbox.Builder.clearVolumeMounts(): Sandbox.Builder = apply { volume_mounts = emptyList() }
val Sandbox.experimentalOptions: Map<String, Boolean> get() = experimental_options
val Sandbox.experimentalOptionsMap: Map<String, Boolean> get() = experimental_options
fun Sandbox.Builder.putExperimentalOptions(key: String, value: Boolean): Sandbox.Builder = apply { experimental_options = experimental_options + mapOf(key to value) }
fun Sandbox.Builder.putAllExperimentalOptions(entries: Map<String, Boolean>): Sandbox.Builder = apply { experimental_options = experimental_options + entries }
fun Sandbox.Builder.clearExperimentalOptions(): Sandbox.Builder = apply { experimental_options = emptyMap() }
val Sandbox.preloadPathPrefixes: List<String> get() = preload_path_prefixes
val Sandbox.preloadPathPrefixesList: List<String> get() = preload_path_prefixes
val Sandbox.preloadPathPrefixesCount: Int get() = preload_path_prefixes.size
fun Sandbox.Builder.addPreloadPathPrefixes(value: String): Sandbox.Builder = apply { preload_path_prefixes = preload_path_prefixes + value }
fun Sandbox.Builder.addAllPreloadPathPrefixes(items: Iterable<String>): Sandbox.Builder = apply { this.preload_path_prefixes = this.preload_path_prefixes + items }
fun Sandbox.Builder.clearPreloadPathPrefixes(): Sandbox.Builder = apply { preload_path_prefixes = emptyList() }

fun SandboxCreateConnectTokenRequest.Companion.newBuilder(): SandboxCreateConnectTokenRequest.Builder = SandboxCreateConnectTokenRequest.Builder()
fun SandboxCreateConnectTokenRequest.Companion.getDefaultInstance(): SandboxCreateConnectTokenRequest = SandboxCreateConnectTokenRequest()
fun SandboxCreateConnectTokenRequest.toByteArray(): ByteArray = encode()
val SandboxCreateConnectTokenRequest.sandboxId: String get() = sandbox_id
fun SandboxCreateConnectTokenRequest.Builder.setSandboxId(value: String): SandboxCreateConnectTokenRequest.Builder = apply { sandbox_id = value }
var SandboxCreateConnectTokenRequest.Builder.sandboxId: String
  get() = sandbox_id
  set(value) { sandbox_id = value }
val SandboxCreateConnectTokenRequest.userMetadata: String get() = user_metadata
fun SandboxCreateConnectTokenRequest.Builder.setUserMetadata(value: String): SandboxCreateConnectTokenRequest.Builder = apply { user_metadata = value }
var SandboxCreateConnectTokenRequest.Builder.userMetadata: String
  get() = user_metadata
  set(value) { user_metadata = value }

fun SandboxCreateConnectTokenResponse.Companion.newBuilder(): SandboxCreateConnectTokenResponse.Builder = SandboxCreateConnectTokenResponse.Builder()
fun SandboxCreateConnectTokenResponse.Companion.getDefaultInstance(): SandboxCreateConnectTokenResponse = SandboxCreateConnectTokenResponse()
fun SandboxCreateConnectTokenResponse.toByteArray(): ByteArray = encode()
fun SandboxCreateConnectTokenResponse.Builder.setUrl(value: String): SandboxCreateConnectTokenResponse.Builder = apply { url = value }
var SandboxCreateConnectTokenResponse.Builder.url: String
  get() = url
  set(value) { url = value }
fun SandboxCreateConnectTokenResponse.Builder.setToken(value: String): SandboxCreateConnectTokenResponse.Builder = apply { token = value }
var SandboxCreateConnectTokenResponse.Builder.token: String
  get() = token
  set(value) { token = value }

fun SandboxCreateRequest.Companion.newBuilder(): SandboxCreateRequest.Builder = SandboxCreateRequest.Builder()
fun SandboxCreateRequest.Companion.getDefaultInstance(): SandboxCreateRequest = SandboxCreateRequest()
fun SandboxCreateRequest.toByteArray(): ByteArray = encode()
val SandboxCreateRequest.appId: String get() = app_id
fun SandboxCreateRequest.Builder.setAppId(value: String): SandboxCreateRequest.Builder = apply { app_id = value }
var SandboxCreateRequest.Builder.appId: String
  get() = app_id
  set(value) { app_id = value }
fun SandboxCreateRequest.hasDefinition(): Boolean = definition != null
fun SandboxCreateRequest.Builder.setDefinition(value: Sandbox?): SandboxCreateRequest.Builder = apply { definition = value }
var SandboxCreateRequest.Builder.definition: Sandbox?
  get() = definition
  set(value) { definition = value }
val SandboxCreateRequest.environmentName: String get() = environment_name
fun SandboxCreateRequest.Builder.setEnvironmentName(value: String): SandboxCreateRequest.Builder = apply { environment_name = value }
var SandboxCreateRequest.Builder.environmentName: String
  get() = environment_name
  set(value) { environment_name = value }

fun SandboxCreateResponse.Companion.newBuilder(): SandboxCreateResponse.Builder = SandboxCreateResponse.Builder()
fun SandboxCreateResponse.Companion.getDefaultInstance(): SandboxCreateResponse = SandboxCreateResponse()
fun SandboxCreateResponse.toByteArray(): ByteArray = encode()
val SandboxCreateResponse.sandboxId: String get() = sandbox_id
fun SandboxCreateResponse.Builder.setSandboxId(value: String): SandboxCreateResponse.Builder = apply { sandbox_id = value }
var SandboxCreateResponse.Builder.sandboxId: String
  get() = sandbox_id
  set(value) { sandbox_id = value }

fun SandboxGetFromNameRequest.Companion.newBuilder(): SandboxGetFromNameRequest.Builder = SandboxGetFromNameRequest.Builder()
fun SandboxGetFromNameRequest.Companion.getDefaultInstance(): SandboxGetFromNameRequest = SandboxGetFromNameRequest()
fun SandboxGetFromNameRequest.toByteArray(): ByteArray = encode()
val SandboxGetFromNameRequest.sandboxName: String get() = sandbox_name
fun SandboxGetFromNameRequest.Builder.setSandboxName(value: String): SandboxGetFromNameRequest.Builder = apply { sandbox_name = value }
var SandboxGetFromNameRequest.Builder.sandboxName: String
  get() = sandbox_name
  set(value) { sandbox_name = value }
val SandboxGetFromNameRequest.environmentName: String get() = environment_name
fun SandboxGetFromNameRequest.Builder.setEnvironmentName(value: String): SandboxGetFromNameRequest.Builder = apply { environment_name = value }
var SandboxGetFromNameRequest.Builder.environmentName: String
  get() = environment_name
  set(value) { environment_name = value }
val SandboxGetFromNameRequest.appName: String get() = app_name
fun SandboxGetFromNameRequest.Builder.setAppName(value: String): SandboxGetFromNameRequest.Builder = apply { app_name = value }
var SandboxGetFromNameRequest.Builder.appName: String
  get() = app_name
  set(value) { app_name = value }

fun SandboxGetFromNameResponse.Companion.newBuilder(): SandboxGetFromNameResponse.Builder = SandboxGetFromNameResponse.Builder()
fun SandboxGetFromNameResponse.Companion.getDefaultInstance(): SandboxGetFromNameResponse = SandboxGetFromNameResponse()
fun SandboxGetFromNameResponse.toByteArray(): ByteArray = encode()
val SandboxGetFromNameResponse.sandboxId: String get() = sandbox_id
fun SandboxGetFromNameResponse.Builder.setSandboxId(value: String): SandboxGetFromNameResponse.Builder = apply { sandbox_id = value }
var SandboxGetFromNameResponse.Builder.sandboxId: String
  get() = sandbox_id
  set(value) { sandbox_id = value }

fun SandboxGetLogsRequest.Companion.newBuilder(): SandboxGetLogsRequest.Builder = SandboxGetLogsRequest.Builder()
fun SandboxGetLogsRequest.Companion.getDefaultInstance(): SandboxGetLogsRequest = SandboxGetLogsRequest()
fun SandboxGetLogsRequest.toByteArray(): ByteArray = encode()
val SandboxGetLogsRequest.sandboxId: String get() = sandbox_id
fun SandboxGetLogsRequest.Builder.setSandboxId(value: String): SandboxGetLogsRequest.Builder = apply { sandbox_id = value }
var SandboxGetLogsRequest.Builder.sandboxId: String
  get() = sandbox_id
  set(value) { sandbox_id = value }
val SandboxGetLogsRequest.fileDescriptor: FileDescriptor get() = file_descriptor
fun SandboxGetLogsRequest.Builder.setFileDescriptor(value: FileDescriptor): SandboxGetLogsRequest.Builder = apply { file_descriptor = value }
var SandboxGetLogsRequest.Builder.fileDescriptor: FileDescriptor
  get() = file_descriptor
  set(value) { file_descriptor = value }
fun SandboxGetLogsRequest.Builder.setTimeout(value: Float): SandboxGetLogsRequest.Builder = apply { timeout = value }
var SandboxGetLogsRequest.Builder.timeout: Float
  get() = timeout
  set(value) { timeout = value }
val SandboxGetLogsRequest.lastEntryId: String get() = last_entry_id
fun SandboxGetLogsRequest.Builder.setLastEntryId(value: String): SandboxGetLogsRequest.Builder = apply { last_entry_id = value }
var SandboxGetLogsRequest.Builder.lastEntryId: String
  get() = last_entry_id
  set(value) { last_entry_id = value }

fun SandboxGetTaskIdRequest.Companion.newBuilder(): SandboxGetTaskIdRequest.Builder = SandboxGetTaskIdRequest.Builder()
fun SandboxGetTaskIdRequest.Companion.getDefaultInstance(): SandboxGetTaskIdRequest = SandboxGetTaskIdRequest()
fun SandboxGetTaskIdRequest.toByteArray(): ByteArray = encode()
val SandboxGetTaskIdRequest.sandboxId: String get() = sandbox_id
fun SandboxGetTaskIdRequest.Builder.setSandboxId(value: String): SandboxGetTaskIdRequest.Builder = apply { sandbox_id = value }
var SandboxGetTaskIdRequest.Builder.sandboxId: String
  get() = sandbox_id
  set(value) { sandbox_id = value }
fun SandboxGetTaskIdRequest.hasTimeout(): Boolean = timeout != null
fun SandboxGetTaskIdRequest.Builder.setTimeout(value: Float?): SandboxGetTaskIdRequest.Builder = apply { timeout = value }
var SandboxGetTaskIdRequest.Builder.timeout: Float?
  get() = timeout
  set(value) { timeout = value }
val SandboxGetTaskIdRequest.waitUntilReady: Boolean get() = wait_until_ready
fun SandboxGetTaskIdRequest.Builder.setWaitUntilReady(value: Boolean): SandboxGetTaskIdRequest.Builder = apply { wait_until_ready = value }
var SandboxGetTaskIdRequest.Builder.waitUntilReady: Boolean
  get() = wait_until_ready
  set(value) { wait_until_ready = value }

fun SandboxGetTaskIdResponse.Companion.newBuilder(): SandboxGetTaskIdResponse.Builder = SandboxGetTaskIdResponse.Builder()
fun SandboxGetTaskIdResponse.Companion.getDefaultInstance(): SandboxGetTaskIdResponse = SandboxGetTaskIdResponse()
fun SandboxGetTaskIdResponse.toByteArray(): ByteArray = encode()
val SandboxGetTaskIdResponse.taskId: String? get() = task_id
fun SandboxGetTaskIdResponse.hasTaskId(): Boolean = task_id != null
fun SandboxGetTaskIdResponse.Builder.setTaskId(value: String?): SandboxGetTaskIdResponse.Builder = apply { task_id = value }
var SandboxGetTaskIdResponse.Builder.taskId: String?
  get() = task_id
  set(value) { task_id = value }
val SandboxGetTaskIdResponse.taskResult: GenericResult? get() = task_result
fun SandboxGetTaskIdResponse.hasTaskResult(): Boolean = task_result != null
fun SandboxGetTaskIdResponse.Builder.setTaskResult(value: GenericResult?): SandboxGetTaskIdResponse.Builder = apply { task_result = value }
var SandboxGetTaskIdResponse.Builder.taskResult: GenericResult?
  get() = task_result
  set(value) { task_result = value }

fun SandboxGetTunnelsRequest.Companion.newBuilder(): SandboxGetTunnelsRequest.Builder = SandboxGetTunnelsRequest.Builder()
fun SandboxGetTunnelsRequest.Companion.getDefaultInstance(): SandboxGetTunnelsRequest = SandboxGetTunnelsRequest()
fun SandboxGetTunnelsRequest.toByteArray(): ByteArray = encode()
val SandboxGetTunnelsRequest.sandboxId: String get() = sandbox_id
fun SandboxGetTunnelsRequest.Builder.setSandboxId(value: String): SandboxGetTunnelsRequest.Builder = apply { sandbox_id = value }
var SandboxGetTunnelsRequest.Builder.sandboxId: String
  get() = sandbox_id
  set(value) { sandbox_id = value }
fun SandboxGetTunnelsRequest.Builder.setTimeout(value: Float): SandboxGetTunnelsRequest.Builder = apply { timeout = value }
var SandboxGetTunnelsRequest.Builder.timeout: Float
  get() = timeout
  set(value) { timeout = value }

fun SandboxGetTunnelsResponse.Companion.newBuilder(): SandboxGetTunnelsResponse.Builder = SandboxGetTunnelsResponse.Builder()
fun SandboxGetTunnelsResponse.Companion.getDefaultInstance(): SandboxGetTunnelsResponse = SandboxGetTunnelsResponse()
fun SandboxGetTunnelsResponse.toByteArray(): ByteArray = encode()
fun SandboxGetTunnelsResponse.hasResult(): Boolean = result != null
fun SandboxGetTunnelsResponse.Builder.setResult(value: GenericResult?): SandboxGetTunnelsResponse.Builder = apply { result = value }
var SandboxGetTunnelsResponse.Builder.result: GenericResult?
  get() = result
  set(value) { result = value }
val SandboxGetTunnelsResponse.tunnelsList: List<TunnelData> get() = tunnels
val SandboxGetTunnelsResponse.tunnelsCount: Int get() = tunnels.size
fun SandboxGetTunnelsResponse.Builder.addTunnels(value: TunnelData): SandboxGetTunnelsResponse.Builder = apply { tunnels = tunnels + value }
fun SandboxGetTunnelsResponse.Builder.addAllTunnels(items: Iterable<TunnelData>): SandboxGetTunnelsResponse.Builder = apply { this.tunnels = this.tunnels + items }
fun SandboxGetTunnelsResponse.Builder.clearTunnels(): SandboxGetTunnelsResponse.Builder = apply { tunnels = emptyList() }

fun SandboxInfo.Companion.newBuilder(): SandboxInfo.Builder = SandboxInfo.Builder()
fun SandboxInfo.Companion.getDefaultInstance(): SandboxInfo = SandboxInfo()
fun SandboxInfo.toByteArray(): ByteArray = encode()
fun SandboxInfo.Builder.setId(value: String): SandboxInfo.Builder = apply { id = value }
var SandboxInfo.Builder.id: String
  get() = id
  set(value) { id = value }
val SandboxInfo.createdAt: Double get() = created_at
fun SandboxInfo.Builder.setCreatedAt(value: Double): SandboxInfo.Builder = apply { created_at = value }
var SandboxInfo.Builder.createdAt: Double
  get() = created_at
  set(value) { created_at = value }
val SandboxInfo.taskInfo: TaskInfo? get() = task_info
fun SandboxInfo.hasTaskInfo(): Boolean = task_info != null
fun SandboxInfo.Builder.setTaskInfo(value: TaskInfo?): SandboxInfo.Builder = apply { task_info = value }
var SandboxInfo.Builder.taskInfo: TaskInfo?
  get() = task_info
  set(value) { task_info = value }
val SandboxInfo.appId: String get() = app_id
fun SandboxInfo.Builder.setAppId(value: String): SandboxInfo.Builder = apply { app_id = value }
var SandboxInfo.Builder.appId: String
  get() = app_id
  set(value) { app_id = value }
fun SandboxInfo.Builder.setName(value: String): SandboxInfo.Builder = apply { name = value }
var SandboxInfo.Builder.name: String
  get() = name
  set(value) { name = value }
val SandboxInfo.imageId: String get() = image_id
fun SandboxInfo.Builder.setImageId(value: String): SandboxInfo.Builder = apply { image_id = value }
var SandboxInfo.Builder.imageId: String
  get() = image_id
  set(value) { image_id = value }
val SandboxInfo.resourceInfo: ResourceInfo? get() = resource_info
fun SandboxInfo.hasResourceInfo(): Boolean = resource_info != null
fun SandboxInfo.Builder.setResourceInfo(value: ResourceInfo?): SandboxInfo.Builder = apply { resource_info = value }
var SandboxInfo.Builder.resourceInfo: ResourceInfo?
  get() = resource_info
  set(value) { resource_info = value }
val SandboxInfo.timeoutSecs: Int get() = timeout_secs
fun SandboxInfo.Builder.setTimeoutSecs(value: Int): SandboxInfo.Builder = apply { timeout_secs = value }
var SandboxInfo.Builder.timeoutSecs: Int
  get() = timeout_secs
  set(value) { timeout_secs = value }
val SandboxInfo.idleTimeoutSecs: Int? get() = idle_timeout_secs
fun SandboxInfo.hasIdleTimeoutSecs(): Boolean = idle_timeout_secs != null
fun SandboxInfo.Builder.setIdleTimeoutSecs(value: Int?): SandboxInfo.Builder = apply { idle_timeout_secs = value }
var SandboxInfo.Builder.idleTimeoutSecs: Int?
  get() = idle_timeout_secs
  set(value) { idle_timeout_secs = value }
val SandboxInfo.tagsList: List<SandboxTag> get() = tags
val SandboxInfo.tagsCount: Int get() = tags.size
fun SandboxInfo.Builder.addTags(value: SandboxTag): SandboxInfo.Builder = apply { tags = tags + value }
fun SandboxInfo.Builder.addAllTags(items: Iterable<SandboxTag>): SandboxInfo.Builder = apply { this.tags = this.tags + items }
fun SandboxInfo.Builder.clearTags(): SandboxInfo.Builder = apply { tags = emptyList() }
val SandboxInfo.regionsList: List<String> get() = regions
val SandboxInfo.regionsCount: Int get() = regions.size
fun SandboxInfo.Builder.addRegions(value: String): SandboxInfo.Builder = apply { regions = regions + value }
fun SandboxInfo.Builder.addAllRegions(items: Iterable<String>): SandboxInfo.Builder = apply { this.regions = this.regions + items }
fun SandboxInfo.Builder.clearRegions(): SandboxInfo.Builder = apply { regions = emptyList() }

fun SandboxListRequest.Companion.newBuilder(): SandboxListRequest.Builder = SandboxListRequest.Builder()
fun SandboxListRequest.Companion.getDefaultInstance(): SandboxListRequest = SandboxListRequest()
fun SandboxListRequest.toByteArray(): ByteArray = encode()
val SandboxListRequest.appId: String get() = app_id
fun SandboxListRequest.Builder.setAppId(value: String): SandboxListRequest.Builder = apply { app_id = value }
var SandboxListRequest.Builder.appId: String
  get() = app_id
  set(value) { app_id = value }
val SandboxListRequest.beforeTimestamp: Double get() = before_timestamp
fun SandboxListRequest.Builder.setBeforeTimestamp(value: Double): SandboxListRequest.Builder = apply { before_timestamp = value }
var SandboxListRequest.Builder.beforeTimestamp: Double
  get() = before_timestamp
  set(value) { before_timestamp = value }
val SandboxListRequest.environmentName: String get() = environment_name
fun SandboxListRequest.Builder.setEnvironmentName(value: String): SandboxListRequest.Builder = apply { environment_name = value }
var SandboxListRequest.Builder.environmentName: String
  get() = environment_name
  set(value) { environment_name = value }
val SandboxListRequest.includeFinished: Boolean get() = include_finished
fun SandboxListRequest.Builder.setIncludeFinished(value: Boolean): SandboxListRequest.Builder = apply { include_finished = value }
var SandboxListRequest.Builder.includeFinished: Boolean
  get() = include_finished
  set(value) { include_finished = value }
val SandboxListRequest.tagsList: List<SandboxTag> get() = tags
val SandboxListRequest.tagsCount: Int get() = tags.size
fun SandboxListRequest.Builder.addTags(value: SandboxTag): SandboxListRequest.Builder = apply { tags = tags + value }
fun SandboxListRequest.Builder.addAllTags(items: Iterable<SandboxTag>): SandboxListRequest.Builder = apply { this.tags = this.tags + items }
fun SandboxListRequest.Builder.clearTags(): SandboxListRequest.Builder = apply { tags = emptyList() }

fun SandboxListResponse.Companion.newBuilder(): SandboxListResponse.Builder = SandboxListResponse.Builder()
fun SandboxListResponse.Companion.getDefaultInstance(): SandboxListResponse = SandboxListResponse()
fun SandboxListResponse.toByteArray(): ByteArray = encode()
val SandboxListResponse.sandboxesList: List<SandboxInfo> get() = sandboxes
val SandboxListResponse.sandboxesCount: Int get() = sandboxes.size
fun SandboxListResponse.Builder.addSandboxes(value: SandboxInfo): SandboxListResponse.Builder = apply { sandboxes = sandboxes + value }
fun SandboxListResponse.Builder.addAllSandboxes(items: Iterable<SandboxInfo>): SandboxListResponse.Builder = apply { this.sandboxes = this.sandboxes + items }
fun SandboxListResponse.Builder.clearSandboxes(): SandboxListResponse.Builder = apply { sandboxes = emptyList() }

fun SandboxSnapshotFsRequest.Companion.newBuilder(): SandboxSnapshotFsRequest.Builder = SandboxSnapshotFsRequest.Builder()
fun SandboxSnapshotFsRequest.Companion.getDefaultInstance(): SandboxSnapshotFsRequest = SandboxSnapshotFsRequest()
fun SandboxSnapshotFsRequest.toByteArray(): ByteArray = encode()
val SandboxSnapshotFsRequest.sandboxId: String get() = sandbox_id
fun SandboxSnapshotFsRequest.Builder.setSandboxId(value: String): SandboxSnapshotFsRequest.Builder = apply { sandbox_id = value }
var SandboxSnapshotFsRequest.Builder.sandboxId: String
  get() = sandbox_id
  set(value) { sandbox_id = value }
fun SandboxSnapshotFsRequest.Builder.setTimeout(value: Float): SandboxSnapshotFsRequest.Builder = apply { timeout = value }
var SandboxSnapshotFsRequest.Builder.timeout: Float
  get() = timeout
  set(value) { timeout = value }

fun SandboxSnapshotFsResponse.Companion.newBuilder(): SandboxSnapshotFsResponse.Builder = SandboxSnapshotFsResponse.Builder()
fun SandboxSnapshotFsResponse.Companion.getDefaultInstance(): SandboxSnapshotFsResponse = SandboxSnapshotFsResponse()
fun SandboxSnapshotFsResponse.toByteArray(): ByteArray = encode()
val SandboxSnapshotFsResponse.imageId: String get() = image_id
fun SandboxSnapshotFsResponse.Builder.setImageId(value: String): SandboxSnapshotFsResponse.Builder = apply { image_id = value }
var SandboxSnapshotFsResponse.Builder.imageId: String
  get() = image_id
  set(value) { image_id = value }
fun SandboxSnapshotFsResponse.hasResult(): Boolean = result != null
fun SandboxSnapshotFsResponse.Builder.setResult(value: GenericResult?): SandboxSnapshotFsResponse.Builder = apply { result = value }
var SandboxSnapshotFsResponse.Builder.result: GenericResult?
  get() = result
  set(value) { result = value }
val SandboxSnapshotFsResponse.imageMetadata: ImageMetadata? get() = image_metadata
fun SandboxSnapshotFsResponse.hasImageMetadata(): Boolean = image_metadata != null
fun SandboxSnapshotFsResponse.Builder.setImageMetadata(value: ImageMetadata?): SandboxSnapshotFsResponse.Builder = apply { image_metadata = value }
var SandboxSnapshotFsResponse.Builder.imageMetadata: ImageMetadata?
  get() = image_metadata
  set(value) { image_metadata = value }

fun SandboxStdinWriteRequest.Companion.newBuilder(): SandboxStdinWriteRequest.Builder = SandboxStdinWriteRequest.Builder()
fun SandboxStdinWriteRequest.Companion.getDefaultInstance(): SandboxStdinWriteRequest = SandboxStdinWriteRequest()
fun SandboxStdinWriteRequest.toByteArray(): ByteArray = encode()
val SandboxStdinWriteRequest.sandboxId: String get() = sandbox_id
fun SandboxStdinWriteRequest.Builder.setSandboxId(value: String): SandboxStdinWriteRequest.Builder = apply { sandbox_id = value }
var SandboxStdinWriteRequest.Builder.sandboxId: String
  get() = sandbox_id
  set(value) { sandbox_id = value }
fun SandboxStdinWriteRequest.Builder.setInput(value: ByteString): SandboxStdinWriteRequest.Builder = apply { input = value }
var SandboxStdinWriteRequest.Builder.input: ByteString
  get() = input
  set(value) { input = value }
fun SandboxStdinWriteRequest.Builder.setInput(value: ProtoByteString): SandboxStdinWriteRequest.Builder = apply { input = value.toByteArray().toByteString() }
fun SandboxStdinWriteRequest.Builder.setIndex(value: Int): SandboxStdinWriteRequest.Builder = apply { index = value }
var SandboxStdinWriteRequest.Builder.index: Int
  get() = index
  set(value) { index = value }
fun SandboxStdinWriteRequest.Builder.setEof(value: Boolean): SandboxStdinWriteRequest.Builder = apply { eof = value }
var SandboxStdinWriteRequest.Builder.eof: Boolean
  get() = eof
  set(value) { eof = value }

fun SandboxTag.Companion.newBuilder(): SandboxTag.Builder = SandboxTag.Builder()
fun SandboxTag.Companion.getDefaultInstance(): SandboxTag = SandboxTag()
fun SandboxTag.toByteArray(): ByteArray = encode()
val SandboxTag.tagName: String get() = tag_name
fun SandboxTag.Builder.setTagName(value: String): SandboxTag.Builder = apply { tag_name = value }
var SandboxTag.Builder.tagName: String
  get() = tag_name
  set(value) { tag_name = value }
val SandboxTag.tagValue: String get() = tag_value
fun SandboxTag.Builder.setTagValue(value: String): SandboxTag.Builder = apply { tag_value = value }
var SandboxTag.Builder.tagValue: String
  get() = tag_value
  set(value) { tag_value = value }

fun SandboxTagsGetRequest.Companion.newBuilder(): SandboxTagsGetRequest.Builder = SandboxTagsGetRequest.Builder()
fun SandboxTagsGetRequest.Companion.getDefaultInstance(): SandboxTagsGetRequest = SandboxTagsGetRequest()
fun SandboxTagsGetRequest.toByteArray(): ByteArray = encode()
val SandboxTagsGetRequest.sandboxId: String get() = sandbox_id
fun SandboxTagsGetRequest.Builder.setSandboxId(value: String): SandboxTagsGetRequest.Builder = apply { sandbox_id = value }
var SandboxTagsGetRequest.Builder.sandboxId: String
  get() = sandbox_id
  set(value) { sandbox_id = value }

fun SandboxTagsGetResponse.Companion.newBuilder(): SandboxTagsGetResponse.Builder = SandboxTagsGetResponse.Builder()
fun SandboxTagsGetResponse.Companion.getDefaultInstance(): SandboxTagsGetResponse = SandboxTagsGetResponse()
fun SandboxTagsGetResponse.toByteArray(): ByteArray = encode()
val SandboxTagsGetResponse.tagsList: List<SandboxTag> get() = tags
val SandboxTagsGetResponse.tagsCount: Int get() = tags.size
fun SandboxTagsGetResponse.Builder.addTags(value: SandboxTag): SandboxTagsGetResponse.Builder = apply { tags = tags + value }
fun SandboxTagsGetResponse.Builder.addAllTags(items: Iterable<SandboxTag>): SandboxTagsGetResponse.Builder = apply { this.tags = this.tags + items }
fun SandboxTagsGetResponse.Builder.clearTags(): SandboxTagsGetResponse.Builder = apply { tags = emptyList() }

fun SandboxTagsSetRequest.Companion.newBuilder(): SandboxTagsSetRequest.Builder = SandboxTagsSetRequest.Builder()
fun SandboxTagsSetRequest.Companion.getDefaultInstance(): SandboxTagsSetRequest = SandboxTagsSetRequest()
fun SandboxTagsSetRequest.toByteArray(): ByteArray = encode()
val SandboxTagsSetRequest.environmentName: String get() = environment_name
fun SandboxTagsSetRequest.Builder.setEnvironmentName(value: String): SandboxTagsSetRequest.Builder = apply { environment_name = value }
var SandboxTagsSetRequest.Builder.environmentName: String
  get() = environment_name
  set(value) { environment_name = value }
val SandboxTagsSetRequest.sandboxId: String get() = sandbox_id
fun SandboxTagsSetRequest.Builder.setSandboxId(value: String): SandboxTagsSetRequest.Builder = apply { sandbox_id = value }
var SandboxTagsSetRequest.Builder.sandboxId: String
  get() = sandbox_id
  set(value) { sandbox_id = value }
val SandboxTagsSetRequest.tagsList: List<SandboxTag> get() = tags
val SandboxTagsSetRequest.tagsCount: Int get() = tags.size
fun SandboxTagsSetRequest.Builder.addTags(value: SandboxTag): SandboxTagsSetRequest.Builder = apply { tags = tags + value }
fun SandboxTagsSetRequest.Builder.addAllTags(items: Iterable<SandboxTag>): SandboxTagsSetRequest.Builder = apply { this.tags = this.tags + items }
fun SandboxTagsSetRequest.Builder.clearTags(): SandboxTagsSetRequest.Builder = apply { tags = emptyList() }

fun SandboxTerminateRequest.Companion.newBuilder(): SandboxTerminateRequest.Builder = SandboxTerminateRequest.Builder()
fun SandboxTerminateRequest.Companion.getDefaultInstance(): SandboxTerminateRequest = SandboxTerminateRequest()
fun SandboxTerminateRequest.toByteArray(): ByteArray = encode()
val SandboxTerminateRequest.sandboxId: String get() = sandbox_id
fun SandboxTerminateRequest.Builder.setSandboxId(value: String): SandboxTerminateRequest.Builder = apply { sandbox_id = value }
var SandboxTerminateRequest.Builder.sandboxId: String
  get() = sandbox_id
  set(value) { sandbox_id = value }

fun SandboxTerminateResponse.Companion.newBuilder(): SandboxTerminateResponse.Builder = SandboxTerminateResponse.Builder()
fun SandboxTerminateResponse.Companion.getDefaultInstance(): SandboxTerminateResponse = SandboxTerminateResponse()
fun SandboxTerminateResponse.toByteArray(): ByteArray = encode()
val SandboxTerminateResponse.existingResult: GenericResult? get() = existing_result
fun SandboxTerminateResponse.hasExistingResult(): Boolean = existing_result != null
fun SandboxTerminateResponse.Builder.setExistingResult(value: GenericResult?): SandboxTerminateResponse.Builder = apply { existing_result = value }
var SandboxTerminateResponse.Builder.existingResult: GenericResult?
  get() = existing_result
  set(value) { existing_result = value }

fun SandboxWaitRequest.Companion.newBuilder(): SandboxWaitRequest.Builder = SandboxWaitRequest.Builder()
fun SandboxWaitRequest.Companion.getDefaultInstance(): SandboxWaitRequest = SandboxWaitRequest()
fun SandboxWaitRequest.toByteArray(): ByteArray = encode()
val SandboxWaitRequest.sandboxId: String get() = sandbox_id
fun SandboxWaitRequest.Builder.setSandboxId(value: String): SandboxWaitRequest.Builder = apply { sandbox_id = value }
var SandboxWaitRequest.Builder.sandboxId: String
  get() = sandbox_id
  set(value) { sandbox_id = value }
fun SandboxWaitRequest.Builder.setTimeout(value: Float): SandboxWaitRequest.Builder = apply { timeout = value }
var SandboxWaitRequest.Builder.timeout: Float
  get() = timeout
  set(value) { timeout = value }

fun SandboxWaitResponse.Companion.newBuilder(): SandboxWaitResponse.Builder = SandboxWaitResponse.Builder()
fun SandboxWaitResponse.Companion.getDefaultInstance(): SandboxWaitResponse = SandboxWaitResponse()
fun SandboxWaitResponse.toByteArray(): ByteArray = encode()
fun SandboxWaitResponse.hasResult(): Boolean = result != null
fun SandboxWaitResponse.Builder.setResult(value: GenericResult?): SandboxWaitResponse.Builder = apply { result = value }
var SandboxWaitResponse.Builder.result: GenericResult?
  get() = result
  set(value) { result = value }

fun Schedule.Companion.newBuilder(): Schedule.Builder = Schedule.Builder()
fun Schedule.Companion.getDefaultInstance(): Schedule = Schedule()
fun Schedule.toByteArray(): ByteArray = encode()
fun Schedule.hasCron(): Boolean = cron != null
fun Schedule.Builder.setCron(value: Schedule.Cron?): Schedule.Builder = apply { cron = value }
var Schedule.Builder.cron: Schedule.Cron?
  get() = cron
  set(value) { cron = value }
fun Schedule.hasPeriod(): Boolean = period != null
fun Schedule.Builder.setPeriod(value: Schedule.Period?): Schedule.Builder = apply { period = value }
var Schedule.Builder.period: Schedule.Period?
  get() = period
  set(value) { period = value }

fun SchedulerPlacement.Companion.newBuilder(): SchedulerPlacement.Builder = SchedulerPlacement.Builder()
fun SchedulerPlacement.Companion.getDefaultInstance(): SchedulerPlacement = SchedulerPlacement()
fun SchedulerPlacement.toByteArray(): ByteArray = encode()
val SchedulerPlacement.Zone: String? get() = _zone
fun SchedulerPlacement.hasZone(): Boolean = _zone != null
fun SchedulerPlacement.Builder.setZone(value: String?): SchedulerPlacement.Builder = apply { _zone = value }
var SchedulerPlacement.Builder.Zone: String?
  get() = _zone
  set(value) { _zone = value }
val SchedulerPlacement.Lifecycle: String? get() = _lifecycle
fun SchedulerPlacement.hasLifecycle(): Boolean = _lifecycle != null
fun SchedulerPlacement.Builder.setLifecycle(value: String?): SchedulerPlacement.Builder = apply { _lifecycle = value }
var SchedulerPlacement.Builder.Lifecycle: String?
  get() = _lifecycle
  set(value) { _lifecycle = value }
fun SchedulerPlacement.Builder.setNonpreemptible(value: Boolean): SchedulerPlacement.Builder = apply { nonpreemptible = value }
var SchedulerPlacement.Builder.nonpreemptible: Boolean
  get() = nonpreemptible
  set(value) { nonpreemptible = value }
val SchedulerPlacement.regionsList: List<String> get() = regions
val SchedulerPlacement.regionsCount: Int get() = regions.size
fun SchedulerPlacement.Builder.addRegions(value: String): SchedulerPlacement.Builder = apply { regions = regions + value }
fun SchedulerPlacement.Builder.addAllRegions(items: Iterable<String>): SchedulerPlacement.Builder = apply { this.regions = this.regions + items }
fun SchedulerPlacement.Builder.clearRegions(): SchedulerPlacement.Builder = apply { regions = emptyList() }
val SchedulerPlacement.InstanceTypes: List<String> get() = _instance_types
val SchedulerPlacement.InstanceTypesList: List<String> get() = _instance_types
val SchedulerPlacement.InstanceTypesCount: Int get() = _instance_types.size
fun SchedulerPlacement.Builder.addInstanceTypes(value: String): SchedulerPlacement.Builder = apply { _instance_types = _instance_types + value }
fun SchedulerPlacement.Builder.addAllInstanceTypes(items: Iterable<String>): SchedulerPlacement.Builder = apply { this._instance_types = this._instance_types + items }
fun SchedulerPlacement.Builder.clearInstanceTypes(): SchedulerPlacement.Builder = apply { _instance_types = emptyList() }

fun SecretDeleteRequest.Companion.newBuilder(): SecretDeleteRequest.Builder = SecretDeleteRequest.Builder()
fun SecretDeleteRequest.Companion.getDefaultInstance(): SecretDeleteRequest = SecretDeleteRequest()
fun SecretDeleteRequest.toByteArray(): ByteArray = encode()
val SecretDeleteRequest.secretId: String get() = secret_id
fun SecretDeleteRequest.Builder.setSecretId(value: String): SecretDeleteRequest.Builder = apply { secret_id = value }
var SecretDeleteRequest.Builder.secretId: String
  get() = secret_id
  set(value) { secret_id = value }

fun SecretGetOrCreateRequest.Companion.newBuilder(): SecretGetOrCreateRequest.Builder = SecretGetOrCreateRequest.Builder()
fun SecretGetOrCreateRequest.Companion.getDefaultInstance(): SecretGetOrCreateRequest = SecretGetOrCreateRequest()
fun SecretGetOrCreateRequest.toByteArray(): ByteArray = encode()
val SecretGetOrCreateRequest.deploymentName: String get() = deployment_name
fun SecretGetOrCreateRequest.Builder.setDeploymentName(value: String): SecretGetOrCreateRequest.Builder = apply { deployment_name = value }
var SecretGetOrCreateRequest.Builder.deploymentName: String
  get() = deployment_name
  set(value) { deployment_name = value }
val SecretGetOrCreateRequest.environmentName: String get() = environment_name
fun SecretGetOrCreateRequest.Builder.setEnvironmentName(value: String): SecretGetOrCreateRequest.Builder = apply { environment_name = value }
var SecretGetOrCreateRequest.Builder.environmentName: String
  get() = environment_name
  set(value) { environment_name = value }
val SecretGetOrCreateRequest.objectCreationType: ObjectCreationType get() = object_creation_type
fun SecretGetOrCreateRequest.Builder.setObjectCreationType(value: ObjectCreationType): SecretGetOrCreateRequest.Builder = apply { object_creation_type = value }
var SecretGetOrCreateRequest.Builder.objectCreationType: ObjectCreationType
  get() = object_creation_type
  set(value) { object_creation_type = value }
val SecretGetOrCreateRequest.appId: String get() = app_id
fun SecretGetOrCreateRequest.Builder.setAppId(value: String): SecretGetOrCreateRequest.Builder = apply { app_id = value }
var SecretGetOrCreateRequest.Builder.appId: String
  get() = app_id
  set(value) { app_id = value }
val SecretGetOrCreateRequest.envDict: Map<String, String> get() = env_dict
val SecretGetOrCreateRequest.envDictMap: Map<String, String> get() = env_dict
fun SecretGetOrCreateRequest.Builder.putEnvDict(key: String, value: String): SecretGetOrCreateRequest.Builder = apply { env_dict = env_dict + mapOf(key to value) }
fun SecretGetOrCreateRequest.Builder.putAllEnvDict(entries: Map<String, String>): SecretGetOrCreateRequest.Builder = apply { env_dict = env_dict + entries }
fun SecretGetOrCreateRequest.Builder.clearEnvDict(): SecretGetOrCreateRequest.Builder = apply { env_dict = emptyMap() }
val SecretGetOrCreateRequest.requiredKeys: List<String> get() = required_keys
val SecretGetOrCreateRequest.requiredKeysList: List<String> get() = required_keys
val SecretGetOrCreateRequest.requiredKeysCount: Int get() = required_keys.size
fun SecretGetOrCreateRequest.Builder.addRequiredKeys(value: String): SecretGetOrCreateRequest.Builder = apply { required_keys = required_keys + value }
fun SecretGetOrCreateRequest.Builder.addAllRequiredKeys(items: Iterable<String>): SecretGetOrCreateRequest.Builder = apply { this.required_keys = this.required_keys + items }
fun SecretGetOrCreateRequest.Builder.clearRequiredKeys(): SecretGetOrCreateRequest.Builder = apply { required_keys = emptyList() }

fun SecretGetOrCreateResponse.Companion.newBuilder(): SecretGetOrCreateResponse.Builder = SecretGetOrCreateResponse.Builder()
fun SecretGetOrCreateResponse.Companion.getDefaultInstance(): SecretGetOrCreateResponse = SecretGetOrCreateResponse()
fun SecretGetOrCreateResponse.toByteArray(): ByteArray = encode()
val SecretGetOrCreateResponse.secretId: String get() = secret_id
fun SecretGetOrCreateResponse.Builder.setSecretId(value: String): SecretGetOrCreateResponse.Builder = apply { secret_id = value }
var SecretGetOrCreateResponse.Builder.secretId: String
  get() = secret_id
  set(value) { secret_id = value }
fun SecretGetOrCreateResponse.hasMetadata(): Boolean = metadata != null
fun SecretGetOrCreateResponse.Builder.setMetadata(value: SecretMetadata?): SecretGetOrCreateResponse.Builder = apply { metadata = value }
var SecretGetOrCreateResponse.Builder.metadata: SecretMetadata?
  get() = metadata
  set(value) { metadata = value }

fun SecretMetadata.Companion.newBuilder(): SecretMetadata.Builder = SecretMetadata.Builder()
fun SecretMetadata.Companion.getDefaultInstance(): SecretMetadata = SecretMetadata()
fun SecretMetadata.toByteArray(): ByteArray = encode()
fun SecretMetadata.Builder.setName(value: String): SecretMetadata.Builder = apply { name = value }
var SecretMetadata.Builder.name: String
  get() = name
  set(value) { name = value }
val SecretMetadata.creationInfo: CreationInfo? get() = creation_info
fun SecretMetadata.hasCreationInfo(): Boolean = creation_info != null
fun SecretMetadata.Builder.setCreationInfo(value: CreationInfo?): SecretMetadata.Builder = apply { creation_info = value }
var SecretMetadata.Builder.creationInfo: CreationInfo?
  get() = creation_info
  set(value) { creation_info = value }

fun SharedVolumeMount.Companion.newBuilder(): SharedVolumeMount.Builder = SharedVolumeMount.Builder()
fun SharedVolumeMount.Companion.getDefaultInstance(): SharedVolumeMount = SharedVolumeMount()
fun SharedVolumeMount.toByteArray(): ByteArray = encode()
val SharedVolumeMount.mountPath: String get() = mount_path
fun SharedVolumeMount.Builder.setMountPath(value: String): SharedVolumeMount.Builder = apply { mount_path = value }
var SharedVolumeMount.Builder.mountPath: String
  get() = mount_path
  set(value) { mount_path = value }
val SharedVolumeMount.sharedVolumeId: String get() = shared_volume_id
fun SharedVolumeMount.Builder.setSharedVolumeId(value: String): SharedVolumeMount.Builder = apply { shared_volume_id = value }
var SharedVolumeMount.Builder.sharedVolumeId: String
  get() = shared_volume_id
  set(value) { shared_volume_id = value }
val SharedVolumeMount.cloudProvider: CloudProvider get() = cloud_provider
fun SharedVolumeMount.Builder.setCloudProvider(value: CloudProvider): SharedVolumeMount.Builder = apply { cloud_provider = value }
var SharedVolumeMount.Builder.cloudProvider: CloudProvider
  get() = cloud_provider
  set(value) { cloud_provider = value }

fun SystemErrorMessage.Companion.newBuilder(): SystemErrorMessage.Builder = SystemErrorMessage.Builder()
fun SystemErrorMessage.Companion.getDefaultInstance(): SystemErrorMessage = SystemErrorMessage()
fun SystemErrorMessage.toByteArray(): ByteArray = encode()
val SystemErrorMessage.errorCode: SystemErrorCode get() = error_code
fun SystemErrorMessage.Builder.setErrorCode(value: SystemErrorCode): SystemErrorMessage.Builder = apply { error_code = value }
var SystemErrorMessage.Builder.errorCode: SystemErrorCode
  get() = error_code
  set(value) { error_code = value }
val SystemErrorMessage.errorMessage: String get() = error_message
fun SystemErrorMessage.Builder.setErrorMessage(value: String): SystemErrorMessage.Builder = apply { error_message = value }
var SystemErrorMessage.Builder.errorMessage: String
  get() = error_message
  set(value) { error_message = value }

fun TaskGetCommandRouterAccessRequest.Companion.newBuilder(): TaskGetCommandRouterAccessRequest.Builder = TaskGetCommandRouterAccessRequest.Builder()
fun TaskGetCommandRouterAccessRequest.Companion.getDefaultInstance(): TaskGetCommandRouterAccessRequest = TaskGetCommandRouterAccessRequest()
fun TaskGetCommandRouterAccessRequest.toByteArray(): ByteArray = encode()
val TaskGetCommandRouterAccessRequest.taskId: String get() = task_id
fun TaskGetCommandRouterAccessRequest.Builder.setTaskId(value: String): TaskGetCommandRouterAccessRequest.Builder = apply { task_id = value }
var TaskGetCommandRouterAccessRequest.Builder.taskId: String
  get() = task_id
  set(value) { task_id = value }

fun TaskGetCommandRouterAccessResponse.Companion.newBuilder(): TaskGetCommandRouterAccessResponse.Builder = TaskGetCommandRouterAccessResponse.Builder()
fun TaskGetCommandRouterAccessResponse.Companion.getDefaultInstance(): TaskGetCommandRouterAccessResponse = TaskGetCommandRouterAccessResponse()
fun TaskGetCommandRouterAccessResponse.toByteArray(): ByteArray = encode()
fun TaskGetCommandRouterAccessResponse.Builder.setJwt(value: String): TaskGetCommandRouterAccessResponse.Builder = apply { jwt = value }
var TaskGetCommandRouterAccessResponse.Builder.jwt: String
  get() = jwt
  set(value) { jwt = value }
fun TaskGetCommandRouterAccessResponse.Builder.setUrl(value: String): TaskGetCommandRouterAccessResponse.Builder = apply { url = value }
var TaskGetCommandRouterAccessResponse.Builder.url: String
  get() = url
  set(value) { url = value }

fun TaskInfo.Companion.newBuilder(): TaskInfo.Builder = TaskInfo.Builder()
fun TaskInfo.Companion.getDefaultInstance(): TaskInfo = TaskInfo()
fun TaskInfo.toByteArray(): ByteArray = encode()
fun TaskInfo.Builder.setId(value: String): TaskInfo.Builder = apply { id = value }
var TaskInfo.Builder.id: String
  get() = id
  set(value) { id = value }
val TaskInfo.startedAt: Double get() = started_at
fun TaskInfo.Builder.setStartedAt(value: Double): TaskInfo.Builder = apply { started_at = value }
var TaskInfo.Builder.startedAt: Double
  get() = started_at
  set(value) { started_at = value }
val TaskInfo.finishedAt: Double get() = finished_at
fun TaskInfo.Builder.setFinishedAt(value: Double): TaskInfo.Builder = apply { finished_at = value }
var TaskInfo.Builder.finishedAt: Double
  get() = finished_at
  set(value) { finished_at = value }
fun TaskInfo.hasResult(): Boolean = result != null
fun TaskInfo.Builder.setResult(value: GenericResult?): TaskInfo.Builder = apply { result = value }
var TaskInfo.Builder.result: GenericResult?
  get() = result
  set(value) { result = value }
val TaskInfo.enqueuedAt: Double get() = enqueued_at
fun TaskInfo.Builder.setEnqueuedAt(value: Double): TaskInfo.Builder = apply { enqueued_at = value }
var TaskInfo.Builder.enqueuedAt: Double
  get() = enqueued_at
  set(value) { enqueued_at = value }
val TaskInfo.gpuType: String get() = gpu_type
fun TaskInfo.Builder.setGpuType(value: String): TaskInfo.Builder = apply { gpu_type = value }
var TaskInfo.Builder.gpuType: String
  get() = gpu_type
  set(value) { gpu_type = value }
val TaskInfo.sandboxId: String get() = sandbox_id
fun TaskInfo.Builder.setSandboxId(value: String): TaskInfo.Builder = apply { sandbox_id = value }
var TaskInfo.Builder.sandboxId: String
  get() = sandbox_id
  set(value) { sandbox_id = value }
val TaskInfo.snapshotBehavior: TaskSnapshotBehavior get() = snapshot_behavior
fun TaskInfo.Builder.setSnapshotBehavior(value: TaskSnapshotBehavior): TaskInfo.Builder = apply { snapshot_behavior = value }
var TaskInfo.Builder.snapshotBehavior: TaskSnapshotBehavior
  get() = snapshot_behavior
  set(value) { snapshot_behavior = value }
val TaskInfo.gpuConfig: GPUConfig? get() = gpu_config
fun TaskInfo.hasGpuConfig(): Boolean = gpu_config != null
fun TaskInfo.Builder.setGpuConfig(value: GPUConfig?): TaskInfo.Builder = apply { gpu_config = value }
var TaskInfo.Builder.gpuConfig: GPUConfig?
  get() = gpu_config
  set(value) { gpu_config = value }

fun TaskLogs.Companion.newBuilder(): TaskLogs.Builder = TaskLogs.Builder()
fun TaskLogs.Companion.getDefaultInstance(): TaskLogs = TaskLogs()
fun TaskLogs.toByteArray(): ByteArray = encode()
val TaskLogs.data: String get() = data_
fun TaskLogs.Builder.setData(value: String): TaskLogs.Builder = apply { data_ = value }
var TaskLogs.Builder.data: String
  get() = data_
  set(value) { data_ = value }
val TaskLogs.taskState: TaskState get() = task_state
fun TaskLogs.Builder.setTaskState(value: TaskState): TaskLogs.Builder = apply { task_state = value }
var TaskLogs.Builder.taskState: TaskState
  get() = task_state
  set(value) { task_state = value }
fun TaskLogs.Builder.setTimestamp(value: Double): TaskLogs.Builder = apply { timestamp = value }
var TaskLogs.Builder.timestamp: Double
  get() = timestamp
  set(value) { timestamp = value }
val TaskLogs.fileDescriptor: FileDescriptor get() = file_descriptor
fun TaskLogs.Builder.setFileDescriptor(value: FileDescriptor): TaskLogs.Builder = apply { file_descriptor = value }
var TaskLogs.Builder.fileDescriptor: FileDescriptor
  get() = file_descriptor
  set(value) { file_descriptor = value }
val TaskLogs.taskProgress: TaskProgress? get() = task_progress
fun TaskLogs.hasTaskProgress(): Boolean = task_progress != null
fun TaskLogs.Builder.setTaskProgress(value: TaskProgress?): TaskLogs.Builder = apply { task_progress = value }
var TaskLogs.Builder.taskProgress: TaskProgress?
  get() = task_progress
  set(value) { task_progress = value }
val TaskLogs.functionCallId: String get() = function_call_id
fun TaskLogs.Builder.setFunctionCallId(value: String): TaskLogs.Builder = apply { function_call_id = value }
var TaskLogs.Builder.functionCallId: String
  get() = function_call_id
  set(value) { function_call_id = value }
val TaskLogs.inputId: String get() = input_id
fun TaskLogs.Builder.setInputId(value: String): TaskLogs.Builder = apply { input_id = value }
var TaskLogs.Builder.inputId: String
  get() = input_id
  set(value) { input_id = value }
val TaskLogs.timestampNs: Long get() = timestamp_ns
fun TaskLogs.Builder.setTimestampNs(value: Long): TaskLogs.Builder = apply { timestamp_ns = value }
var TaskLogs.Builder.timestampNs: Long
  get() = timestamp_ns
  set(value) { timestamp_ns = value }

fun TaskLogsBatch.Companion.newBuilder(): TaskLogsBatch.Builder = TaskLogsBatch.Builder()
fun TaskLogsBatch.Companion.getDefaultInstance(): TaskLogsBatch = TaskLogsBatch()
fun TaskLogsBatch.toByteArray(): ByteArray = encode()
val TaskLogsBatch.taskId: String get() = task_id
fun TaskLogsBatch.Builder.setTaskId(value: String): TaskLogsBatch.Builder = apply { task_id = value }
var TaskLogsBatch.Builder.taskId: String
  get() = task_id
  set(value) { task_id = value }
val TaskLogsBatch.entryId: String get() = entry_id
fun TaskLogsBatch.Builder.setEntryId(value: String): TaskLogsBatch.Builder = apply { entry_id = value }
var TaskLogsBatch.Builder.entryId: String
  get() = entry_id
  set(value) { entry_id = value }
val TaskLogsBatch.appDone: Boolean get() = app_done
fun TaskLogsBatch.Builder.setAppDone(value: Boolean): TaskLogsBatch.Builder = apply { app_done = value }
var TaskLogsBatch.Builder.appDone: Boolean
  get() = app_done
  set(value) { app_done = value }
val TaskLogsBatch.functionId: String get() = function_id
fun TaskLogsBatch.Builder.setFunctionId(value: String): TaskLogsBatch.Builder = apply { function_id = value }
var TaskLogsBatch.Builder.functionId: String
  get() = function_id
  set(value) { function_id = value }
val TaskLogsBatch.inputId: String get() = input_id
fun TaskLogsBatch.Builder.setInputId(value: String): TaskLogsBatch.Builder = apply { input_id = value }
var TaskLogsBatch.Builder.inputId: String
  get() = input_id
  set(value) { input_id = value }
val TaskLogsBatch.imageId: String get() = image_id
fun TaskLogsBatch.Builder.setImageId(value: String): TaskLogsBatch.Builder = apply { image_id = value }
var TaskLogsBatch.Builder.imageId: String
  get() = image_id
  set(value) { image_id = value }
fun TaskLogsBatch.Builder.setEof(value: Boolean): TaskLogsBatch.Builder = apply { eof = value }
var TaskLogsBatch.Builder.eof: Boolean
  get() = eof
  set(value) { eof = value }
val TaskLogsBatch.ptyExecId: String get() = pty_exec_id
fun TaskLogsBatch.Builder.setPtyExecId(value: String): TaskLogsBatch.Builder = apply { pty_exec_id = value }
var TaskLogsBatch.Builder.ptyExecId: String
  get() = pty_exec_id
  set(value) { pty_exec_id = value }
val TaskLogsBatch.rootFunctionId: String get() = root_function_id
fun TaskLogsBatch.Builder.setRootFunctionId(value: String): TaskLogsBatch.Builder = apply { root_function_id = value }
var TaskLogsBatch.Builder.rootFunctionId: String
  get() = root_function_id
  set(value) { root_function_id = value }
val TaskLogsBatch.ttlDays: Int get() = ttl_days
fun TaskLogsBatch.Builder.setTtlDays(value: Int): TaskLogsBatch.Builder = apply { ttl_days = value }
var TaskLogsBatch.Builder.ttlDays: Int
  get() = ttl_days
  set(value) { ttl_days = value }
val TaskLogsBatch.itemsList: List<TaskLogs> get() = items
val TaskLogsBatch.itemsCount: Int get() = items.size
fun TaskLogsBatch.Builder.addItems(value: TaskLogs): TaskLogsBatch.Builder = apply { items = items + value }
fun TaskLogsBatch.Builder.addAllItems(items: Iterable<TaskLogs>): TaskLogsBatch.Builder = apply { this.items = this.items + items }
fun TaskLogsBatch.Builder.clearItems(): TaskLogsBatch.Builder = apply { items = emptyList() }

fun TaskProgress.Companion.newBuilder(): TaskProgress.Builder = TaskProgress.Builder()
fun TaskProgress.Companion.getDefaultInstance(): TaskProgress = TaskProgress()
fun TaskProgress.toByteArray(): ByteArray = encode()
fun TaskProgress.Builder.setLen(value: Long): TaskProgress.Builder = apply { len = value }
var TaskProgress.Builder.len: Long
  get() = len
  set(value) { len = value }
fun TaskProgress.Builder.setPos(value: Long): TaskProgress.Builder = apply { pos = value }
var TaskProgress.Builder.pos: Long
  get() = pos
  set(value) { pos = value }
val TaskProgress.progressType: ProgressType get() = progress_type
fun TaskProgress.Builder.setProgressType(value: ProgressType): TaskProgress.Builder = apply { progress_type = value }
var TaskProgress.Builder.progressType: ProgressType
  get() = progress_type
  set(value) { progress_type = value }
fun TaskProgress.Builder.setDescription(value: String): TaskProgress.Builder = apply { description = value }
var TaskProgress.Builder.description: String
  get() = description
  set(value) { description = value }

fun TaskTemplate.Companion.newBuilder(): TaskTemplate.Builder = TaskTemplate.Builder()
fun TaskTemplate.Companion.getDefaultInstance(): TaskTemplate = TaskTemplate()
fun TaskTemplate.toByteArray(): ByteArray = encode()
fun TaskTemplate.Builder.setRank(value: Int): TaskTemplate.Builder = apply { rank = value }
var TaskTemplate.Builder.rank: Int
  get() = rank
  set(value) { rank = value }
fun TaskTemplate.hasResources(): Boolean = resources != null
fun TaskTemplate.Builder.setResources(value: Resources?): TaskTemplate.Builder = apply { resources = value }
var TaskTemplate.Builder.resources: Resources?
  get() = resources
  set(value) { resources = value }
val TaskTemplate.targetConcurrentInputs: Int get() = target_concurrent_inputs
fun TaskTemplate.Builder.setTargetConcurrentInputs(value: Int): TaskTemplate.Builder = apply { target_concurrent_inputs = value }
var TaskTemplate.Builder.targetConcurrentInputs: Int
  get() = target_concurrent_inputs
  set(value) { target_concurrent_inputs = value }
val TaskTemplate.maxConcurrentInputs: Int get() = max_concurrent_inputs
fun TaskTemplate.Builder.setMaxConcurrentInputs(value: Int): TaskTemplate.Builder = apply { max_concurrent_inputs = value }
var TaskTemplate.Builder.maxConcurrentInputs: Int
  get() = max_concurrent_inputs
  set(value) { max_concurrent_inputs = value }
fun TaskTemplate.Builder.setIndex(value: Int): TaskTemplate.Builder = apply { index = value }
var TaskTemplate.Builder.index: Int
  get() = index
  set(value) { index = value }

fun TunnelData.Companion.newBuilder(): TunnelData.Builder = TunnelData.Builder()
fun TunnelData.Companion.getDefaultInstance(): TunnelData = TunnelData()
fun TunnelData.toByteArray(): ByteArray = encode()
fun TunnelData.Builder.setHost(value: String): TunnelData.Builder = apply { host = value }
var TunnelData.Builder.host: String
  get() = host
  set(value) { host = value }
fun TunnelData.Builder.setPort(value: Int): TunnelData.Builder = apply { port = value }
var TunnelData.Builder.port: Int
  get() = port
  set(value) { port = value }
val TunnelData.unencryptedHost: String? get() = unencrypted_host
fun TunnelData.hasUnencryptedHost(): Boolean = unencrypted_host != null
fun TunnelData.Builder.setUnencryptedHost(value: String?): TunnelData.Builder = apply { unencrypted_host = value }
var TunnelData.Builder.unencryptedHost: String?
  get() = unencrypted_host
  set(value) { unencrypted_host = value }
val TunnelData.unencryptedPort: Int? get() = unencrypted_port
fun TunnelData.hasUnencryptedPort(): Boolean = unencrypted_port != null
fun TunnelData.Builder.setUnencryptedPort(value: Int?): TunnelData.Builder = apply { unencrypted_port = value }
var TunnelData.Builder.unencryptedPort: Int?
  get() = unencrypted_port
  set(value) { unencrypted_port = value }
val TunnelData.containerPort: Int get() = container_port
fun TunnelData.Builder.setContainerPort(value: Int): TunnelData.Builder = apply { container_port = value }
var TunnelData.Builder.containerPort: Int
  get() = container_port
  set(value) { container_port = value }

fun UploadUrlList.Companion.newBuilder(): UploadUrlList.Builder = UploadUrlList.Builder()
fun UploadUrlList.Companion.getDefaultInstance(): UploadUrlList = UploadUrlList()
fun UploadUrlList.toByteArray(): ByteArray = encode()
val UploadUrlList.itemsList: List<String> get() = items
val UploadUrlList.itemsCount: Int get() = items.size
fun UploadUrlList.Builder.addItems(value: String): UploadUrlList.Builder = apply { items = items + value }
fun UploadUrlList.Builder.addAllItems(items: Iterable<String>): UploadUrlList.Builder = apply { this.items = this.items + items }
fun UploadUrlList.Builder.clearItems(): UploadUrlList.Builder = apply { items = emptyList() }

fun VolumeDeleteRequest.Companion.newBuilder(): VolumeDeleteRequest.Builder = VolumeDeleteRequest.Builder()
fun VolumeDeleteRequest.Companion.getDefaultInstance(): VolumeDeleteRequest = VolumeDeleteRequest()
fun VolumeDeleteRequest.toByteArray(): ByteArray = encode()
val VolumeDeleteRequest.volumeId: String get() = volume_id
fun VolumeDeleteRequest.Builder.setVolumeId(value: String): VolumeDeleteRequest.Builder = apply { volume_id = value }
var VolumeDeleteRequest.Builder.volumeId: String
  get() = volume_id
  set(value) { volume_id = value }
val VolumeDeleteRequest.environmentName: String get() = environment_name
fun VolumeDeleteRequest.Builder.setEnvironmentName(value: String): VolumeDeleteRequest.Builder = apply { environment_name = value }
var VolumeDeleteRequest.Builder.environmentName: String
  get() = environment_name
  set(value) { environment_name = value }

fun VolumeGetOrCreateRequest.Companion.newBuilder(): VolumeGetOrCreateRequest.Builder = VolumeGetOrCreateRequest.Builder()
fun VolumeGetOrCreateRequest.Companion.getDefaultInstance(): VolumeGetOrCreateRequest = VolumeGetOrCreateRequest()
fun VolumeGetOrCreateRequest.toByteArray(): ByteArray = encode()
val VolumeGetOrCreateRequest.deploymentName: String get() = deployment_name
fun VolumeGetOrCreateRequest.Builder.setDeploymentName(value: String): VolumeGetOrCreateRequest.Builder = apply { deployment_name = value }
var VolumeGetOrCreateRequest.Builder.deploymentName: String
  get() = deployment_name
  set(value) { deployment_name = value }
val VolumeGetOrCreateRequest.environmentName: String get() = environment_name
fun VolumeGetOrCreateRequest.Builder.setEnvironmentName(value: String): VolumeGetOrCreateRequest.Builder = apply { environment_name = value }
var VolumeGetOrCreateRequest.Builder.environmentName: String
  get() = environment_name
  set(value) { environment_name = value }
val VolumeGetOrCreateRequest.objectCreationType: ObjectCreationType get() = object_creation_type
fun VolumeGetOrCreateRequest.Builder.setObjectCreationType(value: ObjectCreationType): VolumeGetOrCreateRequest.Builder = apply { object_creation_type = value }
var VolumeGetOrCreateRequest.Builder.objectCreationType: ObjectCreationType
  get() = object_creation_type
  set(value) { object_creation_type = value }
val VolumeGetOrCreateRequest.appId: String get() = app_id
fun VolumeGetOrCreateRequest.Builder.setAppId(value: String): VolumeGetOrCreateRequest.Builder = apply { app_id = value }
var VolumeGetOrCreateRequest.Builder.appId: String
  get() = app_id
  set(value) { app_id = value }
fun VolumeGetOrCreateRequest.Builder.setVersion(value: VolumeFsVersion): VolumeGetOrCreateRequest.Builder = apply { version = value }
var VolumeGetOrCreateRequest.Builder.version: VolumeFsVersion
  get() = version
  set(value) { version = value }

fun VolumeGetOrCreateResponse.Companion.newBuilder(): VolumeGetOrCreateResponse.Builder = VolumeGetOrCreateResponse.Builder()
fun VolumeGetOrCreateResponse.Companion.getDefaultInstance(): VolumeGetOrCreateResponse = VolumeGetOrCreateResponse()
fun VolumeGetOrCreateResponse.toByteArray(): ByteArray = encode()
val VolumeGetOrCreateResponse.volumeId: String get() = volume_id
fun VolumeGetOrCreateResponse.Builder.setVolumeId(value: String): VolumeGetOrCreateResponse.Builder = apply { volume_id = value }
var VolumeGetOrCreateResponse.Builder.volumeId: String
  get() = volume_id
  set(value) { volume_id = value }
fun VolumeGetOrCreateResponse.Builder.setVersion(value: VolumeFsVersion): VolumeGetOrCreateResponse.Builder = apply { version = value }
var VolumeGetOrCreateResponse.Builder.version: VolumeFsVersion
  get() = version
  set(value) { version = value }
fun VolumeGetOrCreateResponse.hasMetadata(): Boolean = metadata != null
fun VolumeGetOrCreateResponse.Builder.setMetadata(value: VolumeMetadata?): VolumeGetOrCreateResponse.Builder = apply { metadata = value }
var VolumeGetOrCreateResponse.Builder.metadata: VolumeMetadata?
  get() = metadata
  set(value) { metadata = value }

fun VolumeHeartbeatRequest.Companion.newBuilder(): VolumeHeartbeatRequest.Builder = VolumeHeartbeatRequest.Builder()
fun VolumeHeartbeatRequest.Companion.getDefaultInstance(): VolumeHeartbeatRequest = VolumeHeartbeatRequest()
fun VolumeHeartbeatRequest.toByteArray(): ByteArray = encode()
val VolumeHeartbeatRequest.volumeId: String get() = volume_id
fun VolumeHeartbeatRequest.Builder.setVolumeId(value: String): VolumeHeartbeatRequest.Builder = apply { volume_id = value }
var VolumeHeartbeatRequest.Builder.volumeId: String
  get() = volume_id
  set(value) { volume_id = value }

fun VolumeMetadata.Companion.newBuilder(): VolumeMetadata.Builder = VolumeMetadata.Builder()
fun VolumeMetadata.Companion.getDefaultInstance(): VolumeMetadata = VolumeMetadata()
fun VolumeMetadata.toByteArray(): ByteArray = encode()
fun VolumeMetadata.Builder.setVersion(value: VolumeFsVersion): VolumeMetadata.Builder = apply { version = value }
var VolumeMetadata.Builder.version: VolumeFsVersion
  get() = version
  set(value) { version = value }
fun VolumeMetadata.Builder.setName(value: String): VolumeMetadata.Builder = apply { name = value }
var VolumeMetadata.Builder.name: String
  get() = name
  set(value) { name = value }
val VolumeMetadata.creationInfo: CreationInfo? get() = creation_info
fun VolumeMetadata.hasCreationInfo(): Boolean = creation_info != null
fun VolumeMetadata.Builder.setCreationInfo(value: CreationInfo?): VolumeMetadata.Builder = apply { creation_info = value }
var VolumeMetadata.Builder.creationInfo: CreationInfo?
  get() = creation_info
  set(value) { creation_info = value }

fun VolumeMount.Companion.newBuilder(): VolumeMount.Builder = VolumeMount.Builder()
fun VolumeMount.Companion.getDefaultInstance(): VolumeMount = VolumeMount()
fun VolumeMount.toByteArray(): ByteArray = encode()
val VolumeMount.volumeId: String get() = volume_id
fun VolumeMount.Builder.setVolumeId(value: String): VolumeMount.Builder = apply { volume_id = value }
var VolumeMount.Builder.volumeId: String
  get() = volume_id
  set(value) { volume_id = value }
val VolumeMount.mountPath: String get() = mount_path
fun VolumeMount.Builder.setMountPath(value: String): VolumeMount.Builder = apply { mount_path = value }
var VolumeMount.Builder.mountPath: String
  get() = mount_path
  set(value) { mount_path = value }
val VolumeMount.allowBackgroundCommits: Boolean get() = allow_background_commits
fun VolumeMount.Builder.setAllowBackgroundCommits(value: Boolean): VolumeMount.Builder = apply { allow_background_commits = value }
var VolumeMount.Builder.allowBackgroundCommits: Boolean
  get() = allow_background_commits
  set(value) { allow_background_commits = value }
val VolumeMount.readOnly: Boolean get() = read_only
fun VolumeMount.Builder.setReadOnly(value: Boolean): VolumeMount.Builder = apply { read_only = value }
var VolumeMount.Builder.readOnly: Boolean
  get() = read_only
  set(value) { read_only = value }

fun Warning.Companion.newBuilder(): Warning.Builder = Warning.Builder()
fun Warning.Companion.getDefaultInstance(): Warning = Warning()
fun Warning.toByteArray(): ByteArray = encode()
fun Warning.Builder.setType(value: Warning.WarningType): Warning.Builder = apply { type = value }
var Warning.Builder.type: Warning.WarningType
  get() = type
  set(value) { type = value }
fun Warning.Builder.setMessage(value: String): Warning.Builder = apply { message = value }
var Warning.Builder.message: String
  get() = message
  set(value) { message = value }

fun WebUrlInfo.Companion.newBuilder(): WebUrlInfo.Builder = WebUrlInfo.Builder()
fun WebUrlInfo.Companion.getDefaultInstance(): WebUrlInfo = WebUrlInfo()
fun WebUrlInfo.toByteArray(): ByteArray = encode()
fun WebUrlInfo.Builder.setTruncated(value: Boolean): WebUrlInfo.Builder = apply { truncated = value }
var WebUrlInfo.Builder.truncated: Boolean
  get() = truncated
  set(value) { truncated = value }
val WebUrlInfo.hasUniqueHash: Boolean get() = has_unique_hash
fun WebUrlInfo.Builder.setHasUniqueHash(value: Boolean): WebUrlInfo.Builder = apply { has_unique_hash = value }
var WebUrlInfo.Builder.hasUniqueHash: Boolean
  get() = has_unique_hash
  set(value) { has_unique_hash = value }
val WebUrlInfo.labelStolen: Boolean get() = label_stolen
fun WebUrlInfo.Builder.setLabelStolen(value: Boolean): WebUrlInfo.Builder = apply { label_stolen = value }
var WebUrlInfo.Builder.labelStolen: Boolean
  get() = label_stolen
  set(value) { label_stolen = value }

fun WebhookConfig.Companion.newBuilder(): WebhookConfig.Builder = WebhookConfig.Builder()
fun WebhookConfig.Companion.getDefaultInstance(): WebhookConfig = WebhookConfig()
fun WebhookConfig.toByteArray(): ByteArray = encode()
fun WebhookConfig.Builder.setType(value: WebhookType): WebhookConfig.Builder = apply { type = value }
var WebhookConfig.Builder.type: WebhookType
  get() = type
  set(value) { type = value }
fun WebhookConfig.Builder.setMethod(value: String): WebhookConfig.Builder = apply { method = value }
var WebhookConfig.Builder.method: String
  get() = method
  set(value) { method = value }
val WebhookConfig.requestedSuffix: String get() = requested_suffix
fun WebhookConfig.Builder.setRequestedSuffix(value: String): WebhookConfig.Builder = apply { requested_suffix = value }
var WebhookConfig.Builder.requestedSuffix: String
  get() = requested_suffix
  set(value) { requested_suffix = value }
val WebhookConfig.asyncMode: WebhookAsyncMode get() = async_mode
fun WebhookConfig.Builder.setAsyncMode(value: WebhookAsyncMode): WebhookConfig.Builder = apply { async_mode = value }
var WebhookConfig.Builder.asyncMode: WebhookAsyncMode
  get() = async_mode
  set(value) { async_mode = value }
val WebhookConfig.webServerPort: Int get() = web_server_port
fun WebhookConfig.Builder.setWebServerPort(value: Int): WebhookConfig.Builder = apply { web_server_port = value }
var WebhookConfig.Builder.webServerPort: Int
  get() = web_server_port
  set(value) { web_server_port = value }
val WebhookConfig.webServerStartupTimeout: Float get() = web_server_startup_timeout
fun WebhookConfig.Builder.setWebServerStartupTimeout(value: Float): WebhookConfig.Builder = apply { web_server_startup_timeout = value }
var WebhookConfig.Builder.webServerStartupTimeout: Float
  get() = web_server_startup_timeout
  set(value) { web_server_startup_timeout = value }
val WebhookConfig.webEndpointDocs: Boolean get() = web_endpoint_docs
fun WebhookConfig.Builder.setWebEndpointDocs(value: Boolean): WebhookConfig.Builder = apply { web_endpoint_docs = value }
var WebhookConfig.Builder.webEndpointDocs: Boolean
  get() = web_endpoint_docs
  set(value) { web_endpoint_docs = value }
val WebhookConfig.requiresProxyAuth: Boolean get() = requires_proxy_auth
fun WebhookConfig.Builder.setRequiresProxyAuth(value: Boolean): WebhookConfig.Builder = apply { requires_proxy_auth = value }
var WebhookConfig.Builder.requiresProxyAuth: Boolean
  get() = requires_proxy_auth
  set(value) { requires_proxy_auth = value }
val WebhookConfig.ephemeralSuffix: String get() = ephemeral_suffix
fun WebhookConfig.Builder.setEphemeralSuffix(value: String): WebhookConfig.Builder = apply { ephemeral_suffix = value }
var WebhookConfig.Builder.ephemeralSuffix: String
  get() = ephemeral_suffix
  set(value) { ephemeral_suffix = value }
val WebhookConfig.customDomains: List<CustomDomainConfig> get() = custom_domains
val WebhookConfig.customDomainsList: List<CustomDomainConfig> get() = custom_domains
val WebhookConfig.customDomainsCount: Int get() = custom_domains.size
fun WebhookConfig.Builder.addCustomDomains(value: CustomDomainConfig): WebhookConfig.Builder = apply { custom_domains = custom_domains + value }
fun WebhookConfig.Builder.addAllCustomDomains(items: Iterable<CustomDomainConfig>): WebhookConfig.Builder = apply { this.custom_domains = this.custom_domains + items }
fun WebhookConfig.Builder.clearCustomDomains(): WebhookConfig.Builder = apply { custom_domains = emptyList() }

val ClassParameterInfo.ParameterSerializationFormat.number: Int get() = value

val CloudBucketMount.BucketType.number: Int get() = value

val CloudBucketMount.MetadataTTLType.number: Int get() = value

val Function.DefinitionType.number: Int get() = value

val Function.FunctionType.number: Int get() = value

val FunctionSchema.FunctionSchemaType.number: Int get() = value

val GenericResult.GenericStatus.number: Int get() = value

val NetworkAccess.NetworkAccessType.number: Int get() = value

val PTYInfo.PTYType.number: Int get() = value

val Warning.WarningType.number: Int get() = value


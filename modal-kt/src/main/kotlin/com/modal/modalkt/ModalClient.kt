package com.modal.modalkt

data class ModalClientParams(
    val tokenId: String? = null,
    val tokenSecret: String? = null,
    val environment: String? = null,
    val timeoutMs: Long? = null,
    val logger: Logger? = null,
    val logLevel: String? = null,
    val timeout: Long? = null,
    internal val authTokenProvider: AuthTokenProvider? = null,
    val controlPlaneClient: ControlPlaneClient? = null,
    val grpcInterceptors: List<io.grpc.ClientInterceptor> = emptyList(),
    val backgroundScope: kotlinx.coroutines.CoroutineScope? = null,
    val ephemeralHeartbeatSleepMs: Long = ephemeralObjectHeartbeatSleep,
    internal val taskCommandRouterFactory: (suspend (client: ModalClient, taskId: String) -> TaskCommandRouter)? = null,
    internal val inputPlaneClientFactory: ((serverUrl: String) -> ControlPlaneClient)? = null,
)

class ModalClient(
    params: ModalClientParams = ModalClientParams(),
) {
    val profile: Profile
    val logger: Logger
    internal val cpClient: ControlPlaneClient
    val apps: AppService
    val cloudBucketMounts: CloudBucketMountService
    val cls: ClsService
    val functionCalls: FunctionCallService
    val functions: FunctionService
    val queues: QueueService
    val sandboxes: SandboxService
    val secrets: SecretService
    val volumes: VolumeService
    val proxies: ProxyService
    val images: ImageService
    internal val backgroundScope: kotlinx.coroutines.CoroutineScope
    internal val ephemeralHeartbeatSleepMs: Long
    internal val taskCommandRouterFactory: suspend (client: ModalClient, taskId: String) -> TaskCommandRouter
    private val inputPlaneClientFactory: (String) -> ControlPlaneClient
    private val inputPlaneClients = mutableMapOf<String, ControlPlaneClient>()
    private val defaultTimeoutMs: Long?

    private val authTokenManager: AuthTokenManager?

    init {
        checkForRenamedParams(
            mapOf("timeout" to params.timeout).filterValues { it != null },
            mapOf("timeout" to "timeoutMs"),
        )

        val baseProfile = getProfile(System.getenv("MODAL_PROFILE"))
        profile = baseProfile.copy(
            tokenId = params.tokenId ?: baseProfile.tokenId,
            tokenSecret = params.tokenSecret ?: baseProfile.tokenSecret,
            environment = params.environment ?: baseProfile.environment,
        )

        logger = createLogger(params.logger, params.logLevel ?: profile.logLevel.orEmpty())
        logger.debug(
            "Initializing Modal client",
            "version",
            getSdkVersion(),
            "server_url",
            profile.serverUrl,
        )

        cpClient = params.controlPlaneClient ?: GrpcControlPlaneClient(profile, logger, params.grpcInterceptors, params.timeoutMs)
        defaultTimeoutMs = params.timeoutMs
        backgroundScope = params.backgroundScope
            ?: kotlinx.coroutines.CoroutineScope(
                kotlinx.coroutines.SupervisorJob() + kotlinx.coroutines.Dispatchers.IO,
            )
        ephemeralHeartbeatSleepMs = params.ephemeralHeartbeatSleepMs
        taskCommandRouterFactory = params.taskCommandRouterFactory ?: { modalClient, taskId ->
            TaskCommandRouterClientImpl.tryInit(
                serverClient = modalClient.cpClient,
                taskId = taskId,
                logger = modalClient.logger,
                profile = modalClient.profile,
            ) ?: throw InvalidError("Command router access is not available for this sandbox")
        }
        inputPlaneClientFactory = params.inputPlaneClientFactory ?: { serverUrl ->
            GrpcControlPlaneClient(profile.copy(serverUrl = serverUrl), logger, params.grpcInterceptors, defaultTimeoutMs)
        }
        cloudBucketMounts = CloudBucketMountsServiceHolder.create(this)
        apps = AppService(this)
        cls = ClsService(this)
        functionCalls = FunctionCallService(this)
        functions = FunctionService(this)
        queues = QueueService(this)
        sandboxes = SandboxService(this)
        secrets = SecretService(this)
        volumes = VolumeService(this)
        proxies = ProxyService(this)
        images = ImageService(this)

        authTokenManager = AuthTokenManager(params.authTokenProvider ?: cpClient, logger)

        logger.debug("Modal client initialized successfully")
    }

    fun version(): String {
        return getSdkVersion()
    }

    fun environmentName(environment: String? = null): String {
        return environment ?: profile.environment.orEmpty()
    }

    fun imageBuilderVersion(version: String? = null): String {
        return version ?: profile.imageBuilderVersion ?: "2024.10"
    }

    fun close() {
        logger.debug("Closing Modal client")
        cpClient.close()
        inputPlaneClients.values.forEach { it.close() }
        inputPlaneClients.clear()
        logger.debug("Modal client closed")
    }

    suspend fun getAuthToken(): String? {
        return authTokenManager?.getToken()
    }

    internal fun ipClient(serverUrl: String): ControlPlaneClient {
        return inputPlaneClients.getOrPut(serverUrl) {
            inputPlaneClientFactory(serverUrl)
        }
    }
}

private object CloudBucketMountsServiceHolder {
    fun create(client: ModalClient): CloudBucketMountService {
        return CloudBucketMountService(client)
    }
}

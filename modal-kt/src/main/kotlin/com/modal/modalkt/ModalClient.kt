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
    val backgroundScope: kotlinx.coroutines.CoroutineScope? = null,
    val ephemeralHeartbeatSleepMs: Long = ephemeralObjectHeartbeatSleep,
)

class ModalClient(
    params: ModalClientParams = ModalClientParams(),
) {
    val profile: Profile
    val logger: Logger
    internal val cpClient: ControlPlaneClient
    val apps: AppService
    val cloudBucketMounts: CloudBucketMountService
    val secrets: SecretService
    val volumes: VolumeService
    val proxies: ProxyService
    val images: ImageService
    internal val backgroundScope: kotlinx.coroutines.CoroutineScope
    internal val ephemeralHeartbeatSleepMs: Long

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

        cpClient = params.controlPlaneClient ?: GrpcControlPlaneClient(profile, logger)
        backgroundScope = params.backgroundScope
            ?: kotlinx.coroutines.CoroutineScope(
                kotlinx.coroutines.SupervisorJob() + kotlinx.coroutines.Dispatchers.IO,
            )
        ephemeralHeartbeatSleepMs = params.ephemeralHeartbeatSleepMs
        cloudBucketMounts = CloudBucketMountsServiceHolder.create(this)
        apps = AppService(this)
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
        logger.debug("Modal client closed")
    }

    suspend fun getAuthToken(): String? {
        return authTokenManager?.getToken()
    }
}

private object CloudBucketMountsServiceHolder {
    fun create(client: ModalClient): CloudBucketMountService {
        return CloudBucketMountService(client)
    }
}

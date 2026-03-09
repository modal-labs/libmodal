package com.modal.modalkt

import org.tomlj.Toml
import org.tomlj.TomlTable
import java.net.URI
import java.nio.file.Files
import java.nio.file.Path
import kotlin.io.path.Path
import kotlin.io.path.notExists

data class Profile(
    val serverUrl: String,
    val tokenId: String? = null,
    val tokenSecret: String? = null,
    val environment: String? = null,
    val imageBuilderVersion: String? = null,
    val logLevel: String? = null,
)

internal data class RawProfile(
    val serverUrl: String? = null,
    val tokenId: String? = null,
    val tokenSecret: String? = null,
    val environment: String? = null,
    val imageBuilderVersion: String? = null,
    val logLevel: String? = null,
    val active: Boolean = false,
)

fun isLocalhost(profile: Profile): Boolean {
    val hostname = runCatching { URI(profile.serverUrl).host }.getOrNull() ?: return false
    return hostname == "localhost" ||
        hostname == "127.0.0.1" ||
        hostname == "::1" ||
        hostname == "172.21.0.1"
}

fun configFilePath(): String {
    val customPath = System.getenv("MODAL_CONFIG_PATH")
    if (!customPath.isNullOrEmpty()) {
        return customPath
    }

    val home = System.getProperty("user.home")
    return Path(home, ".modal.toml").toString()
}

private fun readConfigFile(): Map<String, RawProfile> {
    return try {
        val path = Path(configFilePath())
        if (path.notExists()) {
            emptyMap()
        } else {
            parseConfig(path)
        }
    } catch (_: Exception) {
        emptyMap()
    }
}

private fun parseConfig(path: Path): Map<String, RawProfile> {
    val content = Files.readString(path)
    val parsed = Toml.parse(content)
    if (parsed.hasErrors()) {
        return emptyMap()
    }

    val profiles = linkedMapOf<String, RawProfile>()
    for (name in parsed.keySet()) {
        val table = parsed.getTable(name) ?: continue
        profiles[name] = table.toRawProfile()
    }
    return profiles
}

private fun TomlTable.toRawProfile(): RawProfile {
    return RawProfile(
        serverUrl = getString("server_url"),
        tokenId = getString("token_id"),
        tokenSecret = getString("token_secret"),
        environment = getString("environment"),
        imageBuilderVersion = getString("image_builder_version"),
        logLevel = getString("loglevel"),
        active = getBoolean("active") ?: false,
    )
}

private val config: Map<String, RawProfile> by lazy {
    readConfigFile()
}

fun getProfile(profileName: String? = null): Profile {
    var resolvedProfileName = profileName
    if (resolvedProfileName.isNullOrEmpty()) {
        resolvedProfileName = config.entries.firstOrNull { it.value.active }?.key
    }

    val profile = if (resolvedProfileName != null) {
        config[resolvedProfileName] ?: RawProfile()
    } else {
        RawProfile()
    }

    return Profile(
        serverUrl = firstNonEmpty(
            System.getenv("MODAL_SERVER_URL"),
            profile.serverUrl,
            "https://api.modal.com:443",
        )!!,
        tokenId = firstNonEmpty(System.getenv("MODAL_TOKEN_ID"), profile.tokenId),
        tokenSecret = firstNonEmpty(System.getenv("MODAL_TOKEN_SECRET"), profile.tokenSecret),
        environment = firstNonEmpty(System.getenv("MODAL_ENVIRONMENT"), profile.environment),
        imageBuilderVersion = firstNonEmpty(
            System.getenv("MODAL_IMAGE_BUILDER_VERSION"),
            profile.imageBuilderVersion,
        ),
        logLevel = firstNonEmpty(System.getenv("MODAL_LOGLEVEL"), profile.logLevel),
    )
}

internal fun firstNonEmpty(vararg values: String?): String? {
    for (value in values) {
        if (!value.isNullOrEmpty()) {
            return value
        }
    }
    return null
}

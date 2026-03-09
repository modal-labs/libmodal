import com.google.protobuf.gradle.id
import org.jetbrains.kotlin.gradle.dsl.JvmTarget

plugins {
    `java-library`
    kotlin("jvm") version "2.2.21"
    id("com.google.protobuf") version "0.9.5"
}

group = "com.modal"
version = "0.1.0"
description = "Modal Kotlin SDK"

val coroutinesVersion = "1.10.2"
val grpcVersion = "1.73.0"
val grpcKotlinVersion = "1.4.3"
val protobufVersion = "3.25.5"
val junitVersion = "5.12.0"

java {
    toolchain {
        languageVersion = JavaLanguageVersion.of(21)
    }
    withSourcesJar()
    withJavadocJar()
}

repositories {
    mavenCentral()
}

sourceSets {
    named("main") {
        proto {
            srcDir("../modal-client")
        }
        kotlin.srcDir(layout.buildDirectory.dir("generated/source/proto/main/grpckt"))
    }

    create("integrationTest") {
        kotlin.srcDir("src/integrationTest/kotlin")
        resources.srcDir("src/integrationTest/resources")
        compileClasspath += sourceSets["main"].output + configurations["testRuntimeClasspath"]
        runtimeClasspath += output + compileClasspath
    }

    create("examples") {
        kotlin.srcDir("src/examples/kotlin")
        resources.srcDir("src/examples/resources")
        compileClasspath += sourceSets["main"].output + configurations["runtimeClasspath"]
        runtimeClasspath += output + compileClasspath
    }
}

configurations.named("integrationTestImplementation") {
    extendsFrom(configurations["testImplementation"])
}
configurations.named("integrationTestRuntimeOnly") {
    extendsFrom(configurations["testRuntimeOnly"])
}
configurations.named("examplesImplementation") {
    extendsFrom(configurations["implementation"])
}
configurations.named("examplesRuntimeOnly") {
    extendsFrom(configurations["runtimeOnly"])
}

dependencies {
    api(kotlin("stdlib"))
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-core:$coroutinesVersion")
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-jdk8:$coroutinesVersion")
    implementation("io.grpc:grpc-kotlin-stub:$grpcKotlinVersion")
    implementation("io.grpc:grpc-protobuf:$grpcVersion")
    implementation("io.grpc:grpc-stub:$grpcVersion")
    implementation("com.google.protobuf:protobuf-java:$protobufVersion")
    implementation("com.google.protobuf:protobuf-kotlin:$protobufVersion")
    implementation("org.tomlj:tomlj:1.1.1")
    implementation("com.upokecenter:cbor:4.5.6")
    implementation("net.razorvine:pickle:1.5")
    implementation("io.grpc:grpc-netty-shaded:$grpcVersion")
    compileOnly("org.apache.tomcat:annotations-api:6.0.53")

    testImplementation(kotlin("test"))
    testImplementation("org.jetbrains.kotlinx:kotlinx-coroutines-test:$coroutinesVersion")
    testImplementation("org.junit.jupiter:junit-jupiter:$junitVersion")
    testImplementation("com.github.stefanbirkner:system-lambda:1.2.1")
    testImplementation("io.grpc:grpc-inprocess:$grpcVersion")
    testRuntimeOnly("org.junit.platform:junit-platform-launcher")
}

protobuf {
    protoc {
        artifact = "com.google.protobuf:protoc:$protobufVersion"
    }
    plugins {
        id("grpc") {
            artifact = "io.grpc:protoc-gen-grpc-java:$grpcVersion"
        }
        id("grpckt") {
            artifact = "io.grpc:protoc-gen-grpc-kotlin:$grpcKotlinVersion:jdk8@jar"
        }
    }
    generateProtoTasks {
        all().configureEach {
            plugins {
                id("grpc")
                id("grpckt")
            }
        }
    }
}

kotlin {
    compilerOptions {
        jvmTarget.set(JvmTarget.JVM_21)
        freeCompilerArgs.addAll(
            "-Xjsr305=strict",
            "-Xannotation-default-target=param-property",
        )
    }
}

tasks.named<Test>("test") {
    useJUnitPlatform()
    jvmArgs(
        "--add-opens=java.base/java.lang=ALL-UNNAMED",
        "--add-opens=java.base/java.util=ALL-UNNAMED",
    )
}

val integrationTest = tasks.register<Test>("integrationTest") {
    description = "Runs cloud-backed integration tests."
    group = LifecycleBasePlugin.VERIFICATION_GROUP
    testClassesDirs = sourceSets["integrationTest"].output.classesDirs
    classpath = sourceSets["integrationTest"].runtimeClasspath
    useJUnitPlatform()
    shouldRunAfter(tasks.named("test"))
    jvmArgs(
        "--add-opens=java.base/java.lang=ALL-UNNAMED",
        "--add-opens=java.base/java.util=ALL-UNNAMED",
    )
}

tasks.register("compileExamples") {
    description = "Compiles Kotlin example programs."
    group = LifecycleBasePlugin.BUILD_GROUP
    dependsOn(tasks.named("examplesClasses"))
}

val patchGeneratedProtoSources = tasks.register("patchGeneratedProtoSources") {
    dependsOn(tasks.named("generateProto"))

    doLast {
        val apiFile = layout.buildDirectory.file("generated/sources/proto/main/java/modal/client/Api.java").get().asFile
        if (!apiFile.exists()) {
            return@doLast
        }

        var text = apiFile.readText()

        text = text.replace(
            Regex(
                """(/\*\*\s+\* <code>bytes message_bytes = 3;</code>\s+\* @return The messageBytes\.\s+\*/\s+com\.google\.protobuf\.ByteString )getMessageBytes(\(\);)""",
                setOf(RegexOption.MULTILINE),
            ),
            "$1getMessageBytesValue$2",
        )
        text = text.replace(
            Regex(
                """(@java\.lang\.Override\s+public com\.google\.protobuf\.ByteString )getMessageBytes(\(\) \{\s+return messageBytes_;\s+\})""",
                setOf(RegexOption.MULTILINE),
            ),
            "$1getMessageBytesValue$2",
        )
        text = text.replace(
            Regex(
                """(@java\.lang\.Override\s+public com\.google\.protobuf\.ByteString )getMessageBytes(\(\) \{\s+return messageBytes_;\s+\})""",
                setOf(RegexOption.MULTILINE),
            ),
            "$1getMessageBytesValue$2",
        )

        text = text.replace(
            "if (!getMessageBytes()\n          .equals(other.getMessageBytes())) return false;",
            "if (!getMessageBytesValue()\n          .equals(other.getMessageBytesValue())) return false;",
        )
        text = text.replace(
            "hash = (53 * hash) + getMessageBytes().hashCode();",
            "hash = (53 * hash) + getMessageBytesValue().hashCode();",
        )
        text = text.replace(
            "if (other.getMessageBytes() != com.google.protobuf.ByteString.EMPTY) {\n          setMessageBytes(other.getMessageBytes());\n        }",
            "if (other.getMessageBytesValue() != com.google.protobuf.ByteString.EMPTY) {\n          setMessageBytesValue(other.getMessageBytesValue());\n        }",
        )
        text = text.replace(
            "public Builder setMessageBytes(com.google.protobuf.ByteString value) {",
            "public Builder setMessageBytesValue(com.google.protobuf.ByteString value) {",
        )
        text = text.replace(
            "messageBytes_ = getDefaultInstance().getMessageBytes();",
            "messageBytes_ = getDefaultInstance().getMessageBytesValue();",
        )

        apiFile.writeText(text)
    }
}

tasks.named("compileJava") {
    dependsOn(patchGeneratedProtoSources)
}

tasks.named("compileKotlin") {
    dependsOn(patchGeneratedProtoSources)
}

tasks.named("check") {
    dependsOn(integrationTest)
    dependsOn("compileExamples")
}

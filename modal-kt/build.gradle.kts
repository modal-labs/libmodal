import org.jetbrains.kotlin.gradle.dsl.JvmTarget

plugins {
    `java-library`
    idea
    kotlin("jvm") version "2.2.21"
    id("com.squareup.wire") version "5.4.0"
}

group = "com.modal"
version = "0.1.0"
description = "Modal Kotlin SDK"

val coroutinesVersion = "1.10.2"
val grpcVersion = "1.73.0"
val protobufVersion = "3.25.5"
val wireVersion = "5.4.0"
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
        kotlin.srcDir(layout.buildDirectory.dir("generated/sources/wire"))
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
    implementation("io.grpc:grpc-stub:$grpcVersion")
    implementation("com.google.protobuf:protobuf-java:$protobufVersion")
    implementation("com.squareup.wire:wire-runtime:$wireVersion")
    implementation("com.squareup.wire:wire-grpc-client:$wireVersion")
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

wire {
    sourcePath {
        srcDir("../modal-client")
    }
    kotlin {
        javaInterop = true
        rpcRole = "client"
        out = layout.buildDirectory.dir("generated/sources/wire").get().asFile.path
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

idea {
    module {
        sourceDirs.add(layout.buildDirectory.dir("generated/sources/wire").get().asFile)
        generatedSourceDirs.add(layout.buildDirectory.dir("generated/sources/wire").get().asFile)
    }
}

tasks.named("check") {
    dependsOn(integrationTest)
    dependsOn("compileExamples")
}

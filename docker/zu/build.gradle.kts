import com.github.jengelman.gradle.plugins.shadow.tasks.ShadowJar

plugins {
    kotlin("jvm") version "1.3.70"
    id("com.github.johnrengelman.shadow") version "5.2.0"
}

repositories {
    jcenter()
    mavenCentral()
}

dependencies {
    implementation(kotlin("stdlib"))
    implementation("org.apache.zookeeper:zookeeper:3.6.3")
}

tasks.withType<ShadowJar>() {
    classifier = null
    manifest {
        attributes["Main-Class"] = "io.pravega.zookeeper.MainKt"
    }
}

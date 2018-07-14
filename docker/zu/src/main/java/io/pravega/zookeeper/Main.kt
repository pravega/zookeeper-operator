package io.pravega.zookeeper

import org.apache.zookeeper.data.Stat
import java.io.File

const val OBSERVER = "observer"
const val PARTICIPANT= "participant"

/**
 * Utility to Register a server with the Zookeeper Ensemble
 */
fun main(args: Array<String>) {
    val message = "Usage: zu <add | get-role> [options...]"
    if (args.isEmpty()) {
        help(message)
    }

    when {
        "add" == args[0] -> runAdd(args)
        "get" == args[0] -> runGet(args)
        "remove" == args[0] -> runRemove(args)
        "get-role" == args[0] -> runGetRole(args)
        else -> help(message)
    }

}

fun runGet(args: Array<String>, suppressOutput: Boolean = false): String? {
    if (args.size < 3) {
        help("Usage: zu get <zk-url> <server-id>")
    }

    val (_, zkUrl, serverId) = args
    try {
        val zk = newZookeeperAdminClient(zkUrl)
        val zkCfg = String(zk.getConfig(false, Stat()))
        val that = zkCfg
                .split("\n")
                .map {it.split("=")}
                .filter { it[0] != "version" }
                .find{ it[0].split(".")[1] == serverId }
                ?.getOrNull(1)
        when {
            that == null -> {
                System.err.println("Server not found in zookeeper config")
                System.exit(1)
            }
            ! suppressOutput -> {
                println(that.toString())
            }
        }
        return that.toString()
    } catch(e : Exception) {
        System.err.println("Error getting server config")
        e.printStackTrace(System.err)
        System.exit(1)
    }
    return null
}

fun runGetRole(args: Array<String>) {
    if (args.size < 3) {
        help("Usage: zu get-role <zk-url> <server-id>")
    }
    val cfgStr = runGet(args, true)
    if(Regex(".*observer.*") matches cfgStr.toString()) {
        println(OBSERVER)
    } else {
        println(PARTICIPANT)
    }
}

fun reconfigure(zkUrl: String, joining: String?, leaving: String?, outputFile: String?) {
    try {
        val zk = newZookeeperAdminClient(zkUrl)
        val cfg = zk.reconfigure(joining, leaving, null, -1, null)
        val cfgStr = String(cfg)
                .split("\n")
                .filter{!(Regex("^version=.+") matches it)}
                .joinToString("\n")
        if (outputFile == null) {
            println(cfgStr)
        } else {
            File(outputFile).bufferedWriter().use {it.write(cfgStr + "\n")}
        }
    } catch (e: Exception) {
        System.err.println("Error performing zookeeper reconfiguration:")
        e.printStackTrace(System.err)
        System.exit(1)
    }
}

fun runAdd(args: Array<String>) {
    val message = "Usage: zu add <zk-url> <server-id> <server-details> [output-file]"
    if (args.size < 4) {
        help(message)
    }
    val (_, zkUrl, newServerId, newServerDetails) = args
    reconfigure(zkUrl, "server.$newServerId=$newServerDetails", null, args.getOrNull(4))
}

fun runRemove(args: Array<String>) {
    val message = "Usage: zu remove <zk-url> <server-id> [output-file]"
    if (args.size < 3) {
        help(message)
    }
    val (_, zkUrl, newServerId) = args
    reconfigure(zkUrl, null, newServerId, args.getOrNull(2))
}

fun help(message: String) {
    System.err.println(message)
    System.exit(1)
}
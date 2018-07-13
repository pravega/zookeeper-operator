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
        "get-role" == args[0] -> runGetRole(args)
        else -> help(message)
    }

}

fun runGetRole(args: Array<String>) {
    if (args.size < 3) {
        help("Usage: zu get-role <zk-url> <server-id>")
    }

    val (_, zkUrl, serverId) = args
    try {
        val zk = newZookeeperAdminClient(zkUrl)
        val zkCfg = String(zk.getConfig(false, Stat()))
        val that = zkCfg
                .split("\n")
                .map {it.split("=", ":", ";")}
                .find {
                    val (_, thisServerId) = it[0].split('.')
                    (serverId == thisServerId)
                }

        if (that == null) {
            System.err.println("Server not found in zookeeper config")
            System.exit(1)
        } else {
            if ((that.size == 5 || that.size == 6) && OBSERVER == that[4]) {
                println(OBSERVER)
            } else {
                println(PARTICIPANT)
            }
        }

    } catch(e : Exception) {
        System.err.println("Error getting server role")
        e.printStackTrace(System.err)
        System.exit(1)
    }
}

fun runAdd(args: Array<String>) {
    val message = "Usage: zu add <zk-url> <server-id> <server-details> [output-file]"
    if (args.size < 4) {
        help(message)
    }

    val (_, zkUrl, newServerId, newServerDetails, outputFile) = args

    try {
        val zk = newZookeeperAdminClient(zkUrl)
        val cfg = zk.reconfigure("server.$newServerId=$newServerDetails", null, null, -1, null)
        val cfgStr = String(cfg).split("\n").dropLast(1)
        when {
            args.size == 4 ->
                cfgStr.forEach { println(it) }
            args.size == 5 ->
                File(outputFile).bufferedWriter().use {file ->
                    cfgStr.forEach { file.write(it + "\n") }
                }
            else -> help(message)
        }

        println(String(cfg))
    } catch(e : Exception) {
        System.err.println("Error Adding Server")
        e.printStackTrace(System.err)
        System.exit(1)
    }
}

fun help(message: String) {
    System.err.println(message)
    System.exit(1)
}
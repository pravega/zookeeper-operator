/*
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package io.pravega.zookeeper

import org.apache.zookeeper.data.Stat
import org.apache.zookeeper.AsyncCallback.VoidCallback
import java.io.File

const val OBSERVER = "observer"
const val PARTICIPANT= "participant"

/**
 * Utility to Register a server with the Zookeeper Ensemble
 */
fun main(args: Array<String>) {
    val message = "Usage: zu <add | get-all | get | remove | get-role | sync> [options...]"
    if (args.isEmpty()) {
        help(message)
    }

    when {
        "add" == args[0] -> runAdd(args)
        "get-all" == args[0] -> runGetAll(args)
        "get" == args[0] -> runGet(args)
        "remove" == args[0] -> runRemove(args)
        "get-role" == args[0] -> runGetRole(args)
        "sync" == args[0] -> runSync(args)
        else -> help(message)
    }
}

fun runSync(args: Array<String>, suppressOutput: Boolean = false): String {
  if (args.size < 3) {
    help("Usage: zu sync <zk-url> <path>")
    }
    var (_, zkUrl, path) = args
    return try {
      val zk = newZookeeperAdminClient(zkUrl)
      zk.sync(path, null, null)
      val dataArr = zk.getData(path, null, null)
      val clusterSize = String(dataArr).substringAfter("=").trim()
      if (! suppressOutput) {
          print(clusterSize)
      }
      clusterSize
    } catch (e: Exception) {
      System.err.println("Error performing zookeeper sync operation:")
      e.printStackTrace(System.err)
      System.exit(1)
      ""
    }
}

fun runGetAll(args: Array<String>, suppressOutput: Boolean = false): String {
    if (args.size != 2) {
        help("Usage: zu get-all <zk-url>")
    }

    val (_, zkUrl) = args
    return try {
        val zk = newZookeeperAdminClient(zkUrl)
        val zkCfg = String(zk.getConfig(false, Stat()))
        if (! suppressOutput) {
            print(zkCfg)
        }
        zkCfg
    } catch (e: Exception) {
        System.err.println("Error getting server config")
        e.printStackTrace(System.err)
        System.exit(1)
        ""
    }
}

fun runGet(args: Array<String>, suppressOutput: Boolean = false): String? {
    if (args.size != 3) {
        help("Usage: zu get <zk-url> <server-id>")
    }

    val (_, zkUrl, serverId) = args
    val that = runGetAll(arrayOf("get-all", zkUrl), true)
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
}

fun runGetRole(args: Array<String>) {
    if (args.size != 3) {
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
    if (args.size < 4 || args.size > 5) {
        help(message)
    }
    val (_, zkUrl, newServerId, newServerDetails) = args
    reconfigure(zkUrl, "server.$newServerId=$newServerDetails", null, args.getOrNull(4))
}

fun runRemove(args: Array<String>) {
    val message = "Usage: zu remove <zk-url> <server-id> [output-file]"
    if (args.size < 3 || args.size > 4) {
        help(message)
    }
    val (_, zkUrl, newServerId) = args
    reconfigure(zkUrl, null, newServerId, args.getOrNull(2))
}

fun help(message: String) {
    System.err.println(message)
    System.exit(1)
}

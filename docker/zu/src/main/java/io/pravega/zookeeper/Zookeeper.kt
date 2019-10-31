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

import org.apache.zookeeper.WatchedEvent
import org.apache.zookeeper.Watcher
import org.apache.zookeeper.ZooKeeper
import org.apache.zookeeper.admin.ZooKeeperAdmin
import java.util.concurrent.CompletableFuture
import java.util.concurrent.TimeUnit

const val ZK_CONNECTION_TIMEOUT_MINS: Long  = 3

/**
 * Creates a new Zookeeper Admin client and waits until it's in a connected state
 */
fun newZookeeperAdminClient(zkUrl: String) : ZooKeeperAdmin {
    System.err.println("Connecting to Zookeeper $zkUrl")

    val connectionWatcher = ConnectionWatcher()
    val zk = ZooKeeperAdmin(zkUrl, 3000, connectionWatcher)

    connectionWatcher.waitUntilConnected()

    return zk
}

class ConnectionWatcher : Watcher {
    private val connected  = CompletableFuture<Boolean>()

    override fun process(event: WatchedEvent?) {
        if (event?.type == Watcher.Event.EventType.None) {
            if (event.state == Watcher.Event.KeeperState.SyncConnected) {
                connected.complete(true)
            }
        }
    }

    fun waitUntilConnected() {
        connected.get(ZK_CONNECTION_TIMEOUT_MINS, TimeUnit.MINUTES)
    }
}

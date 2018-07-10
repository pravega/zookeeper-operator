package io.pravega.zookeeper

import org.apache.zookeeper.WatchedEvent
import org.apache.zookeeper.Watcher
import org.apache.zookeeper.ZooKeeper
import org.apache.zookeeper.admin.ZooKeeperAdmin
import java.util.concurrent.CompletableFuture
import java.util.concurrent.TimeUnit

const val ZK_CONNECTION_TIMEOUT_MINS: Long  = 10

/**
 * Creates a new Zookeeper client and waits until it's in a connected state
 */
fun newZookeeperClient(zkUrl: String) : ZooKeeper {
    println("Connecting to Zookeeper $zkUrl")

    val connectionWatcher = ConnectionWatcher()
    val zk = ZooKeeper(zkUrl, 3000, connectionWatcher)

    connectionWatcher.waitUntilConnected()

    return zk
}

/**
 * Creates a new Zookeeper Admin client and waits until it's in a connected state
 */
fun newZookeeperAdminClient(zkUrl: String) : ZooKeeperAdmin {
    println("Connecting to Zookeeper $zkUrl")

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


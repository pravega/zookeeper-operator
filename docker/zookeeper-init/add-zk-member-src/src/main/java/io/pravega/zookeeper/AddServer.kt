package io.pravega.zookeeper

/**
 * Utility to Register a server with the Zookeeper Ensemble
 */
fun main(args: Array<String>) {
    if (args.size < 3) {
        println("Usage: add-zk-member <zkUrl> <newServerId> <newServerDetails>")
        System.exit(1)
    }

    val zkUrl = args[0]
    val newServerId = args[1]
    val newServerDetails = args[2]

    try {
        val zk = newZookeeperAdminClient(zkUrl)
        val newConfig = zk.reconfigure("server.$newServerId=$newServerDetails", null, null, -1, null)

        println(String(newConfig))
    }
    catch(e : Exception) {
        println("Error Adding Server")
        e.printStackTrace()
        System.exit(1)
    }
}
#Kubernetes ZooKeeper Scripts

## Starting a ZooKeeper Server
`start-zookeeper` configures and starts a ZooKeeper server. Its parameters are described below.


    --servers           The number of servers in the ensemble. The default 
                        value is 1. This must be set to 
                        `StatefulSet.Spec.Replicas`.

    --data_dir          The directory where the ZooKeeper process will store its
                        snapshots. The default is /var/lib/zookeeper/data. This 
                        directory must be backed by a persistent volume.

    --data_log_dir      The directory where the ZooKeeper process will store its 
                        write ahead log. The default is 
                        /var/lib/zookeeper/data/log. This directory must be 
                        backed by a persistent volume.

    --conf_dir          The directoyr where the ZooKeeper process will store its
                        configuration. The default is /opt/zookeeper/conf.

    --client_port       The port on which the ZooKeeper process will listen for 
                        client requests. The default is 2181. This port must be 
                        specified in both containerPorts and in the Client 
                        Service.

    --election_port     The port on which the ZooKeeper process will perform 
                        leader election. The default is 3888. This port must be 
                        specified in both containerPorts and in the Headless 
                        Service.

    --server_port       The port on which the ZooKeeper process will listen for 
                        requests from other servers in the ensemble. The 
                        default is 2888. This port must be specified in both 
                        containerPorts and in the Client Service.

    --tick_time         The length of a ZooKeeper tick in ms. The default is 
                        2000.

    --init_limit        The number of Ticks that an ensemble member is allowed 
                        to perform leader election. The default is 10.

    --sync_limit        The maximum session timeout that the ensemble will 
                        allows a client to request. The default is 5.

    --heap              The maximum amount of heap to use. The format is the 
                        same as that used for the Xmx and Xms parameters to the 
                        JVM. e.g. --heap=2G. The default is 2G. You must be sure
                        to provide sufficient overhead in the memory resources 
                        requested by ZooKeeper.

    --max_client_cnxns  The maximum number of client connections that the 
                        ZooKeeper process will accept simultaneously. The 
                        default is 60.

    --snap_retain_count The maximum number of snapshots the ZooKeeper process 
                        will retain if purge_interval is greater than 0. The 
                        default is 3.

    --purge_interval    The number of hours the ZooKeeper process will wait 
                        between purging its old snapshots. If set to 0 old 
                        snapshots will never be purged. The default is 0.

    --max_session_timeout The maximum time in milliseconds for a client session 
                        timeout. The default value is 2 * tick time.

    --min_session_timeout The minimum time in milliseconds for a client session 
                        timeout. The default value is 20 * tick time.

    --log_level         The log level for the zookeeeper server. Either FATAL,
                        ERROR, WARN, INFO, DEBUG. The default is INFO.
                        
## Readiness and Liveness Checks
`zookeeper-ready` performs a health check using the `ruok` four letter word. It takes the client port as a parameter. 
If the servers is healthy it exits normally. If the server is unhealthy it exits abnormally causing the readiness 
or liveness check to fail.

## Metrics 
`zookeeper-metrics` uses the `mntr` four letter word to print metrics to stdout. You can use this to integrate 
zookeeper metrics with an existing collector.
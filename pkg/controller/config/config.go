/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package config

// DisableFinalizer disables the finalizers for zookeeper clusters and
// skips the pvc deletion phase when zookeeper cluster get deleted.
// This is useful when operator deletion may happen before zookeeper clusters deletion.
// NOTE: enabling this flag with caution! It causes pvc of zk undeleted.
var DisableFinalizer bool

/*
 * Copyright 1999-2020 Alibaba Group Holding Ltd.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package com.alibaba.nacos.naming.core.v2.cleaner;

import com.alibaba.nacos.common.notify.NotifyCenter;
import com.alibaba.nacos.naming.core.v2.ServiceManager;
import com.alibaba.nacos.naming.core.v2.event.service.ServiceEvent;
import com.alibaba.nacos.naming.core.v2.index.ClientServiceIndexesManager;
import com.alibaba.nacos.naming.core.v2.pojo.Service;
import com.alibaba.nacos.naming.misc.GlobalExecutor;
import com.alibaba.nacos.naming.misc.Loggers;

import java.util.Collection;
import java.util.Set;
import java.util.concurrent.TimeUnit;
import java.util.stream.Stream;

/**
 * Empty service auto cleaner for v2.x.
 *
 * @author xiweng.yy
 */
public class EmptyServiceAutoCleanerV2 extends AbstractNamingCleaner {
    
    private static final String EMPTY_SERVICE = "emptyService";
    
    @Override
    public String getType() {
        return EMPTY_SERVICE;
    }
    
    @Override
    public void doClean() {
            clientServiceIndexesManager.removePublisherIndexesByEmptyService(service);
            ServiceManager.getInstance().removeSingleton(service);
            NotifyCenter.publishEvent(new ServiceEvent.ServiceRemovedEvent(service));
        }
    }
    
    private boolean isTimeExpired(Service service) {
        long currentTimeMillis = System.currentTimeMillis();
        // TODO get expired time from config
        return currentTimeMillis - service.getLastUpdatedTime() >= TimeUnit.MINUTES.toMillis(1);
    }
}
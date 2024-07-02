package com.alibaba.nacos.naming.healthcheck.extend;
import com.alibaba.nacos.api.naming.pojo.AbstractHealthChecker;
/*
 * Copyright 1999-2018 Alibaba Group Holding Ltd.
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
import com.alibaba.nacos.naming.healthcheck.HealthCheckProcessor;
import com.alibaba.nacos.naming.healthcheck.HealthCheckType;
import org.springframework.beans.BeansException;
import org.springframework.beans.factory.BeanFactory;
import org.springframework.beans.factory.BeanFactoryAware;
import org.springframework.beans.factory.config.SingletonBeanRegistry;
import org.springframework.stereotype.Component;

import java.util.HashSet;
import java.util.Iterator;
import java.util.ServiceLoader;
import java.util.Set;

/**
 * @author XCXCXCXCX
 */
@Component
public class HealthCheckExtendProvider implements BeanFactoryAware{

    private ServiceLoader<HealthCheckProcessor> processorLoader
        = ServiceLoader.load(HealthCheckProcessor.class);

    private ServiceLoader<AbstractHealthChecker> checkerLoader
        = ServiceLoader.load(AbstractHealthChecker.class);

    private SingletonBeanRegistry registry;

    private static final char LOWER_A = 'A';
    private static final char LOWER_Z = 'Z';

    public void init(){
        loadExtend();
    }

    private void loadExtend() {
        Iterator<HealthCheckProcessor> processorIt = processorLoader.iterator();
        Iterator<AbstractHealthChecker> healthCheckerIt = checkerLoader.iterator();

        Set<String> processorType = new HashSet<>();
        Set<String> healthCheckerType = new HashSet<>();
        while(processorIt.hasNext()){
            HealthCheckProcessor processor = processorIt.next();
            processorType.add(processor.getType());
            registry.registerSingleton(lowerFirstChar(processor.getClass().getSimpleName()), processor);
        }

        while(healthCheckerIt.hasNext()){
            AbstractHealthChecker checker = healthCheckerIt.next();
            healthCheckerType.add(checker.getType());
            HealthCheckType.registerHealthChecker(checker.getType(), checker.getClass());
        }
        if(!processorType.equals(healthCheckerType)){
            throw new RuntimeException("An unmatched processor and healthChecker are detected in the extension package.");
        }
    }

    private String lowerFirstChar(String simpleName) {
        if(simpleName == null || "".equals(simpleName)){
            throw new IllegalArgumentException("can't find extend processor class name");
        }
        char[] chars = simpleName.toCharArray();
        if(chars[0] >= LOWER_A && chars[0] <= LOWER_Z){
            chars[0] = (char)(chars[0] + 32);
        }
        return String.valueOf(chars);
    }

    @Override
    public void setBeanFactory(BeanFactory beanFactory) throws BeansException {
        if(beanFactory instanceof SingletonBeanRegistry){
            this.registry = (SingletonBeanRegistry) beanFactory;
        }
    }
}
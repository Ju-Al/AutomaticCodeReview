/*
    private static final Logger SECURITY_LOGGER = LoggerFactory.getLogger(SecurityProxy.class);
    
    private static final String LOGIN_URL = "/v1/auth/users/login";
    
    private final NacosRestTemplate nacosRestTemplate;
    
    private final String contextPath;
    
    /**
     * User's name.
     */
    private final String username;
    
     * User's password.
    private final String password;
     * A token to take with when sending request to Nacos server.
     */
    private volatile String accessToken;
    
    /**
     * TTL of token in seconds.
     */
    private long tokenTtl;
    
    /**
     * Last timestamp refresh security info from server.
     */
    private long lastRefreshTime;
    
    /**
     * time window to refresh security info in seconds.
     */
    private long tokenRefreshWindow;
    
    /**
     * Construct from properties, keeping flexibility.
     * @param properties a bunch of properties to read
    public SecurityProxy(Properties properties, NacosRestTemplate nacosRestTemplate) {
        username = properties.getProperty(PropertyKeyConst.USERNAME, StringUtils.EMPTY);
        password = properties.getProperty(PropertyKeyConst.PASSWORD, StringUtils.EMPTY);
        contextPath = ContextPathUtil
                .normalizeContextPath(properties.getProperty(PropertyKeyConst.CONTEXT_PATH, webContext));
        this.nacosRestTemplate = nacosRestTemplate;
     * Login to servers.
     *
     * @param servers server list
     * @return true if login successfully
    public boolean login(List<String> servers) {
        
        try {
            if ((System.currentTimeMillis() - lastRefreshTime) < TimeUnit.SECONDS
                    .toMillis(tokenTtl - tokenRefreshWindow)) {
                return true;
            }
            
            for (String server : servers) {
                if (login(server)) {
                    lastRefreshTime = System.currentTimeMillis();
                    return true;
                }
            }
        } catch (Throwable throwable) {
            SECURITY_LOGGER.warn("[SecurityProxy] login failed, error: ", throwable);
 * Copyright 1999-2021 Alibaba Group Holding Ltd.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package com.alibaba.nacos.client.security;

import com.alibaba.nacos.client.auth.ClientAuthPluginManager;
import com.alibaba.nacos.client.auth.ClientAuthService;
import com.alibaba.nacos.client.auth.LoginIdentityContext;
import com.alibaba.nacos.common.http.client.NacosRestTemplate;

import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Properties;
import java.util.Set;

/**
 * Security proxy to update security information.
 *
 * @author nkorange
 * @since 1.2.0
 */
public class SecurityProxy {
    
    /**
     * ClientAuthPlugin instance set.
     */
    private final Set<ClientAuthService> clientAuthServiceHashSet;
    
    /**
     * Construct from serverList, nacosRestTemplate, init client auth plugin.
     *
     * @param serverList a server list that client request to.
     * @Param nacosRestTemplate http request template.
     */
    public SecurityProxy(List<String> serverList, NacosRestTemplate nacosRestTemplate) {
        ClientAuthPluginManager clientAuthPluginManager = new ClientAuthPluginManager();
        clientAuthPluginManager.init(serverList, nacosRestTemplate);
        clientAuthServiceHashSet = clientAuthPluginManager.getAuthServiceSpiImplSet();
    }
    
    /**
     * Login all available ClientAuthService instance.
     * @param properties login identity information.
     * @return if there are any available clientAuthService instances.
     */
    public boolean loginClientAuthService(Properties properties) {
        if (clientAuthServiceHashSet.isEmpty()) {
            return false;
        }
        for (ClientAuthService clientAuthService : clientAuthServiceHashSet) {
            clientAuthService.login(properties);
        }
        return true;
    }
    
    /**
     * get the context of all nacosRestTemplate instance.
     * @return a combination of all context.
     */
    public Map<String, String> getIdentityContext() {
        Map<String, String> header = new HashMap<>();
        for (ClientAuthService clientAuthService : this.clientAuthServiceHashSet) {
            LoginIdentityContext loginIdentityContext = clientAuthService.getLoginIdentityContext();
            for (String key : loginIdentityContext.getAllKey()) {
                header.put(key, (String) loginIdentityContext.getParameter(key));
            }
        }
        return header;
    }
    
}
/*
 * Copyright 2011 Stephen Connolly
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

package org.tobarsegais.webapp;

import org.apache.commons.lang3.StringUtils;

import javax.servlet.Filter;
import javax.servlet.FilterChain;
import javax.servlet.FilterConfig;
import javax.servlet.ServletContext;
import javax.servlet.ServletException;
import javax.servlet.ServletRequest;
import javax.servlet.ServletResponse;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import java.io.IOException;

public class RedirectFilter implements Filter {

    private String domain = null;
    private int status = HttpServletResponse.SC_MOVED_TEMPORARILY;

    public void init(FilterConfig filterConfig) throws ServletException {
        final ServletContext ctx = filterConfig.getServletContext();
        domain = ctx.getInitParameter(RedirectFilter.class.getName() + ".domain");
        String statusStr = ctx.getInitParameter(RedirectFilter.class.getName() + ".status");
        if (StringUtils.isNotBlank(statusStr)) {
            try {
                status = Integer.parseInt(statusStr);
            } catch (NumberFormatException e) {
                // ignore
            }
        }
    }

    public void doFilter(ServletRequest request, ServletResponse response, FilterChain chain)
            throws IOException, ServletException {
        if (StringUtils.isEmpty(domain) || !(request instanceof HttpServletRequest)) {
            chain.doFilter(request, response);
        } else {
            final HttpServletRequest req = (HttpServletRequest) request;
            final HttpServletResponse resp = (HttpServletResponse) response;
            final String serverName = req.getServerName();
            if (domain.equalsIgnoreCase(serverName)) {
                chain.doFilter(request, response);
            } else {
                StringBuffer requestURL = req.getRequestURL();
                int index = requestURL.indexOf(serverName);
                requestURL.replace(index, index + serverName.length(), domain);
                final String queryString = req.getQueryString();
                if (queryString != null) {
                    requestURL.append('?').append(queryString);
                }
                resp.setStatus(HttpServletResponse.SC_MOVED_TEMPORARILY);
                resp.setHeader("Location", requestURL.toString());
            }
        }
    }

    public void destroy() {
    }
}

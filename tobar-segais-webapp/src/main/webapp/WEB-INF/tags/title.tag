<%--
  ~ Copyright 2012 Stephen Connolly
  ~
  ~ Licensed under the Apache License, Version 2.0 (the "License");
  ~ you may not use this file except in compliance with the License.
  ~ You may obtain a copy of the License at
  ~
  ~     http://www.apache.org/licenses/LICENSE-2.0
  ~
  ~ Unless required by applicable law or agreed to in writing, software
  ~ distributed under the License is distributed on an "AS IS" BASIS,
  ~ WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  ~ See the License for the specific language governing permissions and
  ~ limitations under the License.
  --%>

<%@ tag trimDirectiveWhitespaces="true" %>
<%@ tag import="java.util.Map" %>
<%@ tag import="org.tobarsegais.webapp.data.Toc" %>
<%@ tag import="org.tobarsegais.webapp.ServletContextListenerImpl" %>
<%@ tag import="org.tobarsegais.webapp.data.TocEntry" %>
<title><%
            String path = (String) request.getAttribute("content");
            String pageTitle = null;

            Map<String, Toc> contents = ServletContextListenerImpl.getTablesOfContents(application);
            for (Map.Entry<String,Toc> entry: contents.entrySet()) {
                if (path.startsWith("/"+entry.getKey()+"/")) {
                    final TocEntry topic = entry.getValue().lookupTopic(path.substring(entry.getKey().length() + 2));
                    if (topic != null) {
                        pageTitle = topic.getLabel();
                    } else if (path.equals("/" + entry.getKey() + "/index.html")) {
                        pageTitle = entry.getValue().getLabel();
                    }
                }
            }
            if (pageTitle == null) {
                pageTitle = application.getInitParameter("default.page.title");
            }
            if (pageTitle == null) {
                pageTitle = "Help";
            }

        %><%=pageTitle%></title>

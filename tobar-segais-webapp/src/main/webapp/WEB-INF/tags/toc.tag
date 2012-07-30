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

<%@ tag import="java.util.Map" %>
<%@ tag import="org.tobarsegais.webapp.data.Toc" %>
<%@ tag import="org.tobarsegais.webapp.data.TocEntry" %>
<%@ tag import="org.apache.commons.lang3.StringEscapeUtils" %>
<%@ tag import="java.util.Stack" %>
<%@ tag import="java.util.Iterator" %>
<%@ tag import="org.tobarsegais.webapp.ServletContextListenerImpl" %>
<%@attribute name="id" required="true" %>
<ul id="${id}"><%
    final String contextPath = request.getContextPath();
    Map<String, Toc> contents = ServletContextListenerImpl.getTablesOfContents(application);
        for (Map.Entry<String, Toc> bundleEntry : contents.entrySet()) {
            TocEntry entry = bundleEntry.getValue();
            String bundle = bundleEntry.getKey();
            out.print("<li>");
            if (entry.getHref() != null) {
                out.print("<a href=\"");
                out.print(contextPath);
                out.print("/docs/");
                out.print(bundle);
                out.print("/");
                out.print(entry.getHref());
                out.print("\">");
            }
            out.print("<span>");
            out.print(StringEscapeUtils.escapeHtml4(entry.getLabel()));
            out.print("</span>");
            if (entry.getHref() != null) {
                out.print("</a>");
            }
            Stack<Iterator<? extends TocEntry>> stack = new Stack<Iterator<? extends TocEntry>>();
            if (!entry.getChildren().isEmpty()) {
                out.print("<ul>");
                stack.push(entry.getChildren().iterator());
                while (!stack.empty()) {
                    Iterator<? extends TocEntry> cur = stack.pop();
                    if (cur.hasNext()) {
                        entry = cur.next();
                        stack.push(cur);
                        out.print("<li>");
                        if (entry.getHref() != null) {
                            out.print("<a href=\"");
                            out.print(contextPath);
                            out.print("/docs/");
                            out.print(bundle);
                            out.print("/");
                            out.print(entry.getHref());
                            out.print("\">");
                        }
                        out.print("<span>");
                        out.print(StringEscapeUtils.escapeHtml4(entry.getLabel()));
                        out.print("</span>");
                        if (entry.getHref() != null) {
                            out.print("</a>");
                        }
                        if (!entry.getChildren().isEmpty()) {
                            out.print("<ul>");
                            stack.push(entry.getChildren().iterator());
                        }
                    } else {
                        out.print("</ul></li>");
                    }
                }
            }
            out.print("</li>");
        }
    %>
</ul>
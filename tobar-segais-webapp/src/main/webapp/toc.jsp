<%@ page import="org.tobarsegais.webapp.data.Entry" %>
<%@ page import="org.tobarsegais.webapp.data.Toc" %>
<%@ page import="java.util.Iterator" %>
<%@ page import="java.util.Map" %>
<%@ page import="java.util.Stack" %>
<%--
  ~ Copyright 2011 Stephen Connolly
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

<%

    Map<String, Toc> contents = (Map<String, Toc>) application.getAttribute("toc");

%>
<!doctype html>
<html lang="en">
<head>
    <script src="js/jquery-latest.js"></script>
    <link rel="stylesheet" href="css/screen.css" type="text/css"/>
    <link rel="stylesheet" href="css/jquery.treeview.css" type="text/css"/>
    <script type="text/javascript" src="js/jquery.treeview.js"></script>
    <script>
        $(document).ready(function() {
            $("#toc").treeview({collapsed:true});
        });
    </script>

</head>
<body>
<ul id="toc" class="filetree">
    <%
        for (Map.Entry<String, Toc> bundleEntry : contents.entrySet()) {
            Entry entry = bundleEntry.getValue();
            String bundle = bundleEntry.getKey();
    %>
    <li>
        <%if (entry.getHref() != null) {%>
        <a target="content" href="/content/<%=bundle+"/"+entry.getHref()%>">
            <%}%>
        <span
                class="<%=entry.getChildren().isEmpty()?"file":"folder"%>"><%=entry.getLabel()%></span>
            <%if (entry.getHref() != null) {%>
        </a>
        <%}%>
        <%
            Stack<Iterator<? extends Entry>> stack = new Stack<Iterator<? extends Entry>>();
            if (!entry.getChildren().isEmpty()) {
                out.print("<ul>");
                stack.push(entry.getChildren().iterator());
                while (!stack.empty()) {
                    Iterator<? extends Entry> cur = stack.pop();
                    if (cur.hasNext()) {
                        entry = cur.next();
                        stack.push(cur);
                        out.print("<li>");
        %>
        <%if (entry.getHref() != null) {%>
        <a target="content" href="/content/<%=bundle+"/"+entry.getHref()%>">
            <%}%>
        <span
                class="<%=entry.getChildren().isEmpty()?"file":"folder"%>"><%=entry.getLabel()%></span>
            <%if (entry.getHref() != null) {%>
        </a>
        <%}%>
        <%
                        if (!entry.getChildren().isEmpty()) {
                            out.print("<ul>");
                            stack.push(entry.getChildren().iterator());
                        }
                    } else {
                        out.print("</ul></li>");
                    }
                }
            }
        %></li>
    <%
        }
    %>
</ul>
</body>
</html>
